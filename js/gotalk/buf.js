"use strict";
var Buf;

if (typeof Uint8Array !== 'undefined') {

var utf8 = require('./utf8');

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
