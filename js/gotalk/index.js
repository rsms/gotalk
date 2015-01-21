"use strict";
var protocol = require('./protocol'),
      txt = protocol.text,
      bin = protocol.binary;
var Buf = require('./buf');
var utf8 = require('./utf8');
var EventEmitter = require('./EventEmitter');
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

  // Internal
  ws:            {value:null, writable:true},

  // Used for performing requests
  nextOpID:      {value:0, writable:true},
  pendingRes:    {value:{}},
  hasPendingRes: {get:function(){ for (var k in this.pendingRes) { return true; } }},

  // True if end() has been called while there were outstanding responses
  pendingClose:  {value:false, writable:true}
}); }

Sock.prototype = EventEmitter.mixin(Sock.prototype);
exports.Sock = Sock;


// Adopt a web socket, which should be in an OPEN state
Sock.prototype.adoptWebSocket = function(ws) {
  var s = this;
  if (ws.readyState !== WebSocket.OPEN) {
    throw new Error('web socket readyState != OPEN');
  }
  ws.binaryType = 'arraybuffer';
  s.ws = ws;
  ws.onclose = function(ev) {
    s.emit('close', ev.code !== 1000 ? new Error('web socket #'+ev.code) : undefined);
    ws.onmessage = null;
    ws.onerror = null;
    ws.onclose = null;
    s.ws = null;
  };
  ws.onerror = function(ev) {
    s.emit('close', new Error('web socket error'));
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
  if (!s.pendingClose && s.hasPendingRes) {
    s.pendingClose = true;
  } else {
    if (s.hasPendingRes) {
      var err = new Error('socket is closing');
      for (var k in pendingRes) {
        pendingRes[k](err);
      }
    }
    if (s.ws) {
      s.ws.close();
    } else if (s.conn) {
      s.conn.end();
    }
    s.pendingClose = false;
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

Sock.prototype.startReading = function () {
  var s = this, ws = s.ws, msg;  // msg = current message

  function readMsg(ev) {
    msg = typeof ev.data === 'string' ? txt.parseMsg(ev.data) : bin.parseMsg(Buf(ev.data));
    // console.log('readMsg:',
    //   typeof ev.data === 'string' ? ev.data : Buf(ev.data).toString(),
    //   'msg:', msg, 'ev:', ev);
    if (msg.size !== 0) {
      ws.onmessage = readMsgPayload;
    } else {
      s.handleMsg(msg);
      msg = null;
    }
  }

  function readMsgPayload(ev) {
    var b = ev.data;
    s.handleMsg(msg, typeof b === 'string' ? b : Buf(b));
    msg = null;
    ws.onmessage = readMsg;
  }

  function readVersion(ev) {
    var peerVersion = typeof ev.data === 'string' ? txt.parseVersion(ev.data) :
                                                    bin.parseVersion(Buf(ev.data));
    if (peerVersion !== protocol.Version) {
      ws.close(3000, 'gotalk protocol version mismatch');
    } else {
      ws.onmessage = readMsg;
    }
  }

  // We begin by sending our version and reading the remote side's version
  ws.onmessage = readVersion;

  // Any buffered messages?
  if (ws._bufferedMessages) {
    console.log("flush buffered messages")
    ws._bufferedMessages.forEach(function(data){ ws.onmessage({data:data}); });
    ws._bufferedMessages = null;
  }
};

// ===============================================================================================
// Handling of incoming messages

var msgHandlers = {};

Sock.prototype.handleMsg = function(msg, payload) {
  // console.log('handleMsg:', String.fromCharCode(msg.t), msg, 'payload:', payload);
  return msgHandlers[msg.t].call(this, msg, payload);
};

msgHandlers[protocol.MsgTypeSingleReq] = function (msg, payload) {
  var s = this, handler, result;
  handler = s.handlers.findRequestHandler(msg.name);

  result = function (outbuf) {
    s.sendMsg(protocol.MsgTypeSingleRes, msg.id, null, outbuf);
  };
  result.error = function (err) {
    var errstr = err.message || String(err);
    s.sendMsg(protocol.MsgTypeErrorRes, msg.id, null, errstr);
  };

  if (typeof handler !== 'function') {
    result.error('no such operation');
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
  delete s.pendingRes[id];
  if (typeof callback !== 'function') {
    return; // ignore message
  }
  if (s.pendingClose && !s.hasPendingRes) {
    s.end();
  }
  if (msg.t === protocol.MsgTypeSingleRes) {
    callback(null, payload);
  } else {
    callback(new Error(String(payload)), null);
  }
}

msgHandlers[protocol.MsgTypeSingleRes] = handleRes;
msgHandlers[protocol.MsgTypeErrorRes] = handleRes;

msgHandlers[protocol.MsgTypeNotification] = function (msg, payload) {
  var s = this, handler = s.handlers.findNotificationHandler(msg.name);
  if (handler) {
    handler(payload, msg.name);
  }
};

// ===============================================================================================
// Sending messages

Sock.prototype.sendMsg = function(t, id, name, payload) {
  if (!this.ws || this.ws.readyState > WebSocket.OPEN) {
    throw new Error('socket is closed');
  }
  var payloadSize = (payload && typeof payload === 'string' && this.protocol === protocol.binary) ?
    utf8.sizeOf(payload) :
    payload ? payload.length :
    0;
  var s = this, buf = s.protocol.makeMsg(t, id, name, payloadSize);
  // console.log('sendMsg(',t,id,name,payload,'): protocol.makeMsg =>',
  //   typeof buf === 'string' ? buf : buf.toString());
  s.ws.send(buf);
  if (payloadSize !== 0) {
    s.ws.send(payload);
  }
};

var zeroes = '000';

// callback function(Error, outbuf)
Sock.prototype.bufferRequest = function(op, buf, callback) {
  var s = this, id = s.nextOpID++;
  if (s.nextOpID === 46656) {
    // limit for base36 within 3 digits (36^2=46656)
    s.nextOpID = 0;
  }
  id = id.toString(36);
  id = zeroes.substr(0,3 - id.length) + id;
  s.pendingRes[id] = callback;
  try {
    s.sendMsg(protocol.MsgTypeSingleReq, id, op, buf);
  } catch (err) {
    delete s.pendingRes[id];
    callback(err);
  }
}


Sock.prototype.bufferNotify = function(name, buf) {
  s.sendMsg(protocol.MsgTypeNotification, null, name, buf);
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

function connectWebSocket(s, addr, callback) {
  var ws;
  try {
    ws = new WebSocket(addr);
    ws.binaryType = 'arraybuffer';
    ws.onclose = function (ev) {
      var err = new Error('connection failed');
      if (callback) callback(err);
      s.emit('close', err);
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


gotalk.connect = function connect(addr, callback) {
  var s = Sock(gotalk.defaultHandlers);
  if (addr.substr(0,5) === 'ws://') {
    connectWebSocket(s, addr, callback);
  } else {
    throw new Error('unsupported address');
  }
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

