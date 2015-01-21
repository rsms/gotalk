(typeof window !== "undefined" ? window : this).gotalk = (function(global){
"use strict";

var modules = {${MODULES_MAP}}, __main = {exports:{}}, module;
var require = function(name) {
  return modules[name.replace(/^.\//, "")].exports;
};

${MODULES_SRC}

var gotalk = __main.exports;
// ==================== Browser-additions ====================
//

return gotalk;
})();
