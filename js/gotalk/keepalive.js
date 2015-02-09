"use strict";
// Stay connected by automatically reconnecting w/ exponential back-off.

var netAccess = require('./netaccess');
var protocol = require('./protocol');

// `s` must conform to interface { connect(addr string, cb function(Error)) }
// Returns an object {
//   isConnected bool  // true if currently connected
//   isEnabled bool    // true if enabled
//   enable()          // enables staying connected
//   disable()         // disables trying to stay connected
// }
var keepalive = function(s, addr, minReconnectDelay, maxReconnectDelay) {
  if (!minReconnectDelay) {
    minReconnectDelay = 500;
  } else if (minReconnectDelay < 100) {
    minReconnectDelay = 100;
  }

  if (!maxReconnectDelay || maxReconnectDelay < minReconnectDelay) {
    maxReconnectDelay = 5000;
  }

  var ctx, open, retry, delay = 0, openTimer, opentime;

  ctx = {
    isEnabled: false,
    isConnected: false,
    enable: function() {
      if (!ctx.enabled) {
        ctx.enabled = true;
        delay = 0;
        if (!ctx.isConnected) {
          open();
        }
      }
    },
    disable: function() {
      if (ctx.enabled) {
        clearTimeout(openTimer);
        ctx.enabled = false;
        delay = 0;
      }
    }
  };

  open = function() {
    clearTimeout(openTimer);
    s.open(addr, function(err) {
      opentime = new Date;
      if (err) {
        retry(err);
      } else {
        delay = 0;
        ctx.isConnected = true;
        s.once('close', retry);
      }
    });
  };

  retry = function(err) {
    clearTimeout(openTimer);
    ctx.isConnected = false;
    if (!ctx.enabled) {
      return;
    }
    if (netAccess.available && !netAccess.onLine && 
        !(typeof document !== 'undefined' &&
          document.location &&
          document.location.hostname !== 'localhost' &&
          document.location.hostname !== '127.0.0.1' &&
          document.location.hostname !== '[::1]') )
    {
      netAccess.once('online', retry);
      delay = 0;
      return;
    }
    if (err) {
      if (err.isGotalkProtocolError) {
        if (err.code === protocol.ErrorTimeout) {
          delay = 0;
        } else {
          // We shouldn't retry with the same version of our gotalk library.
          // However, the only sensible thing to do in this case is to let the user code react to
          // the error passed to the close event (e.g. to show a "can't talk to server" UI), and
          // retry in maxReconnectDelay.
          // User code can choose to call `disable()` on its keepalive object in this case.
          delay = maxReconnectDelay;
        }
      } else {
        // increase back off in case of an error
        delay = delay ? Math.min(maxReconnectDelay, delay * 2) : minReconnectDelay;
      }
    } else {
      delay = Math.max(0, minReconnectDelay - ((new Date) - opentime));
    }
    openTimer = setTimeout(open, delay);
  };

  return ctx;
};

module.exports = keepalive;
