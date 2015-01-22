package gotalk

import (
  "encoding/json"
  "errors"
  "io"
  "log"
  "net"
  "os"
  "os/signal"
  "sync"
  "syscall"
)

type Sock interface {
  // Adopt an I/O stream, which should already be in a "connected" state. After calling this,
  // you need to call Handshake and Read to perform the protocol handshake and read messages.
  Adopt(io.ReadWriteCloser)

  // Before reading any messages over a socket, handshake must happen. This function will block
  // until the handshake either succeeds or fails.
  Handshake() error

  // After completing a succesful handshake, call this function to read messages received to this
  // socket. Does not return until the socket is closed.
  Read() error

  // Adopt a listener, which should already be in a "listening" state.
  AdoptListener(net.Listener)

  // Accept connections. Blocks until closed or an error occurs. SockHandler is called for
  // each newly accepted and connected socket, unless nil.
  Accept(SockHandler) error

  // Perform requests
  Request(op string, in interface{}, out interface{}) error
  BufferRequest(op string, in []byte) ([]byte, error)
  StreamRequest(op string) StreamRequest
  Notify(name string, in interface{}) error
  BufferNotify(name string, in []byte) error

  // Access Handlers associated with this socket
  Handlers() Handlers

  // Associate some application-specific data with this socket
  SetUserData(interface{})
  GetUserData() interface{}

  // Enable streaming requests and set the limit for how many streaming requests this socket
  // can handle at the same time. Setting this to `0` disables streaming requests alltogether
  // (the default) while setting this to a large number might be cause for security concerns
  // as a malicious peer could send many "start stream" messages, but never sending
  // any "end stream" messages, slowly exhausting memory.
  // When accepting connections, connected sockets inherit this value.
  SetStreamReqLimit(int)

  // Address of this socket
  Addr() string

  // Close this socket
  Close() error

  // Set a function to be closed when the socket closes
  SetCloseFunc(func(Sock))
}

type SockHandler func(Sock)

type StreamRequest interface {
  Write([]byte) error
  End() error
  Read() ([]byte, error)
}

func NewSock(h Handlers) Sock {
  return &socket{handlers:h}
}

// Creates two sockets which are connected to eachother
func Pipe() (Sock, Sock, error) {
  c1, c2 := net.Pipe()
  s1 := NewSock(DefaultHandlers)
  s2 := NewSock(DefaultHandlers)
  s1.Adopt(c1)
  s2.Adopt(c2)
  // Note: We deliberately ignore performing a handshake
  go s1.Read()
  go s2.Read()
  return s1, s2, nil
}

// Connect to a server via `how` at `addr`. Unless there's an error, the returned socket is
// already reading in a different goroutine and is ready to be used.
func Connect(how, addr string) (Sock, error) {
  c, err := net.Dial(how, addr)
  if err != nil {
    return nil, err
  }
  s := NewSock(DefaultHandlers)
  s.Adopt(c)
  if err := s.Handshake(); err != nil {
    return nil, err
  }
  go s.Read()
  return s, nil
}

// Start a `how` server listening for connections at `addr`. You need to call Accept() on the
// returned socket to start accepting connections.
func Listen(how, addr string) (Sock, error) {
  l, err := net.Listen(how, addr)
  if err != nil {
    return nil, err
  }

  s := NewSock(DefaultHandlers)
  s.AdoptListener(l)

  if how == "unix" || how == "unixpacket" {
    // Unix sockets must be unlink()ed before being reused again.
    // Handle common process-killing signals so we can gracefully shut down.
    sigc := make(chan os.Signal, 1)
    signal.Notify(sigc, os.Interrupt, os.Kill, syscall.SIGTERM)
    go func(c chan os.Signal) {
      <-c  // Wait for a signal
      //sig := <-c  // Wait for a signal
      //log.Printf("Caught signal %s: shutting down.", sig)
      s.Close()  // Stop listening and unlink the socket
      os.Exit(0)
    }(sigc)
  }

  return s, nil
}

// Start a `how` server accepting connections at `addr`.
func Serve(how, addr string, handler SockHandler) error {
  s, err := Listen(how, addr)
  if err != nil {
    return err
  }
  return s.Accept(handler)
}

// -------------------------------------------------------------------------------------

type pendingResMap  map[string]chan interface{}
type pendingReqMap  map[string]chan []byte

type socket struct {
  handlers       Handlers
  wmu            sync.Mutex          // guards writes on conn
  listener       net.Listener        // non-nil after successful call to Listen
  conn           io.ReadWriteCloser  // non-nil after successful call to Connect or accept
  closeFunc      func(Sock)
  userData       interface{}

  // Used for performing requests:
  nextOpID       uint
  pendingRes     pendingResMap
  pendingResMu   sync.RWMutex

  // Used for streaming requests:
  streamReqLimit int
  pendingReq     pendingReqMap
  pendingReqMu   sync.RWMutex
}


func (s *socket) Adopt(c io.ReadWriteCloser) {
  if s.listener != nil || s.conn != nil {
    panic("already adopted")
  }
  s.conn = c
}


func (s *socket) AdoptListener(listener net.Listener) {
  if s.listener != nil || s.conn != nil {
    panic("already adopted")
  }
  s.listener = listener
}


// ===========================================================================================

const (
  reqHandlerTypeBuf = reqHandlerType(iota)
)
type reqHandlerType int

// ----------------------------------------------------------------------------------------------

func (s *socket) getResChan(id string) chan interface{} {
  s.pendingResMu.RLock()
  defer s.pendingResMu.RUnlock()
  if s.pendingRes == nil {
    return nil
  }
  return s.pendingRes[id]
}


func (s *socket) allocResChan() (string, chan interface{}) {
  ch := make(chan interface{})

  s.pendingResMu.Lock()
  defer s.pendingResMu.Unlock()

  id := string(makeFixnumBuf(3, uint64(s.nextOpID), 36))
  s.nextOpID++
  if s.nextOpID == 46656 {
    // limit for base36 within 3 digits (36^2=46656)
    s.nextOpID = 0
  }

  if s.pendingRes == nil {
    s.pendingRes = make(pendingResMap)
  }
  s.pendingRes[id] = ch

  return id, ch
}


func (s *socket) deallocResChan(id string) {
  s.pendingResMu.Lock()
  defer s.pendingResMu.Unlock()
  delete(s.pendingRes, id)
}

// ----------------------------------------------------------------------------------------------

func (s *socket) getReqChan(id string) chan []byte {
  s.pendingReqMu.RLock()
  defer s.pendingReqMu.RUnlock()
  if s.pendingReq == nil {
    return nil
  }
  return s.pendingReq[id]
}


func (s *socket) allocReqChan(id string) chan []byte {
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


func (s *socket) deallocReqChan(id string) {
  s.pendingReqMu.Lock()
  defer s.pendingReqMu.Unlock()
  delete(s.pendingReq, id)
}

// ----------------------------------------------------------------------------------------------

func (s *socket) writeMsg(t MsgType, id, op string, buf []byte) error {
  s.wmu.Lock()
  defer s.wmu.Unlock()
  if _, err := s.conn.Write(MakeMsg(t, id, op, len(buf))); err != nil {
    return err
  }
  _, err := s.conn.Write(buf)
  return err
}


func (s *socket) BufferRequest(op string, buf []byte) ([]byte, error) {
  id, ch := s.allocResChan()
  defer s.deallocResChan(id)

  //fmt.Printf("BufferRequest: writeMsg(%v, %v, %v)\n", id, op, buf)

  if err := s.writeMsg(MsgTypeSingleReq, id, op, buf); err != nil {
    return nil, err
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
      return nil, errors.New(string(resbuf.b))
      // TODO: return an error type which contains the buffer, because the buffer might not be
      // a string.
    }
    // Note: This particular function requires the response to be buffered and not streaming
    return resbuf.b, errors.New("unexpected message "+string(byte(resbuf.t)))
  }
  return nil, nil
}


func (s *socket) Request(op string, in interface{}, out interface{}) error {
  inbuf, err := json.Marshal(in)
  if err != nil {
    return err
  }
  outbuf, err := s.BufferRequest(op, inbuf)
  if err != nil {
    return err
  }
  return json.Unmarshal(outbuf, out)
}


func (s *socket) StreamRequest(op string) StreamRequest {
  return &streamRequest{sock:s, op:op}
}


func (s *socket) BufferNotify(t string, buf []byte) error {
  return s.writeMsg(MsgTypeNotification, "", t, buf)
}

func (s *socket) Notify(t string, v interface{}) error {
  if buf, err := json.Marshal(v); err != nil {
    return err
  } else {
    return s.BufferNotify(t, buf)
  }
}

// ===========================================================================================

type streamRequest struct {
  sock    *socket
  op      string
  id      string
  started bool // request started?
  ended   bool // response ended?
  ch      chan interface{}
}

func (r *streamRequest) finalize() {
  if r.id != "" {
    r.sock.deallocResChan(r.id)
    r.id = ""
  }
}

func (r *streamRequest) Write(b []byte) error {
  if r.started == false {
    r.started = true
    r.id, r.ch = r.sock.allocResChan()
    if err := r.sock.writeMsg(MsgTypeStreamReq, r.id, r.op, b); err != nil {
      r.finalize()
      return err
    }
  } else {
    if err := r.sock.writeMsg(MsgTypeStreamReqPart, r.id, "", b); err != nil {
      r.finalize()
      return err
    }
  }
  return nil
}

func (r *streamRequest) End() error {
  err := r.sock.writeMsg(MsgTypeStreamReqPart, r.id, "", nil)
  if err != nil {
    r.finalize()
  }
  return err
}

func (r *streamRequest) Read() ([]byte, error) {
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

// ===========================================================================================

func (s *socket) readDiscard(readz int) error {
  if readz != 0 {
    // todo: is there a better way to read data w/o copying it into a buffer?
    err := readn(s.conn, make([]byte, readz))
    return err
  }
  return nil
}


func (s *socket) respondErr(readz int, id, errmsg string) error {
  if err := s.readDiscard(readz); err != nil {
    return err
  }
  s.wmu.Lock()
  defer s.wmu.Unlock()
  if _, err := WriteErrorRes(s.conn, id, len(errmsg)); err != nil {
    return err
  }
  _, err := s.conn.Write([]byte(errmsg))
  return err
}


func (s *socket) respondOK(id string, outbuf []byte) error {
  s.wmu.Lock()
  defer s.wmu.Unlock()
  if _, err := WriteSingleRes(s.conn, id, len(outbuf)); err != nil {
    return err
  }
  if len(outbuf) != 0 {
    _, err := s.conn.Write(outbuf)
    return err
  }
  return nil
}


func (s *socket) findHandlerOrResErr(id, op string, size int) interface{} {
  handler := s.handlers.FindRequestHandler(op)
  if handler == nil {
    if err := s.respondErr(size, id, "unknown operation \""+op+"\""); err != nil {
      panic("failed to send error")
    }
  }
  return handler
}


func (s *socket) readSingleReq(id, op string, size int) error {
  handlerval := s.findHandlerOrResErr(id, op, size)
  if handlerval == nil {
    return nil
  }

  handler, ok := handlerval.(BufferReqHandler)
  if ok == false {
    return s.respondErr(size, id, "buffered request not supported")
  }

  // Buffered handler
  inbuf := make([]byte, size)
  if err := readn(s.conn, inbuf); err != nil {
    return err
  }
  // Dispatch handler
  go func() {
    outbuf, err := handler(s, op, inbuf)
    if err != nil {
      if err := s.respondErr(0, id, err.Error()); err != nil {
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


func (s *socket) readStreamReq(id, op string, size int) error {
  if len(s.pendingReq) >= s.streamReqLimit {
    if s.streamReqLimit == 0 {
      return s.respondErr(size, id, "stream request not supported")
    } else {
      return s.respondErr(size, id, "stream request limit")
    }
  }

  handlerval := s.findHandlerOrResErr(id, op, size)
  if handlerval == nil {
    return nil
  }

  handler, ok := handlerval.(StreamReqHandler)
  if ok == false {
    return s.respondErr(size, id, "streaming request not supported")
  }

  // Read first buff
  inbuf := make([]byte, size)
  if err := readn(s.conn, inbuf); err != nil {
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
    return s.writeMsg(MsgTypeStreamRes, id, "", b)
  }

  // Dispatch handler
  go func () {
    if err := handler(s, op, rch, writer); err != nil {
      s.deallocReqChan(id)
      if err := s.respondErr(0, id, err.Error()); err != nil {
        log.Println(err)
        s.Close()
      }
    }
    if wroteEOS == false {
      // automatically writing EOS unless it was written by handler
      if err := s.writeMsg(MsgTypeStreamRes, id, "", nil); err != nil {
        log.Println(err)
        s.Close()
      }
    }
  }()

  return nil
}


func (s *socket) readStreamReqPart(id string, size int) error {
  var b []byte = nil

  if size != 0 {
    b = make([]byte, size)
    if err := readn(s.conn, b); err != nil {
      return err
    }
  }

  if rch := s.getReqChan(id); rch != nil {
    rch <- b
  } else if s.streamReqLimit == 0 {
    return errors.New("illegal message")  // There was no "start stream" message
  } // else: ignore msg

  return nil
}


// -----------------------------------------------------------------------------------------------

type resbuffer struct {
  t MsgType
  b []byte
}

func (s *socket) readRes(t MsgType, id string, size int) error {
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
      ch <- resbuffer{t, buf}

    default:
      panic("unexpected req handler type")
    }
  } else {
    panic("unexpected req handler type")
  }

  return nil
}


func (s *socket) readNotification(name string, size int) error {
  handler := s.handlers.FindNotificationHandler(name)

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


func (s *socket) Handshake() error {
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


func (s *socket) Read() error {
  defer func() {
    // recover from a faulty readLoop by closing the connection
    if r := recover(); r != nil {
      log.Println("gotalk.Sock panic:", r)
      s.Close()
    }
  }()

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
    t, id, name, size, err := ReadMsg(s.conn)
    if err != nil {
      s.Close()
      return err
    }

    //fmt.Printf("readLoop: msg: t=%c  id=%v  name=%v  size=%v\n", byte(t), id, name, size)

    switch t {
      case MsgTypeSingleReq:
        err = s.readSingleReq(id, name, int(size))

      case MsgTypeStreamReq:
        err = s.readStreamReq(id, name, int(size))

      case MsgTypeStreamReqPart:
        err = s.readStreamReqPart(id, int(size))

      case MsgTypeSingleRes, MsgTypeStreamRes, MsgTypeErrorRes:
        err = s.readRes(t, id, int(size))

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


func (s *socket) accept(c net.Conn, sockHandler SockHandler) {
  s2 := NewSock(s.handlers)
  s2.SetStreamReqLimit(s.streamReqLimit)
  s2.Adopt(c)
  if err := s2.Handshake(); err == nil {
    if sockHandler != nil {
      sockHandler(s2)
    }
    s2.Read()
  }
}


func (s *socket) Accept(sockHandler SockHandler) error {
  for {
    c, err := s.listener.Accept()
    if err != nil {
      return err
    }
    go s.accept(c, sockHandler)
  }
  s.listener.Close()
  return nil
}


func (s *socket) Handlers() Handlers {
  return s.handlers
}


func (s *socket) SetUserData(d interface{}) {
  s.userData = d
}

func (s *socket) GetUserData() interface{} {
  return s.userData
}


func (s *socket) SetStreamReqLimit(limit int) {
  s.streamReqLimit = limit
}


func (s *socket) Addr() string {
  if s.conn != nil {
    if netconn, ok := s.conn.(net.Conn); ok {
      return netconn.RemoteAddr().String()
    }
  } else if s.listener != nil {
    return s.listener.Addr().String()
  }
  return ""
}


func (s *socket) Close() error {
  if s.conn != nil {
    err := s.conn.Close()
    s.conn = nil
    if s.closeFunc != nil {
      s.closeFunc(s)
    }
    return err
  } else if s.listener != nil {
    err := s.listener.Close()
    s.listener = nil
    if s.closeFunc != nil {
      s.closeFunc(s)
    }
    return err
  } else {
    return nil
  }
}


func (s *socket) SetCloseFunc(f func(Sock)) {
  s.closeFunc = f
}
