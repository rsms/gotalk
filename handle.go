package gotalk
import (
  "reflect"
  "errors"
  "sync"
  "encoding/json"
)

type Handlers struct {

  // Look up a handler for operation `op`. Returns `nil` if not found.
  // FindRequestHandler(op string) interface{}
  // FindNotificationHandler(name string) BufferNoteHandler

  reqHandlersMu       sync.RWMutex
  reqHandlers         reqHandlerMap
  reqFallbackHandler  interface{}
  notesMu             sync.RWMutex
  noteHandlers        noteHandlerMap
  noteFallbackHandler BufferNoteHandler
}

func NewHandlers() *Handlers {
  return &Handlers{reqHandlers:make(reqHandlerMap), noteHandlers:make(noteHandlerMap)}
}

type BufferReqHandler   func(s *Sock, op string, payload []byte) ([]byte, error)
type BufferNoteHandler  func(s *Sock, name string, payload []byte)

// EOS when <-rch==nil
type StreamReqHandler   func(s *Sock, name string, rch chan []byte, write StreamWriter) error
type StreamWriter       func([]byte) error

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
  DefaultHandlers.HandleRequest(op, fn)
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

type reqHandlerMap  map[string]interface{}
type noteHandlerMap map[string]BufferNoteHandler

func (h *Handlers) setRequestHandler(op string, fn interface{}) {
  h.reqHandlersMu.Lock()
  defer h.reqHandlersMu.Unlock()
  if len(op) == 0 {
    h.reqFallbackHandler = fn
  } else {
    h.reqHandlers[op] = fn
  }
}

// See Handle()
func (h *Handlers) HandleRequest(op string, fn interface{}) {
  h.HandleBufferRequest(op, wrapFuncReqHandler(fn))
}

// See HandleBufferRequest()
func (h *Handlers) HandleBufferRequest(op string, fn BufferReqHandler) {
  h.setRequestHandler(op, fn)
}

// See HandleStreamRequest()
func (h *Handlers) HandleStreamRequest(op string, fn StreamReqHandler) {
  h.setRequestHandler(op, fn)
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

// Look up a handler for operation `op`. Returns `nil` if not found.
func (h *Handlers) FindRequestHandler(op string) interface{} {
  h.reqHandlersMu.RLock()
  defer h.reqHandlersMu.RUnlock()
  if handler := h.reqHandlers[op]; handler != nil {
    return handler
  }
  return h.reqFallbackHandler
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


func typeIsSockPtr(t reflect.Type) bool {
  return t.Kind() == reflect.Ptr && t.Elem().AssignableTo(kSockType)
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

  if fnt.NumIn() == 3 {
    // `func(*Sock, string, interface{}) (interface{}, error)`
    if typeIsSockPtr(fnt.In(0)) == false {
      panic(errMsgBadHandler)
    }
    if fnt.In(1).Kind() != reflect.String {
      panic(errMsgBadHandler)
    }
    paramsType := fnt.In(2)

    return BufferReqHandler(func (s *Sock, op string, inbuf []byte) ([]byte, error) {
      paramsVal, err := decodeParams(paramsType, inbuf)
      if err != nil {
        return nil, err
      }
      r := fnv.Call([]reflect.Value{reflect.ValueOf(s), reflect.ValueOf(op), paramsVal.Elem()})
      return decodeResult(r)
    })

  } else if fnt.NumIn() == 2 {
    // Signature: `func(*Sock, interface{})(interface{}, error)`
    if typeIsSockPtr(fnt.In(0)) == false {
      panic(errMsgBadHandler)
    }
    paramsType := fnt.In(1)

    return BufferReqHandler(func (s *Sock, _ string, inbuf []byte) ([]byte, error) {
      paramsVal, err := decodeParams(paramsType, inbuf)
      if err != nil {
        return nil, err
      }
      r := fnv.Call([]reflect.Value{reflect.ValueOf(s), paramsVal.Elem()})
      return decodeResult(r)
    })

  } else if fnt.NumIn() == 1 {
    if typeIsSockPtr(fnt.In(0)) {
      if fnt.NumOut() == 2 {
        // Signature: `func(*Sock)(interface{}, error)`
        return BufferReqHandler(func (s *Sock, _ string, _ []byte) ([]byte, error) {
          r := fnv.Call([]reflect.Value{reflect.ValueOf(s)})
          return decodeResult(r)
        })
      } else {
        // Signature: `func(*Sock)error`
        f, ok := fn.(func(*Sock)error)
        if ok == false {
          panic(errMsgBadHandler)
        }
        return BufferReqHandler(func (s *Sock, _ string, _ []byte) ([]byte, error) {
          return nil, f(s)
        })
      }

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
    if typeIsSockPtr(fnt.In(0)) == false || fnt.In(1).Kind() != reflect.String {
      panic(errMsgBadHandler)
    }
    paramsType := fnt.In(2)
    return BufferNoteHandler(
      func (s *Sock, name string, inbuf []byte) {
        paramsVal, _ := decodeParams(paramsType, inbuf)
        fnv.Call([]reflect.Value{reflect.ValueOf(s), reflect.ValueOf(name), paramsVal.Elem()})
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

