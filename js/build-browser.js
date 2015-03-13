"use strict";
var fs = require('fs');
var crypto = require('crypto');
var subprocess = require('child_process');

var srcDir = __dirname + '/gotalk';
var browserFile = __dirname + '/browser/browser.js';

function buildAll() {
  var sourceFiles = fs.readdirSync(srcDir).filter(function (filename) {
    return !!filename.match(/\.js$/);
  });

  // Make sure indexjs is the last source file
  sourceFiles.sort(function (a, b) {
    return a === 'index.js' ? 1 :
           b === 'index.js' ? -1 :
           b < a;
  });

  var moduleNames = sourceFiles.filter(function(filename) {
    return filename !== 'index.js';
  }).map(function (filename) {
    return filename.replace(/\.[^\.]+$/, '');
  });

  var sources = sourceFiles.map(function (filename) {
    var moduleName = filename.replace(/\.[^\.]+$/, '');
    var fileContents = fs.readFileSync(srcDir + '/' + filename, 'utf8');
    if (filename === 'index.js') {
      return '__main=function(module) { var exports = module.exports;\n' +
        fileContents +
        '\n};\n' +
        ''
    } else {
      return '__mod["'+moduleName+'"]=function(module) { var exports = module.exports;\n' +
        fileContents +
        '\n};\n';
    }
  });

  var vars = {
    MODULES_MAP: moduleNames.map(function (name) {
      return JSON.stringify(name) + ':{exports:{}}';
    }).join(','),

    MODULES_SRC: sources.join('\n'),
  };

  var browserSrc = fs.readFileSync(browserFile, 'utf8');
  var source = browserSrc.replace(/\$\{([a-zA-Z_]+[a-zA-Z0-9_]*)\}/g, function (m0, name) {
    return vars[name] || '';
  });

  var tmpsrcfile = __dirname + '/.gotalk-debug.js';
  fs.writeFileSync(tmpsrcfile, source);

  console.log('write js/gotalk.js, js/gotalk.js.map');
  var p = subprocess.spawnSync("uglifyjs", [
    '.gotalk-debug.js',
    '--output', 'gotalk.js',
    '--mangle',
    // '--define', 'window',
    // '--reserved', 'window',
    '--lint',
    '--compress',
    '--source-map', __dirname + '/gotalk.js.map',
    // '--source-map-root', './',
    '--source-map-url', 'gotalk.js.map',
    '--source-map-include-sources',
    '--screw-ie8',
  ], {
    cwd: __dirname
  });
  if (p.error) {
    throw p.error;
  }

  buildGoFile();
}


function buildGoFile() {
  // var srcbuf = new Buffer(source, 'utf8');
  var srcbuf = fs.readFileSync(__dirname + '/gotalk.js');
  var srcmapbuf = fs.readFileSync(__dirname + '/gotalk.js.map');
  // console.log(srcbuf.toString('hex'))

  var srcsha1 = crypto.createHash('sha1');
  srcsha1.update(srcbuf);
  var sha1buf = srcsha1.digest();

  var goSource =
    'package gotalkjs\n'+
    'const BrowserLibSizeString = "'+srcbuf.length.toString(10)+'"\n'+
    // 'const BrowserLibSHA1Raw    = '+bufToByteStr(sha1buf, '"')+'\n'+
    'const BrowserLibSHA1Base64 = "'+sha1buf.toString('base64')+'"\n'+
    'const BrowserLibETag = "\\\"'+sha1buf.toString('base64')+'\\\""\n'+
    'const BrowserLibString = '+bufToByteStr(srcbuf, '"')+'\n'+
    'const BrowserLibSourceMapString = '+bufToByteStr(srcmapbuf, '"')+'\n'+
    //'var BrowserLibBytes        = [...]byte{\n  '
    '';
  // var hex2v = srcbuf.toString('hex');
  // for (var c, n = 0, i = 0, L = srcbuf.length; i !== L; ++i) {
  //   c = srcbuf[i];
  //   if (n++ === 20) {
  //     n = 0;
  //     goSource += "\n  ";
  //   }
  //   if (c === 0x0a) {
  //     goSource += "'\\n',";
  //   } else if (c === 0x27) {
  //     goSource += "'\\'',";
  //   } else if (c === 0x5c) {
  //     goSource += "'\\\\',";
  //   } else if (c >= 0x20 && c < 0x7f) {
  //     goSource += "'"+String.fromCharCode(c)+"',"
  //   } else {
  //     goSource += c.toString(10) + ',';
  //   }
  // }
  // goSource += '\n}\n';
  console.log('write js/gotalk.js.go');
  fs.writeFileSync(__dirname + '/gotalk.js.go', goSource);
}


var map = {
  0x09: "\\t",
  0x0a: "\\n",
  0x0d: "\\r",
  0x5c: "\\\\",
};

function strEscByte(c, enclosedByByte, enclosedByChar) {
  var r = map[c];
  if (r !== undefined) {
    return r;
  }
  if (c === enclosedByByte) {
    return "\\" + enclosedByChar;
  } else if (c >= 0x20 && c < 0x7f) {
    return String.fromCharCode(c);
  } else {
    return (c <= 0xf ? '\\x0' : '\\x')+c.toString(16);
  }
}

function bufToByteStr(buf, enclosedByChar, breakUpLines) {
  var enclosedByByte = enclosedByChar.charCodeAt(0);
  var s = (breakUpLines ? enclosedByChar+enclosedByChar+'+\n  ' : '') + enclosedByChar;
  for (var c, i = 0, L = buf.length; i !== L; ++i) {
    c = buf[i];
    if (c === 0x0a && breakUpLines) {
      s += '\\n'+enclosedByChar+'+\n  '+enclosedByChar
    } else {
      s += strEscByte(c, enclosedByByte, enclosedByChar);
    }
  }
  return s + enclosedByChar;
}

// function bufToMultilineByteStr(buf) {
//   var s = '`', b = '`'.charCodeAt(0);
//   for (var c, i = 0, L = buf.length; i !== L; ++i) {
//     c = buf[i];
//     if (c === 0x0a) {
//       s += '\n';
//     } else if (c === b) {
//       throw new Error('backtick inside multiline string is not supported by Go')
//     } else {
//       s += strEscByte(c);
//     }
//   }
//   return s + '`';
// }


buildAll();

console.log('Watching for changes...');
fs.watch(srcDir, buildAll);
fs.watch(browserFile, buildAll);
