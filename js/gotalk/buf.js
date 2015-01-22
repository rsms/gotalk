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
