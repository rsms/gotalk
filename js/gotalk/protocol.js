"use strict";
var Buf = require('./buf');
var utf8 = require('./utf8');

// Version of this protocol
exports.Version = 0;

// Message types
var MsgTypeSingleReq     = exports.MsgTypeSingleReq =     'r'.charCodeAt(0),
    MsgTypeStreamReq     = exports.MsgTypeStreamReq =     's'.charCodeAt(0),
    MsgTypeStreamReqPart = exports.MsgTypeStreamReqPart = 'p'.charCodeAt(0),
    MsgTypeSingleRes     = exports.MsgTypeSingleRes =     'R'.charCodeAt(0),
    MsgTypeStreamRes     = exports.MsgTypeStreamRes =     'S'.charCodeAt(0),
    MsgTypeErrorRes      = exports.MsgTypeErrorRes =      'E'.charCodeAt(0),
    MsgTypeNotification  = exports.MsgTypeNotification =  'n'.charCodeAt(0);

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
  // -> {t:string, id:Buf, name:string, size:string} | null
  parseMsg: function (b) {
    var t, id, name, namez, size = 0, z;

    t = b[0];
    z = 1;

    if (t !== MsgTypeNotification) {
      id = b.slice(z, z + 3);
      z += 3;
    }

    if (t == MsgTypeSingleReq || t == MsgTypeStreamReq || t == MsgTypeNotification) {
      namez = parseInt(b.slice(z, z + 3), 16);
      z += 3;
      name = b.slice(z, z+namez).toString();
      z += namez;
    }

    size = parseInt(b.slice(z, z + 8), 16);

    return {t:t, id:id, name:name, size:size};
  },

  // Create a text string representing a message (w/o any payload.)
  makeMsg: function (t, id, name, size) {
    var b, z = id ? 12 : 9, nameb;

    if (name && name.length !== 0) {
      nameb = Buf.fromString(name);
      z += 3 + nameb.length;
    }

    b = Buf(z);

    b[0] = t;
    z = 1;

    if (id && id.length !== 0) {
      if (typeof id === 'string') {
        b[1] = id.charCodeAt(0);
        b[2] = id.charCodeAt(1);
        b[3] = id.charCodeAt(2);
      } else {
        b[1] = id[0];
        b[2] = id[1];
        b[3] = id[2];
      }
      z += 3;
    }

    if (name && name.length !== 0) {
      nameb = Buf.fromString(name);
      copyBufFixnum(b, z, nameb.length, 3);
      z += 3;
      nameb.copy(b, z);
      z += nameb.length;
    }

    copyBufFixnum(b, z, size, 8);

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
  // -> {t:string, id:string, name:string, size:string} | null
  parseMsg: function (s) {
    // "r001004echo00000005" => ('r', "001", "echo", 5)
    // "R00100000005"        => ('R', "001", "", 5)
    var t, id, name, size = 0, z;

    t = s.charCodeAt(0);
    z = 1;

    if (t !== MsgTypeNotification) {
      id = s.substr(z, 3);
      z += 3;
    }

    if (t == MsgTypeSingleReq || t == MsgTypeStreamReq || t == MsgTypeNotification) {
      name = s.substring(z + 3, s.length - 8);
    }

    size = parseInt(s.substr(s.length - 8), 16);

    return {t:t, id:id, name:name, size:size};
  },


  // Create a text string representing a message (w/o any payload.)
  makeMsg: function (t, id, name, size) {
    var b = String.fromCharCode(t);

    if (id && id.length !== 0) {
      b += id;
    }

    if (name && name.length !== 0) {
      b += makeStrFixnum(utf8.sizeOf(name), 3);
      b += name;
    }

    b += makeStrFixnum(size, 8);

    return b;
  }

}; // exports.text

