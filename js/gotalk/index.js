"use strict";
var protocol = require('./protocol'),
      txt = protocol.text,
      bin = protocol.binary;
var Buf = require('./buf');
var utf8 = require('./utf8');
var EventEmitter = require('./EventEmitter');
var keepalive = require('./keepalive');

var gotalk = exports;

gotalk.protocol = protocol;
gotalk.Buf = Buf;

function decodeJSON(v) {
  var value;
  try {
    value = JSON.parse(v);
  } catch (e) {
    // console.warn('failed to decode JSON "'+(typeof v === 'string' ? v : v.toString())+'":',e);
  }
  return value;
}


// ===============================================================================================

function Sock(handlers) { return Object.create(Sock.prototype, {
  // Public properties
  handlers:      {value:handlers, enumerable:true},
  protocol:      {value: Buf ? protocol.binary : protocol.text, enumerable:true, writable:true},
  heartbeatInterval: {value: 20 * 1000, enumerable:true, writable:true},

  // Internal
  ws:            {value:null, writable:true},
  keepalive:     {value:null, writable:true},

  // Used for performing requests
  nextOpID:      {value:0, writable:true},
  nextStreamID:  {value:0, writable:true},
  pendingRes:    {value:{}, writable:true},
  hasPendingRes: {get:function(){ for (var k in this.pendingRes) { return true; } }},

  // True if end() has been called while there were outstanding responses
  pendingClose:  {value:false, writable:true}
}); }

Sock.prototype = EventEmitter.mixin(Sock.prototype);
exports.Sock = Sock;


var resetSock = function(s, causedByErr) {
  s.pendingClose = false;

  if (s.ws) {
    s.ws.onmessage = null;
    s.ws.onerror = null;
    s.ws.onclose = null;
    s.ws = null;
  }

  s.nextOpID = 0;
  if (s.hasPendingRes) {
    var err = causedByErr || new Error('connection closed');
    // TODO: return a RetryResult kind of error instead of just an error
    for (var k in s.pendingRes) {
      s.pendingRes[k](err);
    }
    s.pendingRes = {};
  }
};


var websocketCloseStatus = {
  1000: 'normal',
  1001: 'going away',
  1002: 'protocol error',
  1003: 'unsupported',
  // 1004 is currently unassigned
  1005: 'no status',
  1006: 'abnormal',
  1007: 'inconsistent',
  1008: 'invalid message',
  1009: 'too large',
};


// Adopt a web socket, which should be in an OPEN state
Sock.prototype.adoptWebSocket = function(ws) {
  var s = this;
  if (ws.readyState !== WebSocket.OPEN) {
    throw new Error('web socket readyState != OPEN');
  }
  ws.binaryType = 'arraybuffer';
  s.ws = ws;
  ws.onclose = function(ev) {
    var err = ws._gotalkCloseError;
    if (!err && ev.code !== 1000) {
      err = new Error('websocket closed: ' + (websocketCloseStatus[ev.code] || '#'+ev.code));
    }
    resetSock(s, err);
    s.emit('close', err);
  };
  ws.onmessage = function(ev) {
    if (!ws._bufferedMessages) ws._bufferedMessages = [];
    ws._bufferedMessages.push(ev.data);
  };
};


Sock.prototype.handshake = function () {
  this.ws.send(this.protocol.versionBuf);
};


Sock.prototype.end = function() {
  // Allow calling twice to "force close" even when there are pending responses
  var s = this;
  if (s.keepalive) {
    s.keepalive.disable();
    s.keepalive = null;
  }
  if (!s.pendingClose && s.hasPendingRes) {
    s.pendingClose = true;
  } else if (s.ws) {
    s.ws.close(1000);
  }
};


Sock.prototype.address = function() {
  var s = this;
  if (s.ws) {
    return s.ws.url;
  }
  return null;
};

// ===============================================================================================
// Reading messages from a connection

var ErrAbnormal = exports.ErrAbnormal = Error("unsupported protocol");
ErrAbnormal.isGotalkProtocolError = true;
ErrAbnormal.code = protocol.ErrorAbnormal;

var ErrUnsupported = exports.ErrUnsupported = Error("unsupported protocol");
ErrUnsupported.isGotalkProtocolError = true;
ErrUnsupported.code = protocol.ErrorUnsupported;

var ErrInvalidMsg = exports.ErrInvalidMsg = Error("invalid protocol message");
ErrInvalidMsg.isGotalkProtocolError = true;
ErrInvalidMsg.code = protocol.ErrorInvalidMsg;

var ErrTimeout = exports.ErrTimeout = Error("timeout");
ErrTimeout.isGotalkProtocolError = true;
ErrTimeout.code = protocol.ErrorTimeout;


Sock.prototype.sendHeartbeat = function (load) {
  var s = this, buf = s.protocol.makeHeartbeatMsg(Math.round(load * protocol.HeartbeatMsgMaxLoad));
  try {
    s.ws.send(buf);
  } catch (err) {
    if (!this.ws || this.ws.readyState > WebSocket.OPEN) {
      err = new Error('socket is closed');
    }
    throw err;
  }
};


Sock.prototype.startSendingHeartbeats = function() {
  var s = this;
  if (s.heartbeatInterval < 10) {
    throw new Error("Sock.heartbeatInterval is too low");
  }
  clearTimeout(s._sendHeartbeatsTimer);
  var send = function() {
    clearTimeout(s._sendHeartbeatsTimer);
    s.sendHeartbeat(0);
    s._sendHeartbeatsTimer = setTimeout(send, s.heartbeatInterval);
  };
  s._sendHeartbeatsTimer = setTimeout(send, 1);
};


Sock.prototype.stopSendingHeartbeats = function() {
  clearTimeout(s._sendHeartbeatsTimer);
};


Sock.prototype.startReading = function () {
  var s = this, ws = s.ws, msg;  // msg = current message

  function readMsg(ev) {
    msg = typeof ev.data === 'string' ? txt.parseMsg(ev.data) : bin.parseMsg(Buf(ev.data));
    // console.log(
    //   'readMsg:',
    //   typeof ev.data === 'string' ? ev.data : Buf(ev.data).toString(),
    //   'msg:',
    //   msg
    // );
    if (msg.t === protocol.MsgTypeProtocolError) {
      var errcode = msg.size;
      if (errcode === protocol.ErrorAbnormal) {
        ws._gotalkCloseError = ErrAbnormal;
      } else if (errcode === protocol.ErrorUnsupported) {
        ws._gotalkCloseError = ErrUnsupported;
      } else if (errcode === protocol.ErrorTimeout) {
        ws._gotalkCloseError = ErrTimeout;
      } else {
        ws._gotalkCloseError = ErrInvalidMsg;
      }
      ws.close(4000 + errcode);
    } else if (msg.size !== 0 && msg.t !== protocol.MsgTypeHeartbeat) {
      ws.onmessage = readMsgPayload;
    } else {
      s.handleMsg(msg);
      msg = null;
    }
  }

  function readMsgPayload(ev) {
    var b = ev.data;
    ws.onmessage = readMsg;
    s.handleMsg(msg, typeof b === 'string' ? b : Buf(b));
    msg = null;
  }

  function readVersion(ev) {
    var peerVersion = typeof ev.data === 'string' ? txt.parseVersion(ev.data) :
                                                    bin.parseVersion(Buf(ev.data));
    if (peerVersion !== protocol.Version) {
      ws._gotalkCloseError = ErrUnsupported;
      s.closeError(protocol.ErrorUnsupported);
    } else {
      ws.onmessage = readMsg;
      if (s.heartbeatInterval > 0) {
        s.startSendingHeartbeats();
      }
    }
  }

  // We begin by sending our version and reading the remote side's version
  ws.onmessage = readVersion;

  // Any buffered messages?
  if (ws._bufferedMessages) {
    ws._bufferedMessages.forEach(function(data){ ws.onmessage({data:data}); });
    ws._bufferedMessages = null;
  }
};

// ===============================================================================================
// Handling of incoming messages

var msgHandlers = {};

Sock.prototype.handleMsg = function(msg, payload) {
  // console.log('handleMsg:', String.fromCharCode(msg.t), msg, 'payload:', payload);
  var msgHandler = msgHandlers[msg.t];
  if (!msgHandler) {
    if (s.ws) {
      s.ws._gotalkCloseError = ErrInvalidMsg;
    }
    s.closeError(protocol.ErrorInvalidMsg);
  } else {
    msgHandler.call(this, msg, payload);
  }
};

msgHandlers[protocol.MsgTypeSingleReq] = function (msg, payload) {
  var s = this, handler, result;
  handler = s.handlers.findRequestHandler(msg.name);

  result = function (outbuf) {
    s.sendMsg(protocol.MsgTypeSingleRes, msg.id, null, 0, outbuf);
  };
  result.error = function (err) {
    var errstr = err.message || String(err);
    s.sendMsg(protocol.MsgTypeErrorRes, msg.id, null, 0, errstr);
  };

  if (typeof handler !== 'function') {
    result.error('no such operation "'+msg.name+'"');
  } else {
    try {
      handler(payload, result, msg.name);
    } catch (err) {
      if (typeof console !== 'undefined') { console.error(err.stack || err); }
      result.error('internal error');
    }
  }
};

function handleRes(msg, payload) {
  var id = typeof msg.id === 'string' ? msg.id : msg.id.toString();
  var s = this, callback = s.pendingRes[id];
  if (msg.t !== protocol.MsgTypeStreamRes || !payload || (payload.length || payload.size) === 0) {
    delete s.pendingRes[id];
    if (s.pendingClose && !s.hasPendingRes) {
      s.end();
    }
  }
  if (typeof callback !== 'function') {
    return; // ignore message
  }
  if (msg.t === protocol.MsgTypeErrorRes) {
    callback(new Error(String(payload)), null);
  } else {
    callback(null, payload);
  }
}

msgHandlers[protocol.MsgTypeSingleRes] = handleRes;
msgHandlers[protocol.MsgTypeStreamRes] = handleRes;
msgHandlers[protocol.MsgTypeErrorRes] = handleRes;

msgHandlers[protocol.MsgTypeNotification] = function (msg, payload) {
  var s = this, handler = s.handlers.findNotificationHandler(msg.name);
  if (handler) {
    handler(payload, msg.name);
  }
};

msgHandlers[protocol.MsgTypeHeartbeat] = function (msg) {
  this.emit('heartbeat', {time:new Date(msg.size * 1000), load:msg.wait});
};

// ===============================================================================================
// Sending messages


Sock.prototype.sendMsg = function(t, id, name, wait, payload) {
  var payloadSize = (payload && typeof payload === 'string' && this.protocol === protocol.binary) ?
    utf8.sizeOf(payload) :
    payload ? payload.length || payload.size :
    0;
  var s = this, buf = s.protocol.makeMsg(t, id, name, wait, payloadSize);
  // console.log('sendMsg(',t,id,name,payload,'): protocol.makeMsg =>',
  //   typeof buf === 'string' ? buf : buf.toString());
  try {
    s.ws.send(buf);
    if (payloadSize !== 0) {
      s.ws.send(payload);
    }
  } catch (err) {
    if (!this.ws || this.ws.readyState > WebSocket.OPEN) {
      err = new Error('socket is closed');
    }
    throw err;
  }
};


Sock.prototype.closeError = function(code) {
  var s = this, buf;
  if (s.ws) {
    try {
      s.ws.send(s.protocol.makeMsg(protocol.MsgTypeProtocolError, null, null, 0, code));
    } catch (e) {}
    s.ws.close(4000 + code);
  }
};

var zeroes = '0000';

// callback function(Error, outbuf)
Sock.prototype.bufferRequest = function(op, buf, callback) {
  var s = this, id = s.nextOpID++;
  if (s.nextOpID === 1679616) {
    // limit for base36 within 4 digits (36^4=1679616)
    s.nextOpID = 0;
  }
  id = id.toString(36);
  id = zeroes.substr(0, 4 - id.length) + id;

  s.pendingRes[id] = callback;
  try {
    s.sendMsg(protocol.MsgTypeSingleReq, id, op, 0, buf);
  } catch (err) {
    delete s.pendingRes[id];
    callback(err);
  }
}


Sock.prototype.bufferNotify = function(name, buf) {
  s.sendMsg(protocol.MsgTypeNotification, null, name, 0, buf);
}


Sock.prototype.request = function(op, value, callback) {
  var buf;
  if (!callback) {
    // no value
    callback = value;
  } else {
    buf = JSON.stringify(value);
  }
  return this.bufferRequest(op, buf, function (err, buf) {
    var value = decodeJSON(buf);
    return callback(err, value);
  });
};


Sock.prototype.notify = function(op, value) {
  var buf = JSON.stringify(value);
  return this.bufferNotify(op, buf);
};


// ===============================================================================================

// Represents a stream request.
// Response(s) arrive by the "data"(buf) event. When the response is complete, a "end"(error)
// event is emitted, where error is non-empty if the request failed.
var StreamRequest = function(s, op, id) {
  return Object.create(StreamRequest.prototype, {
    s:          {value:s},
    op:         {value:op, enumerable:true},
    id:         {value:id, enumerable:true},
    onresponse: {value:function(){}, enumerable:true, write:true}
  });
};

EventEmitter.mixin(StreamRequest.prototype);

StreamRequest.prototype.write = function (buf) {
  if (!this.ended) {
    if (!this.started) {
      this.started = true;
      this.s.sendMsg(protocol.MsgTypeStreamReq, this.id, this.op, 0, buf);
    } else {
      this.s.sendMsg(protocol.MsgTypeStreamReqPart, this.id, null, 0, buf);
    }
    if (!buf || buf.length === 0 || buf.size === 0) {
      this.ended = true;
    }
  }
};

// Finalize the request
StreamRequest.prototype.end = function () {
  this.write(null);
};

Sock.prototype.streamRequest = function(op) {
  var s = this, id = s.nextStreamID++;
  if (s.nextStreamID === 46656) {
    // limit for base36 within 3 digits (36^3=46656)
    s.nextStreamID = 0;
  }
  id = id.toString(36);
  id = '!' + zeroes.substr(0, 3 - id.length) + id;

  var req = StreamRequest(s, op, id);

  s.pendingRes[id] = function (err, buf) {
    if (err) {
      req.emit('end', err);
    } else if (!buf || buf.length === 0) {
      req.emit('end', null);
    } else {
      req.emit('data', buf);
    }
  };

  return req;
};


// ===============================================================================================

function Handlers() { return Object.create(Handlers.prototype, {
  reqHandlers:         {value:{}},
  reqFallbackHandler:  {value:null, writable:true},
  noteHandlers:        {value:{}},
  noteFallbackHandler: {value:null, writable:true}
}); }
exports.Handlers = Handlers;


Handlers.prototype.handleBufferRequest = function(op, handler) {
  if (!op) {
    this.reqFallbackHandler = handler;
  } else {
    this.reqHandlers[op] = handler;
  }
};

Handlers.prototype.handleRequest = function(op, handler) {
  return this.handleBufferRequest(op, function (buf, result, op) {
    var resultWrapper = function(value) {
      return result(JSON.stringify(value));
    };
    resultWrapper.error = result.error;
    var value = decodeJSON(buf);
    handler(value, resultWrapper, op);
  });
};

Handlers.prototype.handleBufferNotification = function(name, handler) {
  if (!name) {
    this.noteFallbackHandler = handler;
  } else {
    this.noteHandlers[name] = handler;
  }
};

Handlers.prototype.handleNotification = function(name, handler) {
  this.handleBufferNotification(name, function (buf, name) {
    handler(decodeJSON(buf), name);
  });
};

Handlers.prototype.findRequestHandler = function(op) {
  var handler = this.reqHandlers[op];
  return handler || this.reqFallbackHandler;
};

Handlers.prototype.findNotificationHandler = function(name) {
  var handler = this.noteHandlers[name];
  return handler || this.noteFallbackHandler;
};

// ===============================================================================================

function openWebSocket(s, addr, callback) {
  var ws;
  try {
    ws = new WebSocket(addr);
    ws.binaryType = 'arraybuffer';
    ws.onclose = function (ev) {
      var err = new Error('connection failed');
      if (callback) callback(err);
    };
    ws.onopen = function(ev) {
      ws.onerror = undefined;
      s.adoptWebSocket(ws);
      s.handshake();
      if (callback) callback(null, s);
      s.emit('open');
      s.startReading();
    };
    ws.onmessage = function(ev) {
      if (!ws._bufferedMessages) ws._bufferedMessages = [];
      ws._bufferedMessages.push(ev.data);
    };
  } catch (err) {
    if (callback) callback(err);
    s.emit('close', err);
  }
}


// gotalk.defaultResponderAddress is defined if the responder has announced a default address
// to which connect to.
if (window.gotalkResponderAt !== undefined) {
  var at = window.gotalkResponderAt;
  delete window.gotalkResponderAt;
  if (at && at.ws) {
    gotalk.defaultResponderAddress = 'ws://' + document.location.host + at.ws;
  }
}


Sock.prototype.open = function(addr, callback) {
  var s = this;
  if (!callback && typeof addr == 'function') {
    callback = addr;
    addr = null;
  }

  if (!addr) {
    if (!gotalk.defaultResponderAddress) {
      throw new Error('address not specified (responder has not announced any default address)')
    }
    addr = gotalk.defaultResponderAddress;
  }

  if (addr.substr(0,3) === 'ws:') {
    openWebSocket(s, addr, callback);
  } else {
    throw new Error('unsupported address');
  }
  return s;
};


// Open a connection to a gotalk responder.
// 
// open(addr string[, onConnect(Error, Sock)]) -> Sock
//   Connect to gotalk responder at `addr`
//
// open([onConnect(Error, Sock)]) -> Sock
//   Connect to default gotalk responder.
//   Throws an error if `gotalk.defaultResponderAddress` isn't defined.
//
gotalk.open = function(addr, onConnect) {
  var s = Sock(gotalk.defaultHandlers);
  s.open(addr, onConnect);
  return s;
};


// If `addr` is not provided, `gotalk.defaultResponderAddress` is used instead.
Sock.prototype.openKeepAlive = function(addr) {
  var s = this;
  if (s.keepalive) {
    s.keepalive.disable();
  }
  s.keepalive = keepalive(s, addr);
  s.keepalive.enable();
  return s;
};


// Returns a new Sock with a persistent connection to a gotalk responder.
// The Connection is automatically kept alive (by reconnecting) until Sock.end() is called.
// If `addr` is not provided, `gotalk.defaultResponderAddress` is used instead.
gotalk.connection = function(addr) {
  var s = Sock(gotalk.defaultHandlers);
  s.openKeepAlive(addr);
  return s;
};


gotalk.defaultHandlers = Handlers();

gotalk.handleBufferRequest = function(op, handler) {
  return gotalk.defaultHandlers.handleBufferRequest(op, handler);
};

gotalk.handle = function(op, handler) {
  return gotalk.defaultHandlers.handleRequest(op, handler);
};

gotalk.handleBufferNotification = function (name, handler) {
  return gotalk.defaultHandlers.handleBufferNotification(name, handler);
};

gotalk.handleNotification = function (name, handler) {
  return gotalk.defaultHandlers.handleNotification(name, handler);
};

// -----------------------------------------------------------------------------------------------


