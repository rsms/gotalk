function noop(){}

exports.global = (
  typeof global != 'undefined' ? global :
  typeof self != 'undefined' ? self :
  typeof window != 'undefined' ? window :
  this
)

exports.console = (
  typeof console != 'undefined' ? console :
  {log:noop,warn:noop,error:noop}
)

exports.document = (
  typeof document != 'undefined' ? document :
  null
)
