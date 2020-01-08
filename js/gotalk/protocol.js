"use strict";
var Buf = require('./buf');
var utf8 = require('./utf8');

// Version of this protocol
exports.Version = 1;

// Message types
var MsgTypeSingleReq     = exports.MsgTypeSingleReq =     0x72 // 'r'.charCodeAt(0)
  , MsgTypeStreamReq     = exports.MsgTypeStreamReq =     0x73 // 's'.charCodeAt(0)
  , MsgTypeStreamReqPart = exports.MsgTypeStreamReqPart = 0x70 // 'p'.charCodeAt(0)
  , MsgTypeSingleRes     = exports.MsgTypeSingleRes =     0x52 // 'R'.charCodeAt(0)
  , MsgTypeStreamRes     = exports.MsgTypeStreamRes =     0x53 // 'S'.charCodeAt(0)
  , MsgTypeErrorRes      = exports.MsgTypeErrorRes =      0x45 // 'E'.charCodeAt(0)
  , MsgTypeRetryRes      = exports.MsgTypeRetryRes =      0x65 // 'e'.charCodeAt(0)
  , MsgTypeNotification  = exports.MsgTypeNotification =  0x6E // 'n'.charCodeAt(0)
  , MsgTypeHeartbeat     = exports.MsgTypeHeartbeat =     0x68 // 'h'.charCodeAt(0)
  , MsgTypeProtocolError = exports.MsgTypeProtocolError = 0x66 // 'f'.charCodeAt(0)
;

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

function parseIntBuf(b, radix) {
  return parseInt(String.fromCharCode(...b), radix)
}


exports.binary = {

  makeFixnum: makeBufFixnum,

  versionBuf: makeBufFixnum(exports.Version, 2),

  parseVersion: function (b) {
    return parseIntBuf(b, 16);
  },

  // Parses a byte buffer containing a message (not including payload data.)
  // If t is MsgTypeHeartbeat, wait==load, size==time.
  // -> {t:string, id:Buf, name:string, wait:int size:int} | null
  parseMsg: function (b) {
    var t, id, name, namez, wait = 0, size = 0, z;
    // Example:
    // R000A00000006
    // R             = type response
    //  0000         = id   10
    //      00000006 = size 6

    t = b[0];
    z = 1;

    if (t === MsgTypeHeartbeat) {
      wait = parseIntBuf(b.subarray(z, z + 4), 16);
      z += 4;
    } else if (t !== MsgTypeNotification && t !== MsgTypeProtocolError) {
      id = b.subarray(z, z + 4);
      z += 4;
    }

    if (t == MsgTypeSingleReq || t == MsgTypeStreamReq || t == MsgTypeNotification) {
      namez = parseIntBuf(b.subarray(z, z + 3), 16);
      z += 3;
      name = utf8.decode(b.subarray(z, z + namez));
      z += namez;
    } else if (t === MsgTypeRetryRes) {
      wait = parseIntBuf(b.subarray(z, z + 8), 16);
      z += 8
    }

    size = parseIntBuf(b.subarray(z, z + 8), 16);

    return {t:t, id:id, name:name, wait:wait, size:size};
  },

  // Create a buf representing a message (w/o any payload)
  makeMsg: function (t, id, name, wait, size) {
    var b, nameb, z = id ? 13 : 9;

    // if there's a name, encode as utf8 and increase buffer size
    if (name && name.length !== 0) {
      nameb = utf8.encode(name);
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

    if (nameb) {
      copyBufFixnum(b, z, nameb.length, 3);
      z += 3;
      b.set(nameb, z);
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

