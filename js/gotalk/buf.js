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

}

module.exports = Buf;
