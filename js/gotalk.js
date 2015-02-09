(function(global){
"use strict";

var __mod = {}, __api = {}, __main, __mainapi;
var evalDepth = 0;

var require = function(name) {
  var m, id = name.replace(/^.\//, "");
  m = __api[id];
  //console.log('require', {name:name, id:id, exports:m});
  if (!m) {
    var prefix = ''; for (var i = 0; i < evalDepth; i++) {
      prefix += '. ';
    }
    var f = __mod[id];
    if (f && evalDepth < 100) {
      __mod[id] = null;
      __api[id] = {exports:{}};
      ++evalDepth;
      f(__api[id]);
      --evalDepth;
      __api[id] = __api[id].exports;
    }
    m = __api[id];
  }
  return m;
};

__mod["EventEmitter"]=function(module) { var exports = module.exports;

function EventEmitter() {}
module.exports = EventEmitter;

EventEmitter.prototype.addListener = function (type, listener) {
  if (typeof listener !== 'function') throw TypeError('listener must be a function');
  if (!this.__events) {
    Object.defineProperty(this, '__events', {value:{}, enumerable:false, writable:true});
    this.__events[type] = [listener];
    return this;
  }
  var listeners = this.__events[type];
  if (listeners === undefined) {
    this.__events[type] = [listener];
    return this;
  }
  listeners.push(listener);
  return this;
};

EventEmitter.prototype.on = EventEmitter.prototype.addListener;

EventEmitter.prototype.once = function (type, listener) {
  var fired = false;
  var trigger_event_once = function() {
    this.removeListener(type, trigger_event_once);
    if (!fired) {
      fired = true;
      listener.apply(this, arguments);
    }
  }
  return this.on(type, trigger_event_once);
};

EventEmitter.prototype.removeListener = function (type, listener) {
  var p, listeners = this.__events ? this.__events[type] : undefined;
  if (listeners !== undefined) {
    while ((p = listeners.indexOf(listener)) !== -1) {
      listeners.splice(p,1);
    }
    if (listeners.length === 0) {
      delete this.__events[type];
    }
    return listeners.length;
  }
  return this;
};

EventEmitter.prototype.removeAllListeners = function (type) {
  if (this.__events) {
    if (type) {
      delete this.__events[type];
    } else {
      delete this.__events;
    }
  }
};

EventEmitter.prototype.listeners = function (type) {
  return type ? (this.__events ? this.__events[type] : undefined) : this.__events;
};

EventEmitter.prototype.emit = function (type) {
  var listeners = this.__events ? this.__events[type] : undefined;
  if (listeners === undefined) {
    return false;
  }
  var i = 0, L = listeners.length, args = Array.prototype.slice.call(arguments,1);
  for (; i !== L; ++i) {
    if (!listeners[i]) {
      console.log('e', type, i, args);
    }
    listeners[i].apply(this, args);
  }
  return true;
};

EventEmitter.mixin = function mixin(obj) {
  var proto = obj;
  while (proto) {
    if (proto.__proto__ === Object.prototype) {
      proto.__proto__ = EventEmitter.prototype;
      return obj;
    }
    proto = proto.__proto__;
  }
  return obj;
};


};

__mod["buf"]=function(module) { var exports = module.exports;
"use strict";
var Buf;

if (typeof Uint8Array !== 'undefined') {

var utf8 = require('./utf8');

Uint8Array.prototype.toString = function (encoding, start, end) {
  // assumes buffer contains UTF8-encoded text
  return utf8.decode(this, start, end);
};

Uint8Array.prototype.slice = Uint8Array.prototype.subarray;

// Copies data from a region of this buffer to a region in the target buffer.
// copy(targetBuffer, [targetStart], [sourceStart], [sourceEnd]) -> Buf
Uint8Array.prototype.copy = function (targetBuffer, targetStart, sourceStart, sourceEnd) {
  var srcBuf = this;
  if (sourceStart) {
    srcBuf = srcBuf.slice(sourceStart, sourceEnd || srcBuf.length - sourceStart);
  }
  targetBuffer.set(srcBuf, targetStart || 0);
};


// Buf(Buf) -> Buf
// Buf(size int) -> Buf
// Buf(ArrayBuffer) -> Buf
Buf = function Buf(v) {
  return v instanceof Uint8Array ? v :
    new Uint8Array(
      v instanceof ArrayBuffer ? v :
      new ArrayBuffer(v)
    );
};

Buf.isBuf = function (v) {
  return v instanceof Uint8Array;
};

Buf.fromString = function (s, encoding) {
  return utf8.encode(s, Buf);
};

}

module.exports = Buf;

};

__mod["keepalive"]=function(module) { var exports = module.exports;
"use strict";
// Stay connected by automatically reconnecting w/ exponential back-off.

var netAccess = require('./netaccess');
var protocol = require('./protocol');

// `s` must conform to interface { connect(addr string, cb function(Error)) }
// Returns an object {
//   isConnected bool  // true if currently connected
//   isEnabled bool    // true if enabled
//   enable()          // enables staying connected
//   disable()         // disables trying to stay connected
// }
var keepalive = function(s, addr, minReconnectDelay, maxReconnectDelay) {
  if (!minReconnectDelay) {
    minReconnectDelay = 500;
  } else if (minReconnectDelay < 100) {
    minReconnectDelay = 100;
  }

  if (!maxReconnectDelay || maxReconnectDelay < minReconnectDelay) {
    maxReconnectDelay = 5000;
  }

  var ctx, open, retry, delay = 0, openTimer, opentime;

  ctx = {
    isEnabled: false,
    isConnected: false,
    enable: function() {
      if (!ctx.enabled) {
        ctx.enabled = true;
        delay = 0;
        if (!ctx.isConnected) {
          open();
        }
      }
    },
    disable: function() {
      if (ctx.enabled) {
        clearTimeout(openTimer);
        ctx.enabled = false;
        delay = 0;
      }
    }
  };

  open = function() {
    clearTimeout(openTimer);
    s.open(addr, function(err) {
      opentime = new Date;
      if (err) {
        retry(err);
      } else {
        delay = 0;
        ctx.isConnected = true;
        s.once('close', retry);
      }
    });
  };

  retry = function(err) {
    clearTimeout(openTimer);
    ctx.isConnected = false;
    if (!ctx.enabled) {
      return;
    }
    if (netAccess.available && !netAccess.onLine && 
        !(typeof document !== 'undefined' &&
          document.location &&
          document.location.hostname !== 'localhost' &&
          document.location.hostname !== '127.0.0.1' &&
          document.location.hostname !== '[::1]') )
    {
      netAccess.once('online', retry);
      delay = 0;
      return;
    }
    if (err) {
      if (err.isGotalkProtocolError) {
        if (err.code === protocol.ErrorTimeout) {
          delay = 0;
        } else {
          // We shouldn't retry with the same version of our gotalk library.
          // However, the only sensible thing to do in this case is to let the user code react to
          // the error passed to the close event (e.g. to show a "can't talk to server" UI), and
          // retry in maxReconnectDelay.
          // User code can choose to call `disable()` on its keepalive object in this case.
          delay = maxReconnectDelay;
        }
      } else {
        // increase back off in case of an error
        delay = delay ? Math.min(maxReconnectDelay, delay * 2) : minReconnectDelay;
      }
    } else {
      delay = Math.max(0, minReconnectDelay - ((new Date) - opentime));
    }
    openTimer = setTimeout(open, delay);
  };

  return ctx;
};

module.exports = keepalive;

};

__mod["netaccess"]=function(module) { var exports = module.exports;
"use strict";
var EventEmitter = require('./EventEmitter');
var m;

if (typeof global !== 'undefined' && global.addEventListener) {
  m = Object.create(EventEmitter.prototype, {
    available: {value:true, enumerable:true},
    onLine:    {value:true, enumerable:true, writable:true}
  });

  if (typeof navigator !== 'undefined') {
    m.onLine = navigator.onLine;
  }

  global.addEventListener("offline", function (ev) {
    m.onLine = false;
    m.emit('offline');
  });

  global.addEventListener("online", function (ev) {
    m.onLine = true;
    m.emit('online');
  });

} else {
  m = {available:false, onLine:true};
}

module.exports = m;

};

__mod["protocol"]=function(module) { var exports = module.exports;
"use strict";
var Buf = require('./buf');
var utf8 = require('./utf8');

// Version of this protocol
exports.Version = 1;

// Message types
var MsgTypeSingleReq     = exports.MsgTypeSingleReq =     'r'.charCodeAt(0),
    MsgTypeStreamReq     = exports.MsgTypeStreamReq =     's'.charCodeAt(0),
    MsgTypeStreamReqPart = exports.MsgTypeStreamReqPart = 'p'.charCodeAt(0),
    MsgTypeSingleRes     = exports.MsgTypeSingleRes =     'R'.charCodeAt(0),
    MsgTypeStreamRes     = exports.MsgTypeStreamRes =     'S'.charCodeAt(0),
    MsgTypeErrorRes      = exports.MsgTypeErrorRes =      'E'.charCodeAt(0),
    MsgTypeRetryRes      = exports.MsgTypeRetryRes =      'e'.charCodeAt(0),
    MsgTypeNotification  = exports.MsgTypeNotification =  'n'.charCodeAt(0),
    MsgTypeHeartbeat     = exports.MsgTypeHeartbeat =     'h'.charCodeAt(0),
    MsgTypeProtocolError = exports.MsgTypeProtocolError = 'f'.charCodeAt(0);

// ProtocolError codes
exports.ErrorAbnormal    = 0
exports.ErrorUnsupported = 1;
exports.ErrorInvalidMsg  = 2;
exports.ErrorTimeout     = 3;

// Maximum value of a heartbeat's "load"
exports.HeartbeatMsgMaxLoad = 0xffff;

// ==============================================================================================
// Binary (byte) protocol

function copyBufFixnum(b, start, n, digits) {
  var i = start || 0, y = 0, c, s = n.toString(16), z = digits - s.length;
  for (; z--;) { b[i++] = 48; }
  for (; !isNaN(c = s.charCodeAt(y++));) { b[i++] = c; }
}

function makeBufFixnum(n, digits) {
  var b = Buf(digits);
  copyBufFixnum(b, 0, n, digits);
  return b;
}

// Note: This code assumes parseInt accepts a Buf

exports.binary = {

  makeFixnum: makeBufFixnum,

  versionBuf: makeBufFixnum(exports.Version, 2),

  parseVersion: function (b) {
    return parseInt(b, 16);
  },

  // Parses a byte buffer containing a message (not including payload data.)
  // If t is MsgTypeHeartbeat, wait==load, size==time.
  // -> {t:string, id:Buf, name:string, wait:int size:int} | null
  parseMsg: function (b) {
    var t, id, name, namez, wait = 0, size = 0, z;

    t = b[0];
    z = 1;

    if (t === MsgTypeHeartbeat) {
      wait = parseInt(b.slice(z, z + 4), 16);
      z += 4;
    } else if (t !== MsgTypeNotification && t !== MsgTypeProtocolError) {
      id = b.slice(z, z + 4);
      z += 4;
    }

    if (t == MsgTypeSingleReq || t == MsgTypeStreamReq || t == MsgTypeNotification) {
      namez = parseInt(b.slice(z, z + 3), 16);
      z += 3;
      name = b.slice(z, z+namez).toString();
      z += namez;
    } else if (t === MsgTypeRetryRes) {
      wait = parseInt(b.slice(z, z + 8), 16);
      z += 8
    }

    size = parseInt(b.slice(z, z + 8), 16);

    return {t:t, id:id, name:name, wait:wait, size:size};
  },

  // Create a buf representing a message (w/o any payload)
  makeMsg: function (t, id, name, wait, size) {
    var b, nameb, z = id ? 13 : 9;

    if (name && name.length !== 0) {
      nameb = Buf.fromString(name);
      z += 3 + nameb.length;
    }

    b = Buf(z);

    b[0] = t;
    z = 1;

    if (id && id.length === 4) {
      if (typeof id === 'string') {
        b[1] = id.charCodeAt(0);
        b[2] = id.charCodeAt(1);
        b[3] = id.charCodeAt(2);
        b[4] = id.charCodeAt(3);
      } else {
        b[1] = id[0];
        b[2] = id[1];
        b[3] = id[2];
        b[4] = id[3];
      }
      z += 4;
    }

    if (name && name.length !== 0) {
      nameb = Buf.fromString(name);
      copyBufFixnum(b, z, nameb.length, 3);
      z += 3;
      nameb.copy(b, z);
      z += nameb.length;
    }

    if (t === MsgTypeRetryRes) {
      copyBufFixnum(b, z, wait, 8);
      z += 8
    }

    copyBufFixnum(b, z, size, 8);

    return b;
  },

  // Create a buf representing a heartbeat message
  makeHeartbeatMsg: function(load) {
    var b = Buf(13), z = 1;
    b[0] = MsgTypeHeartbeat;
    copyBufFixnum(b, z, load, 4);
    z += 4;
    copyBufFixnum(b, z, Math.round((new Date).getTime()/1000), 8);
    z += 8;
    return b;
  }
};


// ==============================================================================================
// Text protocol

var zeroes = '00000000';

function makeStrFixnum(n, digits) {
  var s = n.toString(16);
  return zeroes.substr(0, digits - s.length) + s;
}

exports.text = {

  makeFixnum: makeStrFixnum,

  versionBuf: makeStrFixnum(exports.Version, 2),

  parseVersion: function (buf) {
    return parseInt(buf.substr(0,2), 16);
  },

  // Parses a text string containing a message (not including payload data.)
  // If t is MsgTypeHeartbeat, wait==load, size==time.
  // -> {t:string, id:Buf, name:string, wait:int size:int} | null
  parseMsg: function (s) {
    // "r001004echo00000005" => ('r', "001", "echo", 5)
    // "R00100000005"        => ('R', "001", "", 5)
    var t, id, name, wait = 0, size = 0, z;

    t = s.charCodeAt(0);
    z = 1;

    if (t === MsgTypeHeartbeat) {
      wait = parseInt(s.substr(z, 4), 16);
      z += 4;
    } else if (t !== MsgTypeNotification && t !== MsgTypeProtocolError) {
      id = s.substr(z, 4);
      z += 4;
    }

    if (t == MsgTypeSingleReq || t == MsgTypeStreamReq || t == MsgTypeNotification) {
      name = s.substring(z + 3, s.length - 8);
    } else if (t == MsgTypeRetryRes) {
      wait = parseInt(s.substr(z, 8), 16);
      z += 8
    }

    size = parseInt(s.substr(s.length - 8), 16);

    return {t:t, id:id, name:name, wait:wait, size:size};
  },


  // Create a text string representing a message (w/o any payload.)
  makeMsg: function (t, id, name, wait, size) {
    var b = String.fromCharCode(t);

    if (id && id.length === 4) {
      b += id;
    }

    if (name && name.length !== 0) {
      b += makeStrFixnum(utf8.sizeOf(name), 3);
      b += name;
    }

    if (t === MsgTypeRetryRes) {
      b += makeStrFixnum(wait, 8);
    }

    b += makeStrFixnum(size, 8);

    return b;
  },

  // Create a text string representing a heartbeat message
  makeHeartbeatMsg: function(load) {
    var s = String.fromCharCode(MsgTypeHeartbeat);
    s += makeStrFixnum(load, 4);
    s += makeStrFixnum(Math.round((new Date).getTime()/1000), 8);
    return s;
  }

}; // exports.text


};

__mod["utf8"]=function(module) { var exports = module.exports;
"use strict";
//
// decode(Buf, [start], [end]) -> String
// encode(String, BufFactory) -> Buf
// sizeOf(String) -> int
//

// Returns the number of bytes needed to represent string `s` as UTF8
function sizeOf(s) {
  var z = 0, i = 0, c;
  for (; c = s.charCodeAt(i++); z += (c >> 11 ? 3 : c >> 7 ? 2 : 1) );
  return z;
}
exports.sizeOf = sizeOf;

function mask8(c) {
  return 0xff & c;
}

if (typeof TextDecoder !== 'undefined') {
  // ============================================================================================
  // Native TextDecoder/TextEncoder implementation
  var decoder = new TextDecoder('utf8');
  var encoder = new TextEncoder('utf8');

  exports.decode = function decode(b, start, end) {
    if (start || end) {
      if (!start) start = 0;
      b = b.slice(start, end || b.length - start);
    }
    return decoder.decode(b);
  };

  exports.encode = function encode(s, Buf) {
    return Buf(encoder.encode(s));
  };

} else {
  // ============================================================================================
  // JS implementation

  exports.decode = function decode(b, start, end) {
    var i = start || 0, e = (end || b.length - i), c, lead, s = '';
    for (i = 0; i < e; ) {
      c = b[i++];
      lead = mask8(c);
      if (lead < 0x80) {
        // single byte
      } else if ((lead >> 5) == 0x6) {
        c = ((c << 6) & 0x7ff) + (b[i++] & 0x3f);
      } else if ((lead >> 4) == 0xe) {
        c = ((c << 12) & 0xffff) + ((mask8(b[i++]) << 6) & 0xfff);
        c += b[i++] & 0x3f;
      } else if ((lead >> 3) == 0x1e) {
        c = ((c << 18) & 0x1fffff) + ((mask8(b[i++]) << 12) & 0x3ffff);
        c += (mask8(b[i++]) << 6) & 0xfff;
        c += b[i++] & 0x3f;
      }
      s += String.fromCharCode(c);
    }

    return s;
  };

  exports.encode = function encode(s, Buf) {
    var i = 0, e = s.length, c, j = 0, b = Buf(sizeOf(s));
    for (; i !== e;) {
      c = s.charCodeAt(i++);
      // TODO FIXME: charCodeAt returns UTF16-like codepoints, not UTF32 codepoints, meaning that
      // this code only works for BMP. However, current ES only supports BMP. Ultimately we should
      // dequeue a second UTF16 codepoint when c>BMP.
      if (c < 0x80) {
        b[j++] = c;
      } else if (c < 0x800) {
        b[j++] = (c >> 6)   | 0xc0;
        b[j++] = (c & 0x3f) | 0x80;
      } else if (c < 0x10000) {
        b[j++] = (c >> 12)          | 0xe0;
        b[j++] = ((c >> 6) & 0x3f)  | 0x80;
        b[j++] = (c & 0x3f)         | 0x80;
      } else {
        b[j++] = (c >> 18)          | 0xf0;
        b[j++] = ((c >> 12) & 0x3f) | 0x80;
        b[j++] = ((c >> 6) & 0x3f)  | 0x80;
        b[j++] = (c & 0x3f)         | 0x80;
      }
    }
    return b;
  };

}

// var s = '∆åßf'; // '日本語'
// var b = exports.encode(s);
// console.log('encode("'+s+'") =>', b);
// console.log('decode(',b,') =>', exports.decode(b));

};

__main=function(module) { var exports = module.exports;
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



};


__mainapi = {exports:{}};
__main(__mainapi);

global.gotalk = __mainapi.exports;
})(typeof window !== "undefined" ? window : this);
