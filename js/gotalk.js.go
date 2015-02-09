package gotalkjs
const BrowserLibSizeString = "35184"
const BrowserLibSHA1Base64 = "DwVU/vfrG2rK4jax0TV70+7MdnU="
const BrowserLibETag = "\"DwVU/vfrG2rK4jax0TV70+7MdnU=\""
const BrowserLibString = ""+
  "(function(global){\n"+
  "\"use strict\";\n"+
  "\n"+
  "var __mod = {}, __api = {}, __main, __mainapi;\n"+
  "var evalDepth = 0;\n"+
  "\n"+
  "var require = function(name) {\n"+
  "  var m, id = name.replace(/^.\\//, \"\");\n"+
  "  m = __api[id];\n"+
  "  //console.log('require', {name:name, id:id, exports:m});\n"+
  "  if (!m) {\n"+
  "    var prefix = ''; for (var i = 0; i < evalDepth; i++) {\n"+
  "      prefix += '. ';\n"+
  "    }\n"+
  "    var f = __mod[id];\n"+
  "    if (f && evalDepth < 100) {\n"+
  "      __mod[id] = null;\n"+
  "      __api[id] = {exports:{}};\n"+
  "      ++evalDepth;\n"+
  "      f(__api[id]);\n"+
  "      --evalDepth;\n"+
  "      __api[id] = __api[id].exports;\n"+
  "    }\n"+
  "    m = __api[id];\n"+
  "  }\n"+
  "  return m;\n"+
  "};\n"+
  "\n"+
  "__mod[\"EventEmitter\"]=function(module) { var exports = module.exports;\n"+
  "\n"+
  "function EventEmitter() {}\n"+
  "module.exports = EventEmitter;\n"+
  "\n"+
  "EventEmitter.prototype.addListener = function (type, listener) {\n"+
  "  if (typeof listener !== 'function') throw TypeError('listener must be a function');\n"+
  "  if (!this.__events) {\n"+
  "    Object.defineProperty(this, '__events', {value:{}, enumerable:false, writable:true});\n"+
  "    this.__events[type] = [listener];\n"+
  "    return this;\n"+
  "  }\n"+
  "  var listeners = this.__events[type];\n"+
  "  if (listeners === undefined) {\n"+
  "    this.__events[type] = [listener];\n"+
  "    return this;\n"+
  "  }\n"+
  "  listeners.push(listener);\n"+
  "  return this;\n"+
  "};\n"+
  "\n"+
  "EventEmitter.prototype.on = EventEmitter.prototype.addListener;\n"+
  "\n"+
  "EventEmitter.prototype.once = function (type, listener) {\n"+
  "  var fired = false;\n"+
  "  var trigger_event_once = function() {\n"+
  "    this.removeListener(type, trigger_event_once);\n"+
  "    if (!fired) {\n"+
  "      fired = true;\n"+
  "      listener.apply(this, arguments);\n"+
  "    }\n"+
  "  }\n"+
  "  return this.on(type, trigger_event_once);\n"+
  "};\n"+
  "\n"+
  "EventEmitter.prototype.removeListener = function (type, listener) {\n"+
  "  var p, listeners = this.__events ? this.__events[type] : undefined;\n"+
  "  if (listeners !== undefined) {\n"+
  "    while ((p = listeners.indexOf(listener)) !== -1) {\n"+
  "      listeners.splice(p,1);\n"+
  "    }\n"+
  "    if (listeners.length === 0) {\n"+
  "      delete this.__events[type];\n"+
  "    }\n"+
  "    return listeners.length;\n"+
  "  }\n"+
  "  return this;\n"+
  "};\n"+
  "\n"+
  "EventEmitter.prototype.removeAllListeners = function (type) {\n"+
  "  if (this.__events) {\n"+
  "    if (type) {\n"+
  "      delete this.__events[type];\n"+
  "    } else {\n"+
  "      delete this.__events;\n"+
  "    }\n"+
  "  }\n"+
  "};\n"+
  "\n"+
  "EventEmitter.prototype.listeners = function (type) {\n"+
  "  return type ? (this.__events ? this.__events[type] : undefined) : this.__events;\n"+
  "};\n"+
  "\n"+
  "EventEmitter.prototype.emit = function (type) {\n"+
  "  var listeners = this.__events ? this.__events[type] : undefined;\n"+
  "  if (listeners === undefined) {\n"+
  "    return false;\n"+
  "  }\n"+
  "  var i = 0, L = listeners.length, args = Array.prototype.slice.call(arguments,1);\n"+
  "  for (; i !== L; ++i) {\n"+
  "    if (!listeners[i]) {\n"+
  "      console.log('e', type, i, args);\n"+
  "    }\n"+
  "    listeners[i].apply(this, args);\n"+
  "  }\n"+
  "  return true;\n"+
  "};\n"+
  "\n"+
  "EventEmitter.mixin = function mixin(obj) {\n"+
  "  var proto = obj;\n"+
  "  while (proto) {\n"+
  "    if (proto.__proto__ === Object.prototype) {\n"+
  "      proto.__proto__ = EventEmitter.prototype;\n"+
  "      return obj;\n"+
  "    }\n"+
  "    proto = proto.__proto__;\n"+
  "  }\n"+
  "  return obj;\n"+
  "};\n"+
  "\n"+
  "\n"+
  "};\n"+
  "\n"+
  "__mod[\"buf\"]=function(module) { var exports = module.exports;\n"+
  "\"use strict\";\n"+
  "var Buf;\n"+
  "\n"+
  "if (typeof Uint8Array !== 'undefined') {\n"+
  "\n"+
  "var utf8 = require('./utf8');\n"+
  "\n"+
  "Uint8Array.prototype.toString = function (encoding, start, end) {\n"+
  "  // assumes buffer contains UTF8-encoded text\n"+
  "  return utf8.decode(this, start, end);\n"+
  "};\n"+
  "\n"+
  "Uint8Array.prototype.slice = Uint8Array.prototype.subarray;\n"+
  "\n"+
  "// Copies data from a region of this buffer to a region in the target buffer.\n"+
  "// copy(targetBuffer, [targetStart], [sourceStart], [sourceEnd]) -> Buf\n"+
  "Uint8Array.prototype.copy = function (targetBuffer, targetStart, sourceStart, sourceEnd) {\n"+
  "  var srcBuf = this;\n"+
  "  if (sourceStart) {\n"+
  "    srcBuf = srcBuf.slice(sourceStart, sourceEnd || srcBuf.length - sourceStart);\n"+
  "  }\n"+
  "  targetBuffer.set(srcBuf, targetStart || 0);\n"+
  "};\n"+
  "\n"+
  "\n"+
  "// Buf(Buf) -> Buf\n"+
  "// Buf(size int) -> Buf\n"+
  "// Buf(ArrayBuffer) -> Buf\n"+
  "Buf = function Buf(v) {\n"+
  "  return v instanceof Uint8Array ? v :\n"+
  "    new Uint8Array(\n"+
  "      v instanceof ArrayBuffer ? v :\n"+
  "      new ArrayBuffer(v)\n"+
  "    );\n"+
  "};\n"+
  "\n"+
  "Buf.isBuf = function (v) {\n"+
  "  return v instanceof Uint8Array;\n"+
  "};\n"+
  "\n"+
  "Buf.fromString = function (s, encoding) {\n"+
  "  return utf8.encode(s, Buf);\n"+
  "};\n"+
  "\n"+
  "}\n"+
  "\n"+
  "module.exports = Buf;\n"+
  "\n"+
  "};\n"+
  "\n"+
  "__mod[\"keepalive\"]=function(module) { var exports = module.exports;\n"+
  "\"use strict\";\n"+
  "// Stay connected by automatically reconnecting w/ exponential back-off.\n"+
  "\n"+
  "var netAccess = require('./netaccess');\n"+
  "var protocol = require('./protocol');\n"+
  "\n"+
  "// `s` must conform to interface { connect(addr string, cb function(Error)) }\n"+
  "// Returns an object {\n"+
  "//   isConnected bool  // true if currently connected\n"+
  "//   isEnabled bool    // true if enabled\n"+
  "//   enable()          // enables staying connected\n"+
  "//   disable()         // disables trying to stay connected\n"+
  "// }\n"+
  "var keepalive = function(s, addr, minReconnectDelay, maxReconnectDelay) {\n"+
  "  if (!minReconnectDelay) {\n"+
  "    minReconnectDelay = 500;\n"+
  "  } else if (minReconnectDelay < 100) {\n"+
  "    minReconnectDelay = 100;\n"+
  "  }\n"+
  "\n"+
  "  if (!maxReconnectDelay || maxReconnectDelay < minReconnectDelay) {\n"+
  "    maxReconnectDelay = 5000;\n"+
  "  }\n"+
  "\n"+
  "  var ctx, open, retry, delay = 0, openTimer, opentime;\n"+
  "\n"+
  "  ctx = {\n"+
  "    isEnabled: false,\n"+
  "    isConnected: false,\n"+
  "    enable: function() {\n"+
  "      if (!ctx.enabled) {\n"+
  "        ctx.enabled = true;\n"+
  "        delay = 0;\n"+
  "        if (!ctx.isConnected) {\n"+
  "          open();\n"+
  "        }\n"+
  "      }\n"+
  "    },\n"+
  "    disable: function() {\n"+
  "      if (ctx.enabled) {\n"+
  "        clearTimeout(openTimer);\n"+
  "        ctx.enabled = false;\n"+
  "        delay = 0;\n"+
  "      }\n"+
  "    }\n"+
  "  };\n"+
  "\n"+
  "  open = function() {\n"+
  "    clearTimeout(openTimer);\n"+
  "    s.open(addr, function(err) {\n"+
  "      opentime = new Date;\n"+
  "      if (err) {\n"+
  "        retry(err);\n"+
  "      } else {\n"+
  "        delay = 0;\n"+
  "        ctx.isConnected = true;\n"+
  "        s.once('close', retry);\n"+
  "      }\n"+
  "    });\n"+
  "  };\n"+
  "\n"+
  "  retry = function(err) {\n"+
  "    clearTimeout(openTimer);\n"+
  "    ctx.isConnected = false;\n"+
  "    if (!ctx.enabled) {\n"+
  "      return;\n"+
  "    }\n"+
  "    if (netAccess.available && !netAccess.onLine && \n"+
  "        !(typeof document !== 'undefined' &&\n"+
  "          document.location &&\n"+
  "          document.location.hostname !== 'localhost' &&\n"+
  "          document.location.hostname !== '127.0.0.1' &&\n"+
  "          document.location.hostname !== '[::1]') )\n"+
  "    {\n"+
  "      netAccess.once('online', retry);\n"+
  "      delay = 0;\n"+
  "      return;\n"+
  "    }\n"+
  "    if (err) {\n"+
  "      if (err.isGotalkProtocolError) {\n"+
  "        if (err.code === protocol.ErrorTimeout) {\n"+
  "          delay = 0;\n"+
  "        } else {\n"+
  "          // We shouldn't retry with the same version of our gotalk library.\n"+
  "          // However, the only sensible thing to do in this case is to let the user code react to\n"+
  "          // the error passed to the close event (e.g. to show a \"can't talk to server\" UI), and\n"+
  "          // retry in maxReconnectDelay.\n"+
  "          // User code can choose to call `disable()` on its keepalive object in this case.\n"+
  "          delay = maxReconnectDelay;\n"+
  "        }\n"+
  "      } else {\n"+
  "        // increase back off in case of an error\n"+
  "        delay = delay ? Math.min(maxReconnectDelay, delay * 2) : minReconnectDelay;\n"+
  "      }\n"+
  "    } else {\n"+
  "      delay = Math.max(0, minReconnectDelay - ((new Date) - opentime));\n"+
  "    }\n"+
  "    openTimer = setTimeout(open, delay);\n"+
  "  };\n"+
  "\n"+
  "  return ctx;\n"+
  "};\n"+
  "\n"+
  "module.exports = keepalive;\n"+
  "\n"+
  "};\n"+
  "\n"+
  "__mod[\"netaccess\"]=function(module) { var exports = module.exports;\n"+
  "\"use strict\";\n"+
  "var EventEmitter = require('./EventEmitter');\n"+
  "var m;\n"+
  "\n"+
  "if (typeof global !== 'undefined' && global.addEventListener) {\n"+
  "  m = Object.create(EventEmitter.prototype, {\n"+
  "    available: {value:true, enumerable:true},\n"+
  "    onLine:    {value:true, enumerable:true, writable:true}\n"+
  "  });\n"+
  "\n"+
  "  if (typeof navigator !== 'undefined') {\n"+
  "    m.onLine = navigator.onLine;\n"+
  "  }\n"+
  "\n"+
  "  global.addEventListener(\"offline\", function (ev) {\n"+
  "    m.onLine = false;\n"+
  "    m.emit('offline');\n"+
  "  });\n"+
  "\n"+
  "  global.addEventListener(\"online\", function (ev) {\n"+
  "    m.onLine = true;\n"+
  "    m.emit('online');\n"+
  "  });\n"+
  "\n"+
  "} else {\n"+
  "  m = {available:false, onLine:true};\n"+
  "}\n"+
  "\n"+
  "module.exports = m;\n"+
  "\n"+
  "};\n"+
  "\n"+
  "__mod[\"protocol\"]=function(module) { var exports = module.exports;\n"+
  "\"use strict\";\n"+
  "var Buf = require('./buf');\n"+
  "var utf8 = require('./utf8');\n"+
  "\n"+
  "// Version of this protocol\n"+
  "exports.Version = 1;\n"+
  "\n"+
  "// Message types\n"+
  "var MsgTypeSingleReq     = exports.MsgTypeSingleReq =     'r'.charCodeAt(0),\n"+
  "    MsgTypeStreamReq     = exports.MsgTypeStreamReq =     's'.charCodeAt(0),\n"+
  "    MsgTypeStreamReqPart = exports.MsgTypeStreamReqPart = 'p'.charCodeAt(0),\n"+
  "    MsgTypeSingleRes     = exports.MsgTypeSingleRes =     'R'.charCodeAt(0),\n"+
  "    MsgTypeStreamRes     = exports.MsgTypeStreamRes =     'S'.charCodeAt(0),\n"+
  "    MsgTypeErrorRes      = exports.MsgTypeErrorRes =      'E'.charCodeAt(0),\n"+
  "    MsgTypeRetryRes      = exports.MsgTypeRetryRes =      'e'.charCodeAt(0),\n"+
  "    MsgTypeNotification  = exports.MsgTypeNotification =  'n'.charCodeAt(0),\n"+
  "    MsgTypeHeartbeat     = exports.MsgTypeHeartbeat =     'h'.charCodeAt(0),\n"+
  "    MsgTypeProtocolError = exports.MsgTypeProtocolError = 'f'.charCodeAt(0);\n"+
  "\n"+
  "// ProtocolError codes\n"+
  "exports.ErrorAbnormal    = 0\n"+
  "exports.ErrorUnsupported = 1;\n"+
  "exports.ErrorInvalidMsg  = 2;\n"+
  "exports.ErrorTimeout     = 3;\n"+
  "\n"+
  "// Maximum value of a heartbeat's \"load\"\n"+
  "exports.HeartbeatMsgMaxLoad = 0xffff;\n"+
  "\n"+
  "// ==============================================================================================\n"+
  "// Binary (byte) protocol\n"+
  "\n"+
  "function copyBufFixnum(b, start, n, digits) {\n"+
  "  var i = start || 0, y = 0, c, s = n.toString(16), z = digits - s.length;\n"+
  "  for (; z--;) { b[i++] = 48; }\n"+
  "  for (; !isNaN(c = s.charCodeAt(y++));) { b[i++] = c; }\n"+
  "}\n"+
  "\n"+
  "function makeBufFixnum(n, digits) {\n"+
  "  var b = Buf(digits);\n"+
  "  copyBufFixnum(b, 0, n, digits);\n"+
  "  return b;\n"+
  "}\n"+
  "\n"+
  "// Note: This code assumes parseInt accepts a Buf\n"+
  "\n"+
  "exports.binary = {\n"+
  "\n"+
  "  makeFixnum: makeBufFixnum,\n"+
  "\n"+
  "  versionBuf: makeBufFixnum(exports.Version, 2),\n"+
  "\n"+
  "  parseVersion: function (b) {\n"+
  "    return parseInt(b, 16);\n"+
  "  },\n"+
  "\n"+
  "  // Parses a byte buffer containing a message (not including payload data.)\n"+
  "  // If t is MsgTypeHeartbeat, wait==load, size==time.\n"+
  "  // -> {t:string, id:Buf, name:string, wait:int size:int} | null\n"+
  "  parseMsg: function (b) {\n"+
  "    var t, id, name, namez, wait = 0, size = 0, z;\n"+
  "\n"+
  "    t = b[0];\n"+
  "    z = 1;\n"+
  "\n"+
  "    if (t === MsgTypeHeartbeat) {\n"+
  "      wait = parseInt(b.slice(z, z + 4), 16);\n"+
  "      z += 4;\n"+
  "    } else if (t !== MsgTypeNotification && t !== MsgTypeProtocolError) {\n"+
  "      id = b.slice(z, z + 4);\n"+
  "      z += 4;\n"+
  "    }\n"+
  "\n"+
  "    if (t == MsgTypeSingleReq || t == MsgTypeStreamReq || t == MsgTypeNotification) {\n"+
  "      namez = parseInt(b.slice(z, z + 3), 16);\n"+
  "      z += 3;\n"+
  "      name = b.slice(z, z+namez).toString();\n"+
  "      z += namez;\n"+
  "    } else if (t === MsgTypeRetryRes) {\n"+
  "      wait = parseInt(b.slice(z, z + 8), 16);\n"+
  "      z += 8\n"+
  "    }\n"+
  "\n"+
  "    size = parseInt(b.slice(z, z + 8), 16);\n"+
  "\n"+
  "    return {t:t, id:id, name:name, wait:wait, size:size};\n"+
  "  },\n"+
  "\n"+
  "  // Create a buf representing a message (w/o any payload)\n"+
  "  makeMsg: function (t, id, name, wait, size) {\n"+
  "    var b, nameb, z = id ? 13 : 9;\n"+
  "\n"+
  "    if (name && name.length !== 0) {\n"+
  "      nameb = Buf.fromString(name);\n"+
  "      z += 3 + nameb.length;\n"+
  "    }\n"+
  "\n"+
  "    b = Buf(z);\n"+
  "\n"+
  "    b[0] = t;\n"+
  "    z = 1;\n"+
  "\n"+
  "    if (id && id.length === 4) {\n"+
  "      if (typeof id === 'string') {\n"+
  "        b[1] = id.charCodeAt(0);\n"+
  "        b[2] = id.charCodeAt(1);\n"+
  "        b[3] = id.charCodeAt(2);\n"+
  "        b[4] = id.charCodeAt(3);\n"+
  "      } else {\n"+
  "        b[1] = id[0];\n"+
  "        b[2] = id[1];\n"+
  "        b[3] = id[2];\n"+
  "        b[4] = id[3];\n"+
  "      }\n"+
  "      z += 4;\n"+
  "    }\n"+
  "\n"+
  "    if (name && name.length !== 0) {\n"+
  "      nameb = Buf.fromString(name);\n"+
  "      copyBufFixnum(b, z, nameb.length, 3);\n"+
  "      z += 3;\n"+
  "      nameb.copy(b, z);\n"+
  "      z += nameb.length;\n"+
  "    }\n"+
  "\n"+
  "    if (t === MsgTypeRetryRes) {\n"+
  "      copyBufFixnum(b, z, wait, 8);\n"+
  "      z += 8\n"+
  "    }\n"+
  "\n"+
  "    copyBufFixnum(b, z, size, 8);\n"+
  "\n"+
  "    return b;\n"+
  "  },\n"+
  "\n"+
  "  // Create a buf representing a heartbeat message\n"+
  "  makeHeartbeatMsg: function(load) {\n"+
  "    var b = Buf(13), z = 1;\n"+
  "    b[0] = MsgTypeHeartbeat;\n"+
  "    copyBufFixnum(b, z, load, 4);\n"+
  "    z += 4;\n"+
  "    copyBufFixnum(b, z, Math.round((new Date).getTime()/1000), 8);\n"+
  "    z += 8;\n"+
  "    return b;\n"+
  "  }\n"+
  "};\n"+
  "\n"+
  "\n"+
  "// ==============================================================================================\n"+
  "// Text protocol\n"+
  "\n"+
  "var zeroes = '00000000';\n"+
  "\n"+
  "function makeStrFixnum(n, digits) {\n"+
  "  var s = n.toString(16);\n"+
  "  return zeroes.substr(0, digits - s.length) + s;\n"+
  "}\n"+
  "\n"+
  "exports.text = {\n"+
  "\n"+
  "  makeFixnum: makeStrFixnum,\n"+
  "\n"+
  "  versionBuf: makeStrFixnum(exports.Version, 2),\n"+
  "\n"+
  "  parseVersion: function (buf) {\n"+
  "    return parseInt(buf.substr(0,2), 16);\n"+
  "  },\n"+
  "\n"+
  "  // Parses a text string containing a message (not including payload data.)\n"+
  "  // If t is MsgTypeHeartbeat, wait==load, size==time.\n"+
  "  // -> {t:string, id:Buf, name:string, wait:int size:int} | null\n"+
  "  parseMsg: function (s) {\n"+
  "    // \"r001004echo00000005\" => ('r', \"001\", \"echo\", 5)\n"+
  "    // \"R00100000005\"        => ('R', \"001\", \"\", 5)\n"+
  "    var t, id, name, wait = 0, size = 0, z;\n"+
  "\n"+
  "    t = s.charCodeAt(0);\n"+
  "    z = 1;\n"+
  "\n"+
  "    if (t === MsgTypeHeartbeat) {\n"+
  "      wait = parseInt(s.substr(z, 4), 16);\n"+
  "      z += 4;\n"+
  "    } else if (t !== MsgTypeNotification && t !== MsgTypeProtocolError) {\n"+
  "      id = s.substr(z, 4);\n"+
  "      z += 4;\n"+
  "    }\n"+
  "\n"+
  "    if (t == MsgTypeSingleReq || t == MsgTypeStreamReq || t == MsgTypeNotification) {\n"+
  "      name = s.substring(z + 3, s.length - 8);\n"+
  "    } else if (t == MsgTypeRetryRes) {\n"+
  "      wait = parseInt(s.substr(z, 8), 16);\n"+
  "      z += 8\n"+
  "    }\n"+
  "\n"+
  "    size = parseInt(s.substr(s.length - 8), 16);\n"+
  "\n"+
  "    return {t:t, id:id, name:name, wait:wait, size:size};\n"+
  "  },\n"+
  "\n"+
  "\n"+
  "  // Create a text string representing a message (w/o any payload.)\n"+
  "  makeMsg: function (t, id, name, wait, size) {\n"+
  "    var b = String.fromCharCode(t);\n"+
  "\n"+
  "    if (id && id.length === 4) {\n"+
  "      b += id;\n"+
  "    }\n"+
  "\n"+
  "    if (name && name.length !== 0) {\n"+
  "      b += makeStrFixnum(utf8.sizeOf(name), 3);\n"+
  "      b += name;\n"+
  "    }\n"+
  "\n"+
  "    if (t === MsgTypeRetryRes) {\n"+
  "      b += makeStrFixnum(wait, 8);\n"+
  "    }\n"+
  "\n"+
  "    b += makeStrFixnum(size, 8);\n"+
  "\n"+
  "    return b;\n"+
  "  },\n"+
  "\n"+
  "  // Create a text string representing a heartbeat message\n"+
  "  makeHeartbeatMsg: function(load) {\n"+
  "    var s = String.fromCharCode(MsgTypeHeartbeat);\n"+
  "    s += makeStrFixnum(load, 4);\n"+
  "    s += makeStrFixnum(Math.round((new Date).getTime()/1000), 8);\n"+
  "    return s;\n"+
  "  }\n"+
  "\n"+
  "}; // exports.text\n"+
  "\n"+
  "\n"+
  "};\n"+
  "\n"+
  "__mod[\"utf8\"]=function(module) { var exports = module.exports;\n"+
  "\"use strict\";\n"+
  "//\n"+
  "// decode(Buf, [start], [end]) -> String\n"+
  "// encode(String, BufFactory) -> Buf\n"+
  "// sizeOf(String) -> int\n"+
  "//\n"+
  "\n"+
  "// Returns the number of bytes needed to represent string `s` as UTF8\n"+
  "function sizeOf(s) {\n"+
  "  var z = 0, i = 0, c;\n"+
  "  for (; c = s.charCodeAt(i++); z += (c >> 11 ? 3 : c >> 7 ? 2 : 1) );\n"+
  "  return z;\n"+
  "}\n"+
  "exports.sizeOf = sizeOf;\n"+
  "\n"+
  "function mask8(c) {\n"+
  "  return 0xff & c;\n"+
  "}\n"+
  "\n"+
  "if (typeof TextDecoder !== 'undefined') {\n"+
  "  // ============================================================================================\n"+
  "  // Native TextDecoder/TextEncoder implementation\n"+
  "  var decoder = new TextDecoder('utf8');\n"+
  "  var encoder = new TextEncoder('utf8');\n"+
  "\n"+
  "  exports.decode = function decode(b, start, end) {\n"+
  "    if (start || end) {\n"+
  "      if (!start) start = 0;\n"+
  "      b = b.slice(start, end || b.length - start);\n"+
  "    }\n"+
  "    return decoder.decode(b);\n"+
  "  };\n"+
  "\n"+
  "  exports.encode = function encode(s, Buf) {\n"+
  "    return Buf(encoder.encode(s));\n"+
  "  };\n"+
  "\n"+
  "} else {\n"+
  "  // ============================================================================================\n"+
  "  // JS implementation\n"+
  "\n"+
  "  exports.decode = function decode(b, start, end) {\n"+
  "    var i = start || 0, e = (end || b.length - i), c, lead, s = '';\n"+
  "    for (i = 0; i < e; ) {\n"+
  "      c = b[i++];\n"+
  "      lead = mask8(c);\n"+
  "      if (lead < 0x80) {\n"+
  "        // single byte\n"+
  "      } else if ((lead >> 5) == 0x6) {\n"+
  "        c = ((c << 6) & 0x7ff) + (b[i++] & 0x3f);\n"+
  "      } else if ((lead >> 4) == 0xe) {\n"+
  "        c = ((c << 12) & 0xffff) + ((mask8(b[i++]) << 6) & 0xfff);\n"+
  "        c += b[i++] & 0x3f;\n"+
  "      } else if ((lead >> 3) == 0x1e) {\n"+
  "        c = ((c << 18) & 0x1fffff) + ((mask8(b[i++]) << 12) & 0x3ffff);\n"+
  "        c += (mask8(b[i++]) << 6) & 0xfff;\n"+
  "        c += b[i++] & 0x3f;\n"+
  "      }\n"+
  "      s += String.fromCharCode(c);\n"+
  "    }\n"+
  "\n"+
  "    return s;\n"+
  "  };\n"+
  "\n"+
  "  exports.encode = function encode(s, Buf) {\n"+
  "    var i = 0, e = s.length, c, j = 0, b = Buf(sizeOf(s));\n"+
  "    for (; i !== e;) {\n"+
  "      c = s.charCodeAt(i++);\n"+
  "      // TODO FIXME: charCodeAt returns UTF16-like codepoints, not UTF32 codepoints, meaning that\n"+
  "      // this code only works for BMP. However, current ES only supports BMP. Ultimately we should\n"+
  "      // dequeue a second UTF16 codepoint when c>BMP.\n"+
  "      if (c < 0x80) {\n"+
  "        b[j++] = c;\n"+
  "      } else if (c < 0x800) {\n"+
  "        b[j++] = (c >> 6)   | 0xc0;\n"+
  "        b[j++] = (c & 0x3f) | 0x80;\n"+
  "      } else if (c < 0x10000) {\n"+
  "        b[j++] = (c >> 12)          | 0xe0;\n"+
  "        b[j++] = ((c >> 6) & 0x3f)  | 0x80;\n"+
  "        b[j++] = (c & 0x3f)         | 0x80;\n"+
  "      } else {\n"+
  "        b[j++] = (c >> 18)          | 0xf0;\n"+
  "        b[j++] = ((c >> 12) & 0x3f) | 0x80;\n"+
  "        b[j++] = ((c >> 6) & 0x3f)  | 0x80;\n"+
  "        b[j++] = (c & 0x3f)         | 0x80;\n"+
  "      }\n"+
  "    }\n"+
  "    return b;\n"+
  "  };\n"+
  "\n"+
  "}\n"+
  "\n"+
  "// var s = '\xe2\x88\x86\xc3\xa5\xc3\x9ff'; // '\xe6\x97\xa5\xe6\x9c\xac\xe8\xaa\x9e'\n"+
  "// var b = exports.encode(s);\n"+
  "// console.log('encode(\"'+s+'\") =>', b);\n"+
  "// console.log('decode(',b,') =>', exports.decode(b));\n"+
  "\n"+
  "};\n"+
  "\n"+
  "__main=function(module) { var exports = module.exports;\n"+
  "\"use strict\";\n"+
  "var protocol = require('./protocol'),\n"+
  "      txt = protocol.text,\n"+
  "      bin = protocol.binary;\n"+
  "var Buf = require('./buf');\n"+
  "var utf8 = require('./utf8');\n"+
  "var EventEmitter = require('./EventEmitter');\n"+
  "var keepalive = require('./keepalive');\n"+
  "\n"+
  "var gotalk = exports;\n"+
  "\n"+
  "gotalk.protocol = protocol;\n"+
  "gotalk.Buf = Buf;\n"+
  "\n"+
  "function decodeJSON(v) {\n"+
  "  var value;\n"+
  "  try {\n"+
  "    value = JSON.parse(v);\n"+
  "  } catch (e) {\n"+
  "    // console.warn('failed to decode JSON \"'+(typeof v === 'string' ? v : v.toString())+'\":',e);\n"+
  "  }\n"+
  "  return value;\n"+
  "}\n"+
  "\n"+
  "\n"+
  "// ===============================================================================================\n"+
  "\n"+
  "function Sock(handlers) { return Object.create(Sock.prototype, {\n"+
  "  // Public properties\n"+
  "  handlers:      {value:handlers, enumerable:true},\n"+
  "  protocol:      {value: Buf ? protocol.binary : protocol.text, enumerable:true, writable:true},\n"+
  "  heartbeatInterval: {value: 20 * 1000, enumerable:true, writable:true},\n"+
  "\n"+
  "  // Internal\n"+
  "  ws:            {value:null, writable:true},\n"+
  "  keepalive:     {value:null, writable:true},\n"+
  "\n"+
  "  // Used for performing requests\n"+
  "  nextOpID:      {value:0, writable:true},\n"+
  "  nextStreamID:  {value:0, writable:true},\n"+
  "  pendingRes:    {value:{}, writable:true},\n"+
  "  hasPendingRes: {get:function(){ for (var k in this.pendingRes) { return true; } }},\n"+
  "\n"+
  "  // True if end() has been called while there were outstanding responses\n"+
  "  pendingClose:  {value:false, writable:true}\n"+
  "}); }\n"+
  "\n"+
  "Sock.prototype = EventEmitter.mixin(Sock.prototype);\n"+
  "exports.Sock = Sock;\n"+
  "\n"+
  "\n"+
  "var resetSock = function(s, causedByErr) {\n"+
  "  s.pendingClose = false;\n"+
  "\n"+
  "  if (s.ws) {\n"+
  "    s.ws.onmessage = null;\n"+
  "    s.ws.onerror = null;\n"+
  "    s.ws.onclose = null;\n"+
  "    s.ws = null;\n"+
  "  }\n"+
  "\n"+
  "  s.nextOpID = 0;\n"+
  "  if (s.hasPendingRes) {\n"+
  "    var err = causedByErr || new Error('connection closed');\n"+
  "    // TODO: return a RetryResult kind of error instead of just an error\n"+
  "    for (var k in s.pendingRes) {\n"+
  "      s.pendingRes[k](err);\n"+
  "    }\n"+
  "    s.pendingRes = {};\n"+
  "  }\n"+
  "};\n"+
  "\n"+
  "\n"+
  "var websocketCloseStatus = {\n"+
  "  1000: 'normal',\n"+
  "  1001: 'going away',\n"+
  "  1002: 'protocol error',\n"+
  "  1003: 'unsupported',\n"+
  "  // 1004 is currently unassigned\n"+
  "  1005: 'no status',\n"+
  "  1006: 'abnormal',\n"+
  "  1007: 'inconsistent',\n"+
  "  1008: 'invalid message',\n"+
  "  1009: 'too large',\n"+
  "};\n"+
  "\n"+
  "\n"+
  "// Adopt a web socket, which should be in an OPEN state\n"+
  "Sock.prototype.adoptWebSocket = function(ws) {\n"+
  "  var s = this;\n"+
  "  if (ws.readyState !== WebSocket.OPEN) {\n"+
  "    throw new Error('web socket readyState != OPEN');\n"+
  "  }\n"+
  "  ws.binaryType = 'arraybuffer';\n"+
  "  s.ws = ws;\n"+
  "  ws.onclose = function(ev) {\n"+
  "    var err = ws._gotalkCloseError;\n"+
  "    if (!err && ev.code !== 1000) {\n"+
  "      err = new Error('websocket closed: ' + (websocketCloseStatus[ev.code] || '#'+ev.code));\n"+
  "    }\n"+
  "    resetSock(s, err);\n"+
  "    s.emit('close', err);\n"+
  "  };\n"+
  "  ws.onmessage = function(ev) {\n"+
  "    if (!ws._bufferedMessages) ws._bufferedMessages = [];\n"+
  "    ws._bufferedMessages.push(ev.data);\n"+
  "  };\n"+
  "};\n"+
  "\n"+
  "\n"+
  "Sock.prototype.handshake = function () {\n"+
  "  this.ws.send(this.protocol.versionBuf);\n"+
  "};\n"+
  "\n"+
  "\n"+
  "Sock.prototype.end = function() {\n"+
  "  // Allow calling twice to \"force close\" even when there are pending responses\n"+
  "  var s = this;\n"+
  "  if (s.keepalive) {\n"+
  "    s.keepalive.disable();\n"+
  "    s.keepalive = null;\n"+
  "  }\n"+
  "  if (!s.pendingClose && s.hasPendingRes) {\n"+
  "    s.pendingClose = true;\n"+
  "  } else if (s.ws) {\n"+
  "    s.ws.close(1000);\n"+
  "  }\n"+
  "};\n"+
  "\n"+
  "\n"+
  "Sock.prototype.address = function() {\n"+
  "  var s = this;\n"+
  "  if (s.ws) {\n"+
  "    return s.ws.url;\n"+
  "  }\n"+
  "  return null;\n"+
  "};\n"+
  "\n"+
  "// ===============================================================================================\n"+
  "// Reading messages from a connection\n"+
  "\n"+
  "var ErrAbnormal = exports.ErrAbnormal = Error(\"unsupported protocol\");\n"+
  "ErrAbnormal.isGotalkProtocolError = true;\n"+
  "ErrAbnormal.code = protocol.ErrorAbnormal;\n"+
  "\n"+
  "var ErrUnsupported = exports.ErrUnsupported = Error(\"unsupported protocol\");\n"+
  "ErrUnsupported.isGotalkProtocolError = true;\n"+
  "ErrUnsupported.code = protocol.ErrorUnsupported;\n"+
  "\n"+
  "var ErrInvalidMsg = exports.ErrInvalidMsg = Error(\"invalid protocol message\");\n"+
  "ErrInvalidMsg.isGotalkProtocolError = true;\n"+
  "ErrInvalidMsg.code = protocol.ErrorInvalidMsg;\n"+
  "\n"+
  "var ErrTimeout = exports.ErrTimeout = Error(\"timeout\");\n"+
  "ErrTimeout.isGotalkProtocolError = true;\n"+
  "ErrTimeout.code = protocol.ErrorTimeout;\n"+
  "\n"+
  "\n"+
  "Sock.prototype.sendHeartbeat = function (load) {\n"+
  "  var s = this, buf = s.protocol.makeHeartbeatMsg(Math.round(load * protocol.HeartbeatMsgMaxLoad));\n"+
  "  try {\n"+
  "    s.ws.send(buf);\n"+
  "  } catch (err) {\n"+
  "    if (!this.ws || this.ws.readyState > WebSocket.OPEN) {\n"+
  "      err = new Error('socket is closed');\n"+
  "    }\n"+
  "    throw err;\n"+
  "  }\n"+
  "};\n"+
  "\n"+
  "\n"+
  "Sock.prototype.startSendingHeartbeats = function() {\n"+
  "  var s = this;\n"+
  "  if (s.heartbeatInterval < 10) {\n"+
  "    throw new Error(\"Sock.heartbeatInterval is too low\");\n"+
  "  }\n"+
  "  clearTimeout(s._sendHeartbeatsTimer);\n"+
  "  var send = function() {\n"+
  "    clearTimeout(s._sendHeartbeatsTimer);\n"+
  "    s.sendHeartbeat(0);\n"+
  "    s._sendHeartbeatsTimer = setTimeout(send, s.heartbeatInterval);\n"+
  "  };\n"+
  "  s._sendHeartbeatsTimer = setTimeout(send, 1);\n"+
  "};\n"+
  "\n"+
  "\n"+
  "Sock.prototype.stopSendingHeartbeats = function() {\n"+
  "  clearTimeout(s._sendHeartbeatsTimer);\n"+
  "};\n"+
  "\n"+
  "\n"+
  "Sock.prototype.startReading = function () {\n"+
  "  var s = this, ws = s.ws, msg;  // msg = current message\n"+
  "\n"+
  "  function readMsg(ev) {\n"+
  "    msg = typeof ev.data === 'string' ? txt.parseMsg(ev.data) : bin.parseMsg(Buf(ev.data));\n"+
  "    // console.log(\n"+
  "    //   'readMsg:',\n"+
  "    //   typeof ev.data === 'string' ? ev.data : Buf(ev.data).toString(),\n"+
  "    //   'msg:',\n"+
  "    //   msg\n"+
  "    // );\n"+
  "    if (msg.t === protocol.MsgTypeProtocolError) {\n"+
  "      var errcode = msg.size;\n"+
  "      if (errcode === protocol.ErrorAbnormal) {\n"+
  "        ws._gotalkCloseError = ErrAbnormal;\n"+
  "      } else if (errcode === protocol.ErrorUnsupported) {\n"+
  "        ws._gotalkCloseError = ErrUnsupported;\n"+
  "      } else if (errcode === protocol.ErrorTimeout) {\n"+
  "        ws._gotalkCloseError = ErrTimeout;\n"+
  "      } else {\n"+
  "        ws._gotalkCloseError = ErrInvalidMsg;\n"+
  "      }\n"+
  "      ws.close(4000 + errcode);\n"+
  "    } else if (msg.size !== 0 && msg.t !== protocol.MsgTypeHeartbeat) {\n"+
  "      ws.onmessage = readMsgPayload;\n"+
  "    } else {\n"+
  "      s.handleMsg(msg);\n"+
  "      msg = null;\n"+
  "    }\n"+
  "  }\n"+
  "\n"+
  "  function readMsgPayload(ev) {\n"+
  "    var b = ev.data;\n"+
  "    ws.onmessage = readMsg;\n"+
  "    s.handleMsg(msg, typeof b === 'string' ? b : Buf(b));\n"+
  "    msg = null;\n"+
  "  }\n"+
  "\n"+
  "  function readVersion(ev) {\n"+
  "    var peerVersion = typeof ev.data === 'string' ? txt.parseVersion(ev.data) :\n"+
  "                                                    bin.parseVersion(Buf(ev.data));\n"+
  "    if (peerVersion !== protocol.Version) {\n"+
  "      ws._gotalkCloseError = ErrUnsupported;\n"+
  "      s.closeError(protocol.ErrorUnsupported);\n"+
  "    } else {\n"+
  "      ws.onmessage = readMsg;\n"+
  "      if (s.heartbeatInterval > 0) {\n"+
  "        s.startSendingHeartbeats();\n"+
  "      }\n"+
  "    }\n"+
  "  }\n"+
  "\n"+
  "  // We begin by sending our version and reading the remote side's version\n"+
  "  ws.onmessage = readVersion;\n"+
  "\n"+
  "  // Any buffered messages?\n"+
  "  if (ws._bufferedMessages) {\n"+
  "    ws._bufferedMessages.forEach(function(data){ ws.onmessage({data:data}); });\n"+
  "    ws._bufferedMessages = null;\n"+
  "  }\n"+
  "};\n"+
  "\n"+
  "// ===============================================================================================\n"+
  "// Handling of incoming messages\n"+
  "\n"+
  "var msgHandlers = {};\n"+
  "\n"+
  "Sock.prototype.handleMsg = function(msg, payload) {\n"+
  "  // console.log('handleMsg:', String.fromCharCode(msg.t), msg, 'payload:', payload);\n"+
  "  var msgHandler = msgHandlers[msg.t];\n"+
  "  if (!msgHandler) {\n"+
  "    if (s.ws) {\n"+
  "      s.ws._gotalkCloseError = ErrInvalidMsg;\n"+
  "    }\n"+
  "    s.closeError(protocol.ErrorInvalidMsg);\n"+
  "  } else {\n"+
  "    msgHandler.call(this, msg, payload);\n"+
  "  }\n"+
  "};\n"+
  "\n"+
  "msgHandlers[protocol.MsgTypeSingleReq] = function (msg, payload) {\n"+
  "  var s = this, handler, result;\n"+
  "  handler = s.handlers.findRequestHandler(msg.name);\n"+
  "\n"+
  "  result = function (outbuf) {\n"+
  "    s.sendMsg(protocol.MsgTypeSingleRes, msg.id, null, 0, outbuf);\n"+
  "  };\n"+
  "  result.error = function (err) {\n"+
  "    var errstr = err.message || String(err);\n"+
  "    s.sendMsg(protocol.MsgTypeErrorRes, msg.id, null, 0, errstr);\n"+
  "  };\n"+
  "\n"+
  "  if (typeof handler !== 'function') {\n"+
  "    result.error('no such operation \"'+msg.name+'\"');\n"+
  "  } else {\n"+
  "    try {\n"+
  "      handler(payload, result, msg.name);\n"+
  "    } catch (err) {\n"+
  "      if (typeof console !== 'undefined') { console.error(err.stack || err); }\n"+
  "      result.error('internal error');\n"+
  "    }\n"+
  "  }\n"+
  "};\n"+
  "\n"+
  "function handleRes(msg, payload) {\n"+
  "  var id = typeof msg.id === 'string' ? msg.id : msg.id.toString();\n"+
  "  var s = this, callback = s.pendingRes[id];\n"+
  "  if (msg.t !== protocol.MsgTypeStreamRes || !payload || (payload.length || payload.size) === 0) {\n"+
  "    delete s.pendingRes[id];\n"+
  "    if (s.pendingClose && !s.hasPendingRes) {\n"+
  "      s.end();\n"+
  "    }\n"+
  "  }\n"+
  "  if (typeof callback !== 'function') {\n"+
  "    return; // ignore message\n"+
  "  }\n"+
  "  if (msg.t === protocol.MsgTypeErrorRes) {\n"+
  "    callback(new Error(String(payload)), null);\n"+
  "  } else {\n"+
  "    callback(null, payload);\n"+
  "  }\n"+
  "}\n"+
  "\n"+
  "msgHandlers[protocol.MsgTypeSingleRes] = handleRes;\n"+
  "msgHandlers[protocol.MsgTypeStreamRes] = handleRes;\n"+
  "msgHandlers[protocol.MsgTypeErrorRes] = handleRes;\n"+
  "\n"+
  "msgHandlers[protocol.MsgTypeNotification] = function (msg, payload) {\n"+
  "  var s = this, handler = s.handlers.findNotificationHandler(msg.name);\n"+
  "  if (handler) {\n"+
  "    handler(payload, msg.name);\n"+
  "  }\n"+
  "};\n"+
  "\n"+
  "msgHandlers[protocol.MsgTypeHeartbeat] = function (msg) {\n"+
  "  this.emit('heartbeat', {time:new Date(msg.size * 1000), load:msg.wait});\n"+
  "};\n"+
  "\n"+
  "// ===============================================================================================\n"+
  "// Sending messages\n"+
  "\n"+
  "\n"+
  "Sock.prototype.sendMsg = function(t, id, name, wait, payload) {\n"+
  "  var payloadSize = (payload && typeof payload === 'string' && this.protocol === protocol.binary) ?\n"+
  "    utf8.sizeOf(payload) :\n"+
  "    payload ? payload.length || payload.size :\n"+
  "    0;\n"+
  "  var s = this, buf = s.protocol.makeMsg(t, id, name, wait, payloadSize);\n"+
  "  // console.log('sendMsg(',t,id,name,payload,'): protocol.makeMsg =>',\n"+
  "  //   typeof buf === 'string' ? buf : buf.toString());\n"+
  "  try {\n"+
  "    s.ws.send(buf);\n"+
  "    if (payloadSize !== 0) {\n"+
  "      s.ws.send(payload);\n"+
  "    }\n"+
  "  } catch (err) {\n"+
  "    if (!this.ws || this.ws.readyState > WebSocket.OPEN) {\n"+
  "      err = new Error('socket is closed');\n"+
  "    }\n"+
  "    throw err;\n"+
  "  }\n"+
  "};\n"+
  "\n"+
  "\n"+
  "Sock.prototype.closeError = function(code) {\n"+
  "  var s = this, buf;\n"+
  "  if (s.ws) {\n"+
  "    try {\n"+
  "      s.ws.send(s.protocol.makeMsg(protocol.MsgTypeProtocolError, null, null, 0, code));\n"+
  "    } catch (e) {}\n"+
  "    s.ws.close(4000 + code);\n"+
  "  }\n"+
  "};\n"+
  "\n"+
  "var zeroes = '0000';\n"+
  "\n"+
  "// callback function(Error, outbuf)\n"+
  "Sock.prototype.bufferRequest = function(op, buf, callback) {\n"+
  "  var s = this, id = s.nextOpID++;\n"+
  "  if (s.nextOpID === 1679616) {\n"+
  "    // limit for base36 within 4 digits (36^4=1679616)\n"+
  "    s.nextOpID = 0;\n"+
  "  }\n"+
  "  id = id.toString(36);\n"+
  "  id = zeroes.substr(0, 4 - id.length) + id;\n"+
  "\n"+
  "  s.pendingRes[id] = callback;\n"+
  "  try {\n"+
  "    s.sendMsg(protocol.MsgTypeSingleReq, id, op, 0, buf);\n"+
  "  } catch (err) {\n"+
  "    delete s.pendingRes[id];\n"+
  "    callback(err);\n"+
  "  }\n"+
  "}\n"+
  "\n"+
  "\n"+
  "Sock.prototype.bufferNotify = function(name, buf) {\n"+
  "  s.sendMsg(protocol.MsgTypeNotification, null, name, 0, buf);\n"+
  "}\n"+
  "\n"+
  "\n"+
  "Sock.prototype.request = function(op, value, callback) {\n"+
  "  var buf;\n"+
  "  if (!callback) {\n"+
  "    // no value\n"+
  "    callback = value;\n"+
  "  } else {\n"+
  "    buf = JSON.stringify(value);\n"+
  "  }\n"+
  "  return this.bufferRequest(op, buf, function (err, buf) {\n"+
  "    var value = decodeJSON(buf);\n"+
  "    return callback(err, value);\n"+
  "  });\n"+
  "};\n"+
  "\n"+
  "\n"+
  "Sock.prototype.notify = function(op, value) {\n"+
  "  var buf = JSON.stringify(value);\n"+
  "  return this.bufferNotify(op, buf);\n"+
  "};\n"+
  "\n"+
  "\n"+
  "// ===============================================================================================\n"+
  "\n"+
  "// Represents a stream request.\n"+
  "// Response(s) arrive by the \"data\"(buf) event. When the response is complete, a \"end\"(error)\n"+
  "// event is emitted, where error is non-empty if the request failed.\n"+
  "var StreamRequest = function(s, op, id) {\n"+
  "  return Object.create(StreamRequest.prototype, {\n"+
  "    s:          {value:s},\n"+
  "    op:         {value:op, enumerable:true},\n"+
  "    id:         {value:id, enumerable:true},\n"+
  "    onresponse: {value:function(){}, enumerable:true, write:true}\n"+
  "  });\n"+
  "};\n"+
  "\n"+
  "EventEmitter.mixin(StreamRequest.prototype);\n"+
  "\n"+
  "StreamRequest.prototype.write = function (buf) {\n"+
  "  if (!this.ended) {\n"+
  "    if (!this.started) {\n"+
  "      this.started = true;\n"+
  "      this.s.sendMsg(protocol.MsgTypeStreamReq, this.id, this.op, 0, buf);\n"+
  "    } else {\n"+
  "      this.s.sendMsg(protocol.MsgTypeStreamReqPart, this.id, null, 0, buf);\n"+
  "    }\n"+
  "    if (!buf || buf.length === 0 || buf.size === 0) {\n"+
  "      this.ended = true;\n"+
  "    }\n"+
  "  }\n"+
  "};\n"+
  "\n"+
  "// Finalize the request\n"+
  "StreamRequest.prototype.end = function () {\n"+
  "  this.write(null);\n"+
  "};\n"+
  "\n"+
  "Sock.prototype.streamRequest = function(op) {\n"+
  "  var s = this, id = s.nextStreamID++;\n"+
  "  if (s.nextStreamID === 46656) {\n"+
  "    // limit for base36 within 3 digits (36^3=46656)\n"+
  "    s.nextStreamID = 0;\n"+
  "  }\n"+
  "  id = id.toString(36);\n"+
  "  id = '!' + zeroes.substr(0, 3 - id.length) + id;\n"+
  "\n"+
  "  var req = StreamRequest(s, op, id);\n"+
  "\n"+
  "  s.pendingRes[id] = function (err, buf) {\n"+
  "    if (err) {\n"+
  "      req.emit('end', err);\n"+
  "    } else if (!buf || buf.length === 0) {\n"+
  "      req.emit('end', null);\n"+
  "    } else {\n"+
  "      req.emit('data', buf);\n"+
  "    }\n"+
  "  };\n"+
  "\n"+
  "  return req;\n"+
  "};\n"+
  "\n"+
  "\n"+
  "// ===============================================================================================\n"+
  "\n"+
  "function Handlers() { return Object.create(Handlers.prototype, {\n"+
  "  reqHandlers:         {value:{}},\n"+
  "  reqFallbackHandler:  {value:null, writable:true},\n"+
  "  noteHandlers:        {value:{}},\n"+
  "  noteFallbackHandler: {value:null, writable:true}\n"+
  "}); }\n"+
  "exports.Handlers = Handlers;\n"+
  "\n"+
  "\n"+
  "Handlers.prototype.handleBufferRequest = function(op, handler) {\n"+
  "  if (!op) {\n"+
  "    this.reqFallbackHandler = handler;\n"+
  "  } else {\n"+
  "    this.reqHandlers[op] = handler;\n"+
  "  }\n"+
  "};\n"+
  "\n"+
  "Handlers.prototype.handleRequest = function(op, handler) {\n"+
  "  return this.handleBufferRequest(op, function (buf, result, op) {\n"+
  "    var resultWrapper = function(value) {\n"+
  "      return result(JSON.stringify(value));\n"+
  "    };\n"+
  "    resultWrapper.error = result.error;\n"+
  "    var value = decodeJSON(buf);\n"+
  "    handler(value, resultWrapper, op);\n"+
  "  });\n"+
  "};\n"+
  "\n"+
  "Handlers.prototype.handleBufferNotification = function(name, handler) {\n"+
  "  if (!name) {\n"+
  "    this.noteFallbackHandler = handler;\n"+
  "  } else {\n"+
  "    this.noteHandlers[name] = handler;\n"+
  "  }\n"+
  "};\n"+
  "\n"+
  "Handlers.prototype.handleNotification = function(name, handler) {\n"+
  "  this.handleBufferNotification(name, function (buf, name) {\n"+
  "    handler(decodeJSON(buf), name);\n"+
  "  });\n"+
  "};\n"+
  "\n"+
  "Handlers.prototype.findRequestHandler = function(op) {\n"+
  "  var handler = this.reqHandlers[op];\n"+
  "  return handler || this.reqFallbackHandler;\n"+
  "};\n"+
  "\n"+
  "Handlers.prototype.findNotificationHandler = function(name) {\n"+
  "  var handler = this.noteHandlers[name];\n"+
  "  return handler || this.noteFallbackHandler;\n"+
  "};\n"+
  "\n"+
  "// ===============================================================================================\n"+
  "\n"+
  "function openWebSocket(s, addr, callback) {\n"+
  "  var ws;\n"+
  "  try {\n"+
  "    ws = new WebSocket(addr);\n"+
  "    ws.binaryType = 'arraybuffer';\n"+
  "    ws.onclose = function (ev) {\n"+
  "      var err = new Error('connection failed');\n"+
  "      if (callback) callback(err);\n"+
  "    };\n"+
  "    ws.onopen = function(ev) {\n"+
  "      ws.onerror = undefined;\n"+
  "      s.adoptWebSocket(ws);\n"+
  "      s.handshake();\n"+
  "      if (callback) callback(null, s);\n"+
  "      s.emit('open');\n"+
  "      s.startReading();\n"+
  "    };\n"+
  "    ws.onmessage = function(ev) {\n"+
  "      if (!ws._bufferedMessages) ws._bufferedMessages = [];\n"+
  "      ws._bufferedMessages.push(ev.data);\n"+
  "    };\n"+
  "  } catch (err) {\n"+
  "    if (callback) callback(err);\n"+
  "    s.emit('close', err);\n"+
  "  }\n"+
  "}\n"+
  "\n"+
  "\n"+
  "// gotalk.defaultResponderAddress is defined if the responder has announced a default address\n"+
  "// to which connect to.\n"+
  "if (window.gotalkResponderAt !== undefined) {\n"+
  "  var at = window.gotalkResponderAt;\n"+
  "  delete window.gotalkResponderAt;\n"+
  "  if (at && at.ws) {\n"+
  "    gotalk.defaultResponderAddress = 'ws://' + document.location.host + at.ws;\n"+
  "  }\n"+
  "}\n"+
  "\n"+
  "\n"+
  "Sock.prototype.open = function(addr, callback) {\n"+
  "  var s = this;\n"+
  "  if (!callback && typeof addr == 'function') {\n"+
  "    callback = addr;\n"+
  "    addr = null;\n"+
  "  }\n"+
  "\n"+
  "  if (!addr) {\n"+
  "    if (!gotalk.defaultResponderAddress) {\n"+
  "      throw new Error('address not specified (responder has not announced any default address)')\n"+
  "    }\n"+
  "    addr = gotalk.defaultResponderAddress;\n"+
  "  }\n"+
  "\n"+
  "  if (addr.substr(0,3) === 'ws:') {\n"+
  "    openWebSocket(s, addr, callback);\n"+
  "  } else {\n"+
  "    throw new Error('unsupported address');\n"+
  "  }\n"+
  "  return s;\n"+
  "};\n"+
  "\n"+
  "\n"+
  "// Open a connection to a gotalk responder.\n"+
  "// \n"+
  "// open(addr string[, onConnect(Error, Sock)]) -> Sock\n"+
  "//   Connect to gotalk responder at `addr`\n"+
  "//\n"+
  "// open([onConnect(Error, Sock)]) -> Sock\n"+
  "//   Connect to default gotalk responder.\n"+
  "//   Throws an error if `gotalk.defaultResponderAddress` isn't defined.\n"+
  "//\n"+
  "gotalk.open = function(addr, onConnect) {\n"+
  "  var s = Sock(gotalk.defaultHandlers);\n"+
  "  s.open(addr, onConnect);\n"+
  "  return s;\n"+
  "};\n"+
  "\n"+
  "\n"+
  "// If `addr` is not provided, `gotalk.defaultResponderAddress` is used instead.\n"+
  "Sock.prototype.openKeepAlive = function(addr) {\n"+
  "  var s = this;\n"+
  "  if (s.keepalive) {\n"+
  "    s.keepalive.disable();\n"+
  "  }\n"+
  "  s.keepalive = keepalive(s, addr);\n"+
  "  s.keepalive.enable();\n"+
  "  return s;\n"+
  "};\n"+
  "\n"+
  "\n"+
  "// Returns a new Sock with a persistent connection to a gotalk responder.\n"+
  "// The Connection is automatically kept alive (by reconnecting) until Sock.end() is called.\n"+
  "// If `addr` is not provided, `gotalk.defaultResponderAddress` is used instead.\n"+
  "gotalk.connection = function(addr) {\n"+
  "  var s = Sock(gotalk.defaultHandlers);\n"+
  "  s.openKeepAlive(addr);\n"+
  "  return s;\n"+
  "};\n"+
  "\n"+
  "\n"+
  "gotalk.defaultHandlers = Handlers();\n"+
  "\n"+
  "gotalk.handleBufferRequest = function(op, handler) {\n"+
  "  return gotalk.defaultHandlers.handleBufferRequest(op, handler);\n"+
  "};\n"+
  "\n"+
  "gotalk.handle = function(op, handler) {\n"+
  "  return gotalk.defaultHandlers.handleRequest(op, handler);\n"+
  "};\n"+
  "\n"+
  "gotalk.handleBufferNotification = function (name, handler) {\n"+
  "  return gotalk.defaultHandlers.handleBufferNotification(name, handler);\n"+
  "};\n"+
  "\n"+
  "gotalk.handleNotification = function (name, handler) {\n"+
  "  return gotalk.defaultHandlers.handleNotification(name, handler);\n"+
  "};\n"+
  "\n"+
  "// -----------------------------------------------------------------------------------------------\n"+
  "\n"+
  "\n"+
  "\n"+
  "};\n"+
  "\n"+
  "\n"+
  "__mainapi = {exports:{}};\n"+
  "__main(__mainapi);\n"+
  "\n"+
  "global.gotalk = __mainapi.exports;\n"+
  "})(typeof window !== \"undefined\" ? window : this);\n"+
  ""
