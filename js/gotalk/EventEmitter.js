
function EventEmitter() {}
module.exports = EventEmitter;

EventEmitter.prototype.addListener = function (type, listener) {
  if (typeof listener !== 'function') throw TypeError('listener must be a function');
  if (!this.__events) {
    Object.defineProperty(this, '__events', {value:{}, enumerable:false, writable:true});
    this.__events[type] = [listener];
    return 1;
  }
  var listeners = this.__events[type];
  if (listeners === undefined) {
    this.__events[type] = [listener];
    return 1;
  }
  listeners.push(listener);
  return listeners.length;
};

EventEmitter.prototype.on = EventEmitter.prototype.addListener;

EventEmitter.prototype.once = function (type, listener) {
  var fired = false;
  function trigger_event_once() {
    this.removeListener(type, trigger_event_once);
    if (!fired) {
      fired = true;
      listener.apply(this, arguments);
    }
  }
  return this.on(type, trigger_event_once);
};

EventEmitter.prototype.removeListener = function (type, listener) {
  var p, listeners = this.__events ? this.__events[type] : undefined;
  if (listeners !== undefined) {
    while ((p = listeners.indexOf(listener)) !== -1) {
      delete listeners[p];
    }
    if (listeners.length === 0) {
      delete this.__events[type];
    }
    return listeners.length;
  }
  return 0;
};

EventEmitter.prototype.removeAllListeners = function (type) {
  if (this.__events) {
    if (type) {
      delete this.__events[type];
    } else {
      delete this.__events;
    }
  }
};

EventEmitter.prototype.listeners = function (type) {
  return type ? (this.__events ? this.__events[type] : undefined) : this.__events;
};

EventEmitter.prototype.emit = function (type) {
  var listeners = this.__events ? this.__events[type] : undefined;
  if (listeners === undefined) {
    return false;
  }
  var i = 0, L = listeners.length, args = Array.prototype.slice.call(arguments,1);
  for (; i !== L; ++i) {
    listeners[i].apply(this, args);
  }
  return true;
};

EventEmitter.mixin = function mixin(obj) {
  var proto = obj;
  while (proto) {
    if (proto.__proto__ === Object.prototype) {
      proto.__proto__ = EventEmitter.prototype;
      return obj;
    }
    proto = proto.__proto__;
  }
  return obj;
};

