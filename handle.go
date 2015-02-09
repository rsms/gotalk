package gotalk
import (
  "reflect"
  "errors"
  "sync"
  "encoding/json"
  "io"
)

type Handlers struct {
  bufReqHandlersMu          sync.RWMutex
  bufReqHandlers            bufReqHandlerMap
  bufReqFallbackHandler     BufferReqHandler

  streamReqHandlersMu       sync.RWMutex
  streamReqHandlers         streamReqHandlerMap
  streamReqFallbackHandler  StreamReqHandler

  notesMu                   sync.RWMutex
  noteHandlers              noteHandlerMap
  noteFallbackHandler       BufferNoteHandler
}

func NewHandlers() *Handlers {
  return &Handlers{
    bufReqHandlers:    make(bufReqHandlerMap),
    streamReqHandlers: make(streamReqHandlerMap),
    noteHandlers:      make(noteHandlerMap)}
}

// If a handler panics, it's assumed that the effect of the panic was isolated to the active
// request. Panic is recovered, a stack trace is logged, and connection is closed.
type BufferReqHandler   func(s *Sock, op string, payload []byte) ([]byte, error)
type BufferNoteHandler  func(s *Sock, name string, payload []byte)

// EOS when <-rch==nil
type StreamReqHandler   func(s *Sock, name string, rch chan []byte, out io.WriteCloser) error

var DefaultHandlers = NewHandlers()

// Handle operation with automatic JSON encoding of values.
//
// `fn` must conform to one of the following signatures:
//   func(*Sock, string, interface{}) (interface{}, error) -- takes socket, op and parameters
//   func(*Sock, interface{}) (interface{}, error)         -- takes socket and parameters
//   func(interface{}) (interface{}, error)                -- takes parameters, but no socket
//   func(*Sock) (interface{}, error)                      -- takes no parameters
//   func() (interface{},error)                            -- takes no socket or parameters
//
// Where optionally the `interface{}` return value can be omitted, i.e:
//   func(*Sock, string, interface{}) error
//   func(*Sock, interface{}) error
//   func(interface{}) error
//   func(*Sock) error
//   func() error
//
// If `op` is empty, handle all requests which doesn't have a specific handler registered.
func Handle(op string, fn interface{}) {
  DefaultHandlers.Handle(op, fn)
}

// Handle operation with raw input and output buffers. If `op` is empty, handle
// all requests which doesn't have a specific handler registered.
func HandleBufferRequest(op string, fn BufferReqHandler) {
  DefaultHandlers.HandleBufferRequest(op, fn)
}

// Handle operation by reading and writing directly from/to the underlying stream.
// If `op` is empty, handle all requests which doesn't have a specific handler registered.
func HandleStreamRequest(op string, fn StreamReqHandler) {
  DefaultHandlers.HandleStreamRequest(op, fn)
}

// Handle notifications of a certain name with automatic JSON encoding of values.
//
// `fn` must conform to one of the following signatures:
//   func(s *Sock, name string, v interface{}) -- takes socket, name and parameters
//   func(name string, v interface{})          -- takes name and parameters, but no socket
//   func(v interface{})                       -- takes only parameters
//
// If `name` is empty, handle all notifications which doesn't have a specific handler
// registered.
func HandleNotification(name string, fn interface{}) {
  DefaultHandlers.HandleNotification(name, fn)
}

// Handle notifications of a certain name with raw input buffers. If `name` is empty, handle
// all notifications which doesn't have a specific handler registered.
func HandleBufferNotification(name string, fn BufferNoteHandler) {
  DefaultHandlers.HandleBufferNotification(name, fn)
}

// -------------------------------------------------------------------------------------

type bufReqHandlerMap    map[string]BufferReqHandler
type streamReqHandlerMap map[string]StreamReqHandler
type noteHandlerMap      map[string]BufferNoteHandler


// See Handle()
func (h *Handlers) Handle(op string, fn interface{}) {
  h.HandleBufferRequest(op, wrapFuncReqHandler(fn))
}

// See HandleBufferRequest()
func (h *Handlers) HandleBufferRequest(op string, fn BufferReqHandler) {
  h.bufReqHandlersMu.Lock()
  defer h.bufReqHandlersMu.Unlock()
  if len(op) == 0 {
    h.bufReqFallbackHandler = fn
  } else {
    h.bufReqHandlers[op] = fn
  }
}

// See HandleStreamRequest()
func (h *Handlers) HandleStreamRequest(op string, fn StreamReqHandler) {
  h.streamReqHandlersMu.Lock()
  defer h.streamReqHandlersMu.Unlock()
  if len(op) == 0 {
    h.streamReqFallbackHandler = fn
  } else {
    h.streamReqHandlers[op] = fn
  }
}

// See HandleNotification()
func (h *Handlers) HandleNotification(name string, fn interface{}) {
  h.HandleBufferNotification(name, wrapFuncNotHandler(fn))
}

// See HandleBufferNotification()
func (h *Handlers) HandleBufferNotification(name string, fn BufferNoteHandler) {
  h.notesMu.Lock()
  defer h.notesMu.Unlock()
  if len(name) == 0 {
    h.noteFallbackHandler = fn
  } else {
    h.noteHandlers[name] = fn
  }
}

// Look up a single-buffer handler for operation `op`. Returns `nil` if not found.
func (h *Handlers) FindBufferRequestHandler(op string) BufferReqHandler {
  h.bufReqHandlersMu.RLock()
  defer h.bufReqHandlersMu.RUnlock()
  if handler := h.bufReqHandlers[op]; handler != nil {
    return handler
  }
  return h.bufReqFallbackHandler
}

// Look up a stream handler for operation `op`. Returns `nil` if not found.
func (h *Handlers) FindStreamRequestHandler(op string) StreamReqHandler {
  h.streamReqHandlersMu.RLock()
  defer h.streamReqHandlersMu.RUnlock()
  if handler := h.streamReqHandlers[op]; handler != nil {
    return handler
  }
  return h.streamReqFallbackHandler
}

// Look up a handler for notification `name`. Returns `nil` if not found.
func (h *Handlers) FindNotificationHandler(name string) BufferNoteHandler {
  h.notesMu.RLock()
  defer h.notesMu.RUnlock()
  if handler := h.noteHandlers[name]; handler != nil {
    return handler
  }
  return h.noteFallbackHandler
}

// -------------------------------------------------------------------------------------

var (
  errMsgBadHandler = "invalid handler func signature (see gotalk.Handlers)"
  errUnexpectedParamType = errors.New("unexpected parameter type")

  kErrorType = reflect.TypeOf(new(error)).Elem()
  kSockType = reflect.TypeOf(new(Sock)).Elem()
)


func valToErr(r reflect.Value) error {
  v := r.Interface()
  if err, ok := v.(error); ok {
    return err
  } else if s, ok := v.(string); ok {
    return errors.New(s)
  }
  return errors.New("error")  // fixme
}


func decodeResult(r []reflect.Value) ([]byte, error) {
  if len(r) == 2 {
    if r[1].IsNil() {
      return json.Marshal(r[0].Interface())
    } else {
      return nil, valToErr(r[1])
    }
  } else if r[0].IsNil() {
    return nil, nil
  } else {
    return nil, valToErr(r[0])
  }
}


func decodeParams(paramsType reflect.Type, inbuf []byte) (*reflect.Value, error) {
  paramsVal := reflect.New(paramsType)
  params := paramsVal.Interface()
  if err := json.Unmarshal(inbuf, &params); err != nil {
    return &paramsVal, errUnexpectedParamType
  }
  return &paramsVal, nil
}


type sockPtrToValueFunc func(*Sock)reflect.Value

func typeIsSockPtr(t reflect.Type) (ok bool, sockPtrToValue sockPtrToValueFunc) {
  if t.Kind() == reflect.Ptr {
    if t.Elem().AssignableTo(kSockType) {
      return true, func(s *Sock) reflect.Value { return reflect.ValueOf(s) }
    } else if t.Elem().ConvertibleTo(kSockType) {
      return true, func(s *Sock) reflect.Value { return reflect.ValueOf(s).Convert(t) }
    }
  }
  return false, nil
}


func wrapFuncReqHandler(fn interface{}) BufferReqHandler {
  // `fn` must conform to one of the following signatures:
  //   `func(*Sock, interface{})(interface{}, error)` -- takes socket and parameters
  //   `func(interface{})(interface{}, error)`        -- takes parameters, but no socket
  //   `func(*Sock)(interface{}, error)`              -- takes no parameters
  //   `func()(interface{},error)`                    -- takes no socket or parameters
  fnv := reflect.ValueOf(fn)
  fnt := fnv.Type()

  if fnt.Kind() != reflect.Func {
    panic("handler must be a function")
  }

  if fnt.NumIn() > 3 || fnt.NumOut() < 1 || fnt.NumOut() > 2 ||
     fnt.Out(fnt.NumOut() - 1).Implements(kErrorType) == false {
    panic(errMsgBadHandler)
  }

  var in0IsSockPtr bool
  var sockPtrToValue sockPtrToValueFunc
  if fnt.NumIn() > 0 {
    in0IsSockPtr, sockPtrToValue = typeIsSockPtr(fnt.In(0))
    if in0IsSockPtr == false && fnt.NumIn() > 1 {
      panic(errMsgBadHandler)
    }
  }

  if fnt.NumIn() == 3 {
    // `func(*Sock, string, interface{}) (interface{}, error)`
    if fnt.In(1).Kind() != reflect.String {
      panic(errMsgBadHandler)
    }
    paramsType := fnt.In(2)

    return BufferReqHandler(func (s *Sock, op string, inbuf []byte) ([]byte, error) {
      paramsVal, err := decodeParams(paramsType, inbuf)
      if err != nil {
        return nil, err
      }
      r := fnv.Call([]reflect.Value{sockPtrToValue(s), reflect.ValueOf(op), paramsVal.Elem()})
      return decodeResult(r)
    })

  } else if fnt.NumIn() == 2 {
    // Signature: `func(*Sock, interface{})(interface{}, error)`
    paramsType := fnt.In(1)

    return BufferReqHandler(func (s *Sock, _ string, inbuf []byte) ([]byte, error) {
      paramsVal, err := decodeParams(paramsType, inbuf)
      if err != nil {
        return nil, err
      }
      r := fnv.Call([]reflect.Value{sockPtrToValue(s), paramsVal.Elem()})
      return decodeResult(r)
    })

  } else if fnt.NumIn() == 1 {
    if in0IsSockPtr {
      // Signature: `func(*Sock)(interface{}, error)`
      return BufferReqHandler(func (s *Sock, _ string, _ []byte) ([]byte, error) {
        r := fnv.Call([]reflect.Value{sockPtrToValue(s)})
        return decodeResult(r)
      })
    } else {
      // Signature: `func(interface{})(interface{}, error)`
      paramsType := fnt.In(0)
      return BufferReqHandler(func (_ *Sock, _ string, inbuf []byte) ([]byte, error) {
        paramsVal, err := decodeParams(paramsType, inbuf)
        if err != nil {
          return nil, err
        }
        r := fnv.Call([]reflect.Value{paramsVal.Elem()})
        return decodeResult(r)
      })
    }

  } else {
    if fnt.NumOut() == 2 {
      // Signature: `func()(interface{},error)`
      return BufferReqHandler(func (_ *Sock, _ string, _ []byte) ([]byte, error) {
        r := fnv.Call(nil)
        return decodeResult(r)
      })
    } else {
      // Signature: `func()error`
      f, ok := fn.(func()error)
      if ok == false {
        panic(errMsgBadHandler)
      }
      return BufferReqHandler(func (_ *Sock, _ string, _ []byte) ([]byte, error) {
        return nil, f()
      })
    }
  }

}


func wrapFuncNotHandler(fn interface{}) BufferNoteHandler {
  // `fn` must conform to one of the following signatures:
  //   `func(*Sock, string, interface{})` -- takes socket, name and parameters
  //   `func(string, interface{})`        -- takes name and parameters, but no socket
  //   `func(interface{})`                -- takes only parameters
  fnv := reflect.ValueOf(fn)
  fnt := fnv.Type()

  if fnt.Kind() != reflect.Func {
    panic("handler must be a function")
  }

  if fnt.NumIn() > 3 || fnt.NumOut() > 0 {
    panic(errMsgBadHandler)
  }

  if fnt.NumIn() == 3 {
    // Signature: `func(*Sock, string, interface{})`
    in0IsSockPtr, sockPtrToValue := typeIsSockPtr(fnt.In(0))
    if in0IsSockPtr == false || fnt.In(1).Kind() != reflect.String {
      panic(errMsgBadHandler)
    }
    paramsType := fnt.In(2)
    return BufferNoteHandler(
      func (s *Sock, name string, inbuf []byte) {
        paramsVal, _ := decodeParams(paramsType, inbuf)
        fnv.Call([]reflect.Value{sockPtrToValue(s), reflect.ValueOf(name), paramsVal.Elem()})
      })
  } else if fnt.NumIn() == 2 {
    // Signature: `func(string, interface{})`
    if fnt.In(0).Kind() != reflect.String {
      panic(errMsgBadHandler)
    }
    paramsType := fnt.In(1)
    return BufferNoteHandler(
      func (_ *Sock, name string, inbuf []byte) {
        paramsVal, _ := decodeParams(paramsType, inbuf)
        fnv.Call([]reflect.Value{reflect.ValueOf(name), paramsVal.Elem()})
      })
  } else {
    // Signature: `func(interface{})`
    paramsType := fnt.In(0)
    return BufferNoteHandler(
      func (_ *Sock, _ string, inbuf []byte) {
        paramsVal, _ := decodeParams(paramsType, inbuf)
        fnv.Call([]reflect.Value{paramsVal.Elem()})
      })
  }
}

