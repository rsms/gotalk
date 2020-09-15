import * as utf8 from './utf8'

export var Buf = (function() {
  if (typeof Uint8Array == 'undefined') {
    return null
  }

  // Buf(Buf) -> Buf
  // Buf(size int) -> Buf
  // Buf(ArrayBuffer) -> Buf
  return function Buf(v) {
    return v instanceof Uint8Array ? v :
      new Uint8Array(
        v instanceof ArrayBuffer ? v :
        new ArrayBuffer(v)
      );
  };

})()
