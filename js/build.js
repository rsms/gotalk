#!/usr/bin/env node
const { build, sha1, file } = require("estrella")
const pkg = require("./package.json")
const fs = require("fs").promises

process.chdir(__dirname)

build({
  entry: "gotalk/index.js",
  outfile: "gotalk.js",
  bundle: true,
  sourcemap: true,
  globalName: "gotalk",
  target: "es5",
  // format: "cjs",
  define: { VERSION: pkg.version },
  onEnd: postProcess,
})


async function postProcess(config, diagnostics) {
  let jsbuf = await file.read(config.outfile)

  // find sourceMappingURL
  let sourceMappingURLBuf = Buffer.from("//# sourceMappingURL")
  let i = jsbuf.indexOf(sourceMappingURLBuf)
  if (i == -1) { throw new Error("could not find '//# sourceMappingURL' in gotalk.js") }

  // js without sourcemap url
  let jsbuf1 = jsbuf.subarray(0, i)

  // js with nodejs compat code added
  let jsbuf2 = Buffer.concat([
    jsbuf.subarray(0, i),
    Buffer.from(`if(typeof module!="undefined")module.exports=gotalk;\n`),
    jsbuf.subarray(i),
  ])

  // go source with embedded js
  let goSource = `
  package gotalk

  const JSLibSHA1Base64 = "${sha1(jsbuf1, 'base64')}"
  const JSLibString = ${bytesToGoString(jsbuf1)}
  `.trim().replace(/\n +/g,"\n") + "\n"

  await Promise.all([
    file.write(config.outfile, jsbuf2),
    file.write("../jslib.go", goSource, {log:true}),
  ])
}


function bytesToGoString(buf) {
  const charmap = {
    0x09: "\\t",
    0x0a: "\\n",
    0x0d: "\\r",
    0x5c: "\\\\",
  }
  function strEscByte(c) {
    let r = charmap[c]
    return (
      r !== undefined       ? r :     // special representation, e.g. "\n"
      c == 0x22             ? '\\"' : // '"'
      c >= 0x20 && c < 0x7f ? String.fromCharCode(c) : // verbatim
                              (c <= 0xf ? '\\x0' : '\\x') + c.toString(16)  // "\xHH"
    )
  }
  let s = '"';
  for (let c, i = 0, L = buf.length; i !== L; ++i) {
    c = buf[i];
    s += strEscByte(c)
  }
  return s + '"';
}
