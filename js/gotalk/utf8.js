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
