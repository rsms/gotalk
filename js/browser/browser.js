(function(global){
"use strict";

var __mod = {}, __api = {}, __main, __mainapi;
var evalDepth = 0;

var require = function(name) {
  var m, id = name.replace(/^.\//, "");
  m = __api[id];
  //console.log('require', {name:name, id:id, exports:m});
  if (!m) {
    var prefix = ''; for (var i = 0; i < evalDepth; i++) {
      prefix += '. ';
    }
    var f = __mod[id];
    if (f && evalDepth < 100) {
      __mod[id] = null;
      __api[id] = {exports:{}};
      ++evalDepth;
      f(__api[id]);
      --evalDepth;
      __api[id] = __api[id].exports;
    }
    m = __api[id];
  }
  return m;
};

${MODULES_SRC}

__mainapi = {exports:{}};
__main(__mainapi);

global.gotalk = __mainapi.exports;
})(this);
