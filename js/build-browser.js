"use strict";
var fs = require('fs');

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
      return '(function(module, exports) {\n' +
        fileContents +
        '\n})(__main, __main.exports);\n' +
        ''
    } else {
      return '(function(module) { var exports = module.exports;\n' +
        fileContents +
        '\n})(modules["'+moduleName+'"]);\n';
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

  fs.writeFileSync(__dirname + '/gotalk.js', source);
  console.log('Built gotalk.js ('+source.length+' bytes)');
}


buildAll();

console.log('Watching for changes...');
fs.watch(srcDir, buildAll);
fs.watch(browserFile, buildAll);
