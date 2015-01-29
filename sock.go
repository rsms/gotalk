package gotalk
import (
  "encoding/json"
  "errors"
  "io"
  "log"
  "net"
  "sync"
  "fmt"
)


type Fault struct {
  // True if this is a responder fault (protocol: "RetryResult") and the requestor is
  // allowed to retry the request. False is the fault is a hard error (protocol: "ErrorResult").
  CanRetry bool

  // When CanRetry is true, this is the number of milliseconds the requestor must wait before
  // retrying
  Wait    int

  // Description of the fault
  Message []byte
}

// error interface
func (f *Fault) Error() string {
  return string(f.Message)
}

func newErrorFault(m []byte) *Fault {
  return &Fault{false, 0, m}
}
func newRetryFault(wait int, m []byte) *Fault {
  return &Fault{true, wait, m}
}

// ----------------------------------------------------------------------------------------------

type pendingResMap  map[string]chan interface{}
type pendingReqMap  map[string]chan []byte

type Sock struct {
  // Handlers associated with this socket
  Handlers *Handlers

  // Associate some application-specific data with this socket
  UserData interface{}

  // Enable streaming requests and set the limit for how many streaming requests this socket
  // can handle at the same time. Setting this to `0` disables streaming requests alltogether
  // (the default) while setting this to a large number might be cause for security concerns
  // as a malicious peer could send many "start stream" messages, but never sending
  // any "end stream" messages, slowly exhausting memory.
  StreamReqLimit int

  // A function to be called when the socket closes
  CloseHandler func(*Sock)

  // -------------------------------------------------------------------------
  // Used by connected sockets
  wmu            sync.Mutex          // guards writes on conn
  conn           io.ReadWriteCloser  // non-nil after successful call to Connect or accept

  // Used for sending requests:
  nextOpID       uint
  pendingRes     pendingResMap
  pendingResMu   sync.RWMutex

  // Used for handling streaming requests:
  pendingReq     pendingReqMap
  pendingReqMu   sync.RWMutex
}


func NewSock(h *Handlers) *Sock {
  return &Sock{Handlers:h}
}


// Creates two sockets which are connected to eachother without any resource limits.
// If `handlers` is nil, DefaultHandlers are used. If `limits` is nil, DefaultLimits are used.
func Pipe(handlers *Handlers, limits Limits) (*Sock, *Sock, error) {
  if handlers == nil {
    handlers = DefaultHandlers
  }
  c1, c2 := net.Pipe()
  s1 := NewSock(handlers)
  s2 := NewSock(handlers)
  s1.Adopt(c1)
  s2.Adopt(c2)
  // Note: We deliberately ignore performing a handshake
  if limits == nil {
    limits = DefaultLimits
  }
  go s1.Read(limits)
  go s2.Read(limits)
  return s1, s2, nil
}


// Connect to a server via `how` at `addr`. Unless there's an error, the returned socket is
// already reading in a different goroutine and is ready to be used.
// If `limits` is nil, DefaultLimits are used.
func Connect(how, addr string, limits Limits) (*Sock, error) {
  c, err := net.Dial(how, addr)
  if err != nil {
    return nil, err
  }
  s := NewSock(DefaultHandlers)
  s.Adopt(c)
  if err := s.Handshake(); err != nil {
    return nil, err
  }
  if limits == nil {
    limits = DefaultLimits
  }
  go s.Read(limits)
  return s, nil
}


// Adopt an I/O stream, which should already be in a "connected" state. After calling this,
// you need to call Handshake and Read to perform the protocol handshake and read messages.
func (s *Sock) Adopt(c io.ReadWriteCloser) {
  if s.conn != nil {
    panic("already adopted")
  }
  s.conn = c
}

// ----------------------------------------------------------------------------------------------

func (s *Sock) getResChan(id string) chan interface{} {
  s.pendingResMu.RLock()
  defer s.pendingResMu.RUnlock()
  if s.pendingRes == nil {
    return nil
  }
  return s.pendingRes[id]
}


func (s *Sock) allocResChan() (string, chan interface{}) {
  ch := make(chan interface{})

  s.pendingResMu.Lock()
  defer s.pendingResMu.Unlock()

  id := string(makeFixnumBuf(3, uint64(s.nextOpID), 36))
  s.nextOpID++
  if s.nextOpID == 46656 {
    // limit for base36 within 3 digits (36^3=46656)
    s.nextOpID = 0
  }

  if s.pendingRes == nil {
    s.pendingRes = make(pendingResMap)
  }
  s.pendingRes[id] = ch

  return id, ch
}


func (s *Sock) deallocResChan(id string) {
  s.pendingResMu.Lock()
  defer s.pendingResMu.Unlock()
  delete(s.pendingRes, id)
}

// ----------------------------------------------------------------------------------------------

func (s *Sock) getReqChan(id string) chan []byte {
  s.pendingReqMu.RLock()
  defer s.pendingReqMu.RUnlock()
  if s.pendingReq == nil {
    return nil
  }
  return s.pendingReq[id]
}


func (s *Sock) allocReqChan(id string) chan []byte {
  ch := make(chan []byte, 1)

  s.pendingReqMu.Lock()
  defer s.pendingReqMu.Unlock()

  if s.pendingReq == nil {
    s.pendingReq = make(pendingReqMap)
  }
  if s.pendingReq[id] != nil {
    panic("identical request ID in two different requests")
  }
  s.pendingReq[id] = ch
  return ch
}


func (s *Sock) deallocReqChan(id string) {
  s.pendingReqMu.Lock()
  defer s.pendingReqMu.Unlock()
  delete(s.pendingReq, id)
}

// ----------------------------------------------------------------------------------------------

func (s *Sock) writeMsg(t MsgType, id, op string, wait int, buf []byte) error {
  if s.conn == nil {
    panic("not connected")
  }
  s.wmu.Lock()
  defer s.wmu.Unlock()
  if _, err := s.conn.Write(MakeMsg(t, id, op, wait, len(buf))); err != nil {
    return err
  }
  if len(buf) != 0 {
    _, err := s.conn.Write(buf)
    return err
  }
  return nil
}


const reqHandlerTypeBuf = reqHandlerType(iota)
type reqHandlerType int


// Send a single-buffer request
func (s *Sock) BufferRequest(op string, buf []byte) ([]byte, *Fault) {
  id, ch := s.allocResChan()
  defer s.deallocResChan(id)

  //fmt.Printf("BufferRequest: writeMsg(%v, %v, %v)\n", id, op, buf)

  if err := s.writeMsg(MsgTypeSingleReq, id, op, 0, buf); err != nil {
    return nil, newErrorFault([]byte(err.Error()))
  }

  // Wait for response to be read in readLoop
  ch <- reqHandlerTypeBuf
  // Note: Don't call any code in between here that can panic, as that could cause readLoop to
  // deadlock.
  resval := <-ch  // response buffer

  if resbuf, ok := resval.(resbuffer); ok {
    if resbuf.t == MsgTypeSingleRes {
      return resbuf.b, nil
    } else if resbuf.t == MsgTypeErrorRes {
      return nil, newErrorFault(resbuf.b)
    } else if resbuf.t == MsgTypeRetryRes {
      return nil, newRetryFault(resbuf.wait, resbuf.b)
    }
    // Note: We require the response to be buffered and not streaming
    return resbuf.b, newErrorFault([]byte("unexpected message "+string(byte(resbuf.t))))
  }
  return nil, nil
}


// Send a single-value request where the input and output values are JSON-encoded
func (s *Sock) Request(op string, in interface{}, out interface{}) *Fault {
  inbuf, err := json.Marshal(in)
  if err != nil {
    return newErrorFault([]byte(err.Error()))
  }
  outbuf, fault := s.BufferRequest(op, inbuf)
  if fault != nil {
    return fault
  }
  if err := json.Unmarshal(outbuf, out); err != nil {
    return newErrorFault([]byte(err.Error()))
  }
  return nil
}


// Send a multi-buffer streaming request
func (s *Sock) StreamRequest(op string) *StreamRequest {
  return &StreamRequest{sock:s, op:op}
}


// Send a single-buffer notification
func (s *Sock) BufferNotify(t string, buf []byte) error {
  return s.writeMsg(MsgTypeNotification, "", t, 0, buf)
}

// Send a single-value request where the value is JSON-encoded
func (s *Sock) Notify(t string, v interface{}) error {
  if buf, err := json.Marshal(v); err != nil {
    return err
  } else {
    return s.BufferNotify(t, buf)
  }
}

// ----------------------------------------------------------------------------------------------

type StreamRequest struct {
  sock    *Sock
  op      string
  id      string
  started bool // request started?
  ended   bool // response ended?
  ch      chan interface{}
}

func (r *StreamRequest) finalize() {
  if r.id != "" {
    r.sock.deallocResChan(r.id)
    r.id = ""
  }
}

func (r *StreamRequest) Write(b []byte) error {
  if r.started == false {
    r.started = true
    r.id, r.ch = r.sock.allocResChan()
    if err := r.sock.writeMsg(MsgTypeStreamReq, r.id, r.op, 0, b); err != nil {
      r.finalize()
      return err
    }
  } else {
    if err := r.sock.writeMsg(MsgTypeStreamReqPart, r.id, "", 0, b); err != nil {
      r.finalize()
      return err
    }
  }
  return nil
}

func (r *StreamRequest) End() error {
  err := r.sock.writeMsg(MsgTypeStreamReqPart, r.id, "", 0, nil)
  if err != nil {
    r.finalize()
  }
  return err
}

func (r *StreamRequest) Read() ([]byte, error) {
  if r.ended == true {
    return nil, nil
  }

  // Wait for result chunk to be read in readLoop
  r.ch <- reqHandlerTypeBuf
  resval := <- r.ch

  // Interpret resbuf
  if resbuf, ok := resval.(resbuffer); ok {
    if resbuf.t == MsgTypeErrorRes {
      r.ended = true
      return nil, errors.New(string(resbuf.b))
    } else if resbuf.t == MsgTypeSingleRes {
      r.ended = true
    }
    return resbuf.b, nil
  }
  return nil, nil
}

// ----------------------------------------------------------------------------------------------

func (s *Sock) readDiscard(readz int) error {
  if readz != 0 {
    // todo: is there a better way to read data w/o copying it into a buffer?
    err := readn(s.conn, make([]byte, readz))
    return err
  }
  return nil
}


func (s *Sock) respondError(readz int, id, msg string) error {
  if err := s.readDiscard(readz); err != nil {
    return err
  }
  return s.writeMsg(MsgTypeErrorRes, id, "", 0, []byte(msg))
}


func (s *Sock) respondRetry(readz int, id string, wait int, msg string) error {
  if err := s.readDiscard(readz); err != nil {
    return err
  }
  return s.writeMsg(MsgTypeRetryRes, id, "", wait, []byte(msg))
}


func (s *Sock) respondOK(id string, b []byte) error {
  return s.writeMsg(MsgTypeSingleRes, id, "", 0, b)
}


func (s *Sock) readBufferReq(limits Limits, id, op string, size int) error {
  if limits.incBufferReq() == false {
    return s.respondRetry(size, id, limitWaitBufferReq(), "request rate limit")
  }

  handler := s.Handlers.FindBufferRequestHandler(op)
  if handler == nil {
    err := s.respondError(size, id, "unknown operation \""+op+"\"")
    limits.decBufferReq()
    return err
  }

  inbuf := make([]byte, size)
  if err := readn(s.conn, inbuf); err != nil {
    limits.decBufferReq()
    return err
  }

  // Dispatch handler
  go func() {
    defer func() {
      if r := recover(); r != nil {
        if s.conn != nil {
          if err := s.respondError(0, id, fmt.Sprint(r)); err != nil {
            log.Println(err)
            s.Close()
          }
        }
      }
      limits.decBufferReq()
    }()
    outbuf, err := handler(s, op, inbuf)
    if err != nil {
      log.Println(err)
      if err := s.respondError(0, id, err.Error()); err != nil {
        log.Println(err)
        s.Close()
      }
    } else {
      if err := s.respondOK(id, outbuf); err != nil {
        log.Println(err)
        s.Close()
      }
    }
  }()

  return nil
}


func (s *Sock) readStreamReq(limits Limits, id, op string, size int) error {
  if limits.incStreamReq() == false {
    if limits.streamReqEnabled() {
      return s.respondRetry(size, id, limitWaitStreamReq(), "request rate limit")
    } else {
      return s.respondError(size, id, "stream requests not supported")
    }
  }

  handler := s.Handlers.FindStreamRequestHandler(op)
  if handler == nil {
    err := s.respondError(size, id, "unknown operation \""+op+"\"")
    limits.decStreamReq()
    return err
  }

  // Read first buff
  inbuf := make([]byte, size)
  if err := readn(s.conn, inbuf); err != nil {
    limits.decStreamReq()
    return err
  }

  // Create read chan
  rch := s.allocReqChan(id)
  rch <- inbuf

  // Create result writer
  wroteEOS := false
  writer := func (b []byte) error {
    if len(b) == 0 {
      wroteEOS = true
    }
    return s.writeMsg(MsgTypeStreamRes, id, "", 0, b)
  }

  // Dispatch handler
  go func () {
    // TODO: recover?
    if err := handler(s, op, rch, writer); err != nil {
      s.deallocReqChan(id)
      if err := s.respondError(0, id, err.Error()); err != nil {
        log.Println(err)
        s.Close()
      }
    }
    if wroteEOS == false {
      // automatically writing EOS unless it was written by handler
      if err := s.writeMsg(MsgTypeStreamRes, id, "", 0, nil); err != nil {
        log.Println(err)
        s.Close()
      }
    }
    limits.decStreamReq()
  }()

  return nil
}


func (s *Sock) readStreamReqPart(limits Limits, id string, size int) error {
  rch := s.getReqChan(id)
  if rch == nil {
    return errors.New("illegal message")  // There was no "start stream" message
  }

  var b []byte = nil

  if size != 0 {
    b = make([]byte, size)
    if err := readn(s.conn, b); err != nil {
      limits.decStreamReq()
      return err
    }
  }

  rch <- b
  return nil
}

// -----------------------------------------------------------------------------------------------

type resbuffer struct {
  t    MsgType
  wait int
  b    []byte
}

func (s *Sock) readRes(t MsgType, id string, wait, size int) error {
  ch := s.getResChan(id)

  if ch == nil {
    // Unexpected response: discard and ignore
    return s.readDiscard(size)
  }

  handlerTv := <-ch

  if handlerType, ok := handlerTv.(reqHandlerType); ok {
    switch (handlerType) {
    case reqHandlerTypeBuf:
      // Request handler expects buffer
      var buf []byte
      if size != 0 {
        buf = make([]byte, size)
        if err := readn(s.conn, buf); err != nil {
          return err
        }
      }
      ch <- resbuffer{t, wait, buf}

    default:
      panic("unexpected req handler type")
    }
  } else {
    panic("unexpected req handler type")
  }

  return nil
}


func (s *Sock) readNotification(name string, size int) error {
  handler := s.Handlers.FindNotificationHandler(name)

  if handler == nil {
    // read any payload and ignore notification
    return s.readDiscard(size)
  }

  // Read any payload
  var buf []byte
  if size != 0 {
    buf = make([]byte, size)
    if err := readn(s.conn, buf); err != nil {
      return err
    }
  }

  handler(s, name, buf)
  return nil
}


// Before reading any messages over a socket, handshake must happen. This function will block
// until the handshake either succeeds or fails.
func (s *Sock) Handshake() error {
  // Write, read and compare version
  if _, err := WriteVersion(s.conn); err != nil {
    s.Close()
    return err
  }
  if _, err := ReadVersion(s.conn); err != nil {
    s.Close()
    return err
  }
  return nil
}


// After completing a succesful handshake, call this function to read messages received to this
// socket. Does not return until the socket is closed.
func (s *Sock) Read(limits Limits) error {
  for {

    // debug: read a chunk and print it
    // b := make([]byte, 128)
    // if _, err := s.conn.Read(b); err != nil {
    //   log.Println(err)
    //   break
    // }
    // fmt.Printf("Read: %v\n", string(b))
    // continue

    // Read next message
    t, id, name, wait, size, err := ReadMsg(s.conn)
    if err != nil {
      s.Close()
      return err
    }

    //fmt.Printf("readLoop: msg: t=%c  id=%v  name=%v  size=%v\n", byte(t), id, name, size)

    switch t {
      case MsgTypeSingleReq:
        err = s.readBufferReq(limits, id, name, int(size))

      case MsgTypeStreamReq:
        err = s.readStreamReq(limits, id, name, int(size))

      case MsgTypeStreamReqPart:
        err = s.readStreamReqPart(limits, id, int(size))

      case MsgTypeSingleRes, MsgTypeStreamRes, MsgTypeErrorRes, MsgTypeRetryRes:
        err = s.readRes(t, id, int(wait), int(size))

      case MsgTypeNotification:
        err = s.readNotification(name, int(size))

      default:
        return errors.New("unexpected protocol message type")
    }

    if err != nil {
      s.Close()
      return err
    }
  }

  // never reached
  return nil
}


// Address of this socket
func (s *Sock) Addr() string {
  if s.conn != nil {
    if netconn, ok := s.conn.(net.Conn); ok {
      return netconn.RemoteAddr().String()
    }
  }
  return ""
}


// Close this socket
func (s *Sock) Close() error {
  if s.conn != nil {
    err := s.conn.Close()
    s.conn = nil
    if s.CloseHandler != nil {
      s.CloseHandler(s)
    }
    return err
  }
  return nil
}
