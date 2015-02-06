"use strict";
var EventEmitter = require('./EventEmitter');
var m;

if (typeof global !== 'undefined' && global.addEventListener) {
  m = Object.create(EventEmitter.prototype, {
    available: {value:true, enumerable:true},
    onLine:    {value:true, enumerable:true, writable:true}
  });

  if (typeof navigator !== 'undefined') {
    m.onLine = navigator.onLine;
  }

  global.addEventListener("offline", function (ev) {
    m.onLine = false;
    m.emit('offline');
  });

  global.addEventListener("online", function (ev) {
    m.onLine = true;
    m.emit('online');
  });

} else {
  m = {available:false, onLine:true};
}

module.exports = m;
