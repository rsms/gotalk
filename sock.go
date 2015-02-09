package gotalk
import (
  "encoding/json"
  "errors"
  "io"
  "log"
  "net"
  "sync"
  "fmt"
  "time"
  "runtime"
)

var (
  ErrUnexpectedStreamingRes = errors.New("unexpected streaming response")
)

type pendingResMap  map[string]chan Response
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

  // A function to be called when the socket closes.
  // If the socket was closed because of a protocol error, `code` is >=0 and represents a
  // ProtocolError* constant.
  CloseHandler func(s *Sock, code int)

  // Automatically retry requests which can be retried
  AutoRetryRequests bool

  // HeartbeatInterval controls how much time a socket waits between sending its heartbeats.
  // If this is 0, automatic sending of heartbeats is disabled. Defaults to 20 seconds.
  HeartbeatInterval time.Duration

  // If not nil, this function is invoked when a heartbeat is recevied
  OnHeartbeat func(load int, t time.Time)

  // -------------------------------------------------------------------------
  // Used by connected sockets
  wmu            sync.Mutex          // guards writes on conn
  conn           io.ReadWriteCloser  // non-nil after successful call to Connect or accept
  closeCode      int  // protocol error, or -1

  // Used for sending requests:
  nextOpID       uint32
  pendingRes     pendingResMap
  pendingResMu   sync.RWMutex

  // Used for handling streaming requests:
  pendingReq     pendingReqMap
  pendingReqMu   sync.RWMutex
}


func NewSock(h *Handlers) *Sock {
  return &Sock{Handlers:h, closeCode:-1, HeartbeatInterval:20 * time.Second}
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
func Connect(how, addr string) (*Sock, error) {
  s := NewSock(DefaultHandlers)
  return s, s.Connect(how, addr, DefaultLimits)
}


// Adopt an I/O stream, which should already be in a "connected" state.
// After adopting a new connection, you should call Handshake to perform the protocol
// handshake, followed by Read to read messages.
func (s *Sock) Adopt(c io.ReadWriteCloser) {
  s.conn = c
  s.closeCode = -1
}

// ----------------------------------------------------------------------------------------------

// Connect to a server via `how` at `addr`
func (s *Sock) Connect(how, addr string, limits Limits) error {
  c, err := net.Dial(how, addr)
  if err != nil {
    return err
  }
  s.Adopt(c)
  if err := s.Handshake(); err != nil {
    return err
  }
  go s.Read(limits)
  return nil
}

// Access the socket's underlying connection
func (s *Sock) Conn() io.ReadWriteCloser {
  return s.conn
}

// ----------------------------------------------------------------------------------------------

func (s *Sock) getResChan(id string) chan Response {
  s.pendingResMu.RLock()
  defer s.pendingResMu.RUnlock()
  if s.pendingRes == nil {
    return nil
  }
  return s.pendingRes[id]
}


func (s *Sock) allocResChan(ch chan Response) string {
  s.pendingResMu.Lock()
  defer s.pendingResMu.Unlock()

  id := string(FormatRequestID(s.nextOpID))
  s.nextOpID++

  if s.pendingRes == nil {
    s.pendingRes = make(pendingResMap)
  }
  s.pendingRes[id] = ch

  return id
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

var errNotConnected = errors.New("not connected")

func (s *Sock) writeMsg(t MsgType, id, op string, wait int, buf []byte) error {
  if s.conn == nil { return errNotConnected }
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


func (s *Sock) writeMsgString(t MsgType, id, op string, wait int, str string) error {
  if s.conn == nil { return errNotConnected }
  s.wmu.Lock()
  defer s.wmu.Unlock()
  if _, err := s.conn.Write(MakeMsg(t, id, op, wait, len(str))); err != nil {
    return err
  }
  if len(str) != 0 {
    _, err := io.WriteString(s.conn, str)
    return err
  }
  return nil
}


// Send a single-buffer request. A response should be received from reschan.
func (s *Sock) SendRequest(r *Request, reschan chan Response) error {
  id := s.allocResChan(reschan)
  if err := s.writeMsg(r.MsgType, id, r.Op, 0, r.Data); err != nil {
    s.deallocResChan(id)
    if s.closeCode != -1 {
      return protocolError(s.closeCode)
    }
    return err
  }
  return nil
}


// Send a single-buffer request, wait for and return the response.
// Automatically retries the request if needed.
func (s *Sock) BufferRequest(op string, buf []byte) ([]byte, error) {
  reschan := make(chan Response)
  req := NewRequest(op, buf)
  for {
    err := s.SendRequest(req, reschan)
    if err != nil {
      return nil, err
    }

    res := <- reschan

    if res.IsError() {
      return nil, &res
    } else if res.IsStreaming() {
      return nil, ErrUnexpectedStreamingRes
    } else if res.IsRetry() {
      if res.Wait != 0 {
        time.Sleep(res.Wait)
      }
    } else {
      return res.Data, nil
    }
  }
}


// Send a single-value request where the input and output values are JSON-encoded
func (s *Sock) Request(op string, in interface{}, out interface{}) error {
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


// Send a multi-buffer streaming request
func (s *Sock) StreamRequest(op string) (*StreamRequest, chan Response) {
  reschan := make(chan Response)
  id := s.allocResChan(reschan)
  return &StreamRequest{sock:s, op:op, id:id}, reschan
}


// Send a single-buffer notification
func (s *Sock) BufferNotify(name string, buf []byte) error {
  return s.writeMsg(MsgTypeNotification, "", name, 0, buf)
}

// Send a single-value request where the value is JSON-encoded
func (s *Sock) Notify(name string, v interface{}) error {
  if buf, err := json.Marshal(v); err != nil {
    return err
  } else {
    return s.BufferNotify(name, buf)
  }
}

// ----------------------------------------------------------------------------------------------

func (s *Sock) readDiscard(readz int) error {
  if readz != 0 {
    // todo: is there a better way to read data w/o copying it into a buffer?
    _, err := readn(s.conn, make([]byte, readz))
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


type readDeadline interface {
  SetReadDeadline(time.Time) error
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

  // Read complete payload
  inbuf := make([]byte, size)
  if _, err := readn(s.conn, inbuf); err != nil {
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

// -----------------------------------------------------------------------------------------------

type streamWriter struct {
  s        *Sock
  id       string
  wroteEOS bool
}

func (w *streamWriter) Write(b []byte) (int, error) {
  z := len(b)
  if z == 0 {
    w.wroteEOS = true
  }
  return z, w.s.writeMsg(MsgTypeStreamRes, w.id, "", 0, b)
}

func (w *streamWriter) WriteString(s string) (n int, err error) {
  z := len(s)
  if z == 0 {
    w.wroteEOS = true
  }
  return z, w.s.writeMsgString(MsgTypeStreamRes, w.id, "", 0, s)
}

func (w *streamWriter) Close() error {
  if !w.wroteEOS {
    w.wroteEOS = true
    return w.s.writeMsg(MsgTypeStreamRes, w.id, "", 0, nil)
  }
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
  if _, err := readn(s.conn, inbuf); err != nil {
    limits.decStreamReq()
    return err
  }

  // Create read chan
  rch := s.allocReqChan(id)
  rch <- inbuf

  // Dispatch handler
  go func () {
    // TODO: recover?
    out := &streamWriter{s, id, false}
    if err := handler(s, op, rch, out); err != nil {
      s.deallocReqChan(id)
      if err := s.respondError(0, id, err.Error()); err != nil {
        log.Println(err)
        s.Close()
      }
    }
    if err := out.Close(); err != nil {
      s.Close()
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
    if _, err := readn(s.conn, b); err != nil {
      limits.decStreamReq()
      return err
    }
  }

  rch <- b
  return nil
}

// -----------------------------------------------------------------------------------------------

func (s *Sock) readResponse(t MsgType, id string, wait, size int) error {
  ch := s.getResChan(id)
  if ch == nil {
    // Unexpected response: discard and ignore
    return s.readDiscard(size)
  }

  if t != MsgTypeStreamRes || size == 0 {
    s.deallocResChan(id)
  }

  // read payload
  var buf []byte
  if size != 0 {
    buf = make([]byte, size)
    if _, err := readn(s.conn, buf); err != nil {
      return err
    }
  }

  ch <- Response{t, buf, time.Duration(wait) * time.Millisecond}

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
    if _, err := readn(s.conn, buf); err != nil {
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


var (
  ErrAbnormal    = errors.New("abnormal condition")
  ErrUnsupported = errors.New("unsupported protocol")
  ErrInvalidMsg  = errors.New("invalid protocol message")
  ErrTimeout     = errors.New("timeout")
)

func protocolError(code int) error {
  switch code {
    case ProtocolErrorAbnormal:    return ErrAbnormal
    case ProtocolErrorUnsupported: return ErrUnsupported
    case ProtocolErrorInvalidMsg:  return ErrInvalidMsg
    case ProtocolErrorTimeout:     return ErrTimeout
    default: return errors.New("unknown error")
  }
}

func (s *Sock) sendHeartbeats(stopChan chan bool) {
  // Sleep for a very short amount of time to allow modification of HeartbeatInterval after
  // e.g. a call to Connect
  time.Sleep(time.Millisecond)
  for {
    // load is just the number of current goroutines. There has to be a more interesting "load"
    // number to convey...
    g := float32(runtime.NumGoroutine() - 3) / 100000.0
    if g > 1 {
      g = 1
    } else if g < 0 {
      g = 0
    }
    if err := s.SendHeartbeat(g); err != nil {
      return
    }
    select {
      case <-time.After(s.HeartbeatInterval): continue
      case <-stopChan: return
    }
  }
}


func (s *Sock) SendHeartbeat(load float32) error {
  s.wmu.Lock()
  defer s.wmu.Unlock()
  if s.conn == nil { return errors.New("not connected") }
  msg := MakeHeartbeatMsg(uint16(load * float32(HeartbeatMsgMaxLoad)))
  _, err := s.conn.Write(msg)
  return err
}


type netLocalAddressable interface {
  LocalAddr() net.Addr
}


// After completing a succesful handshake, call this function to read messages received to this
// socket. Does not return until the socket is closed.
// If HeartbeatInterval > 0 this method also sends automatic heartbeats.
func (s *Sock) Read(limits Limits) error {
  readTimeout := limits.ReadTimeout()
  hasReadDeadline := readTimeout != time.Duration(0)
  var rd readDeadline

  // Pipes doesn't support deadlines
  netaddr, ok := s.conn.(netLocalAddressable)
  isPipe := ok && netaddr.LocalAddr().Network() == "pipe"
  if hasReadDeadline && isPipe {
    hasReadDeadline = false
  }

  // Start sending heartbeats
  var heartbeatStopChan chan bool
  if s.HeartbeatInterval > 0 && !isPipe {
    if s.HeartbeatInterval < time.Millisecond {
      panic("HeartbeatInterval < time.Millisecond")
    }
    heartbeatStopChan = make(chan bool)
    go s.sendHeartbeats(heartbeatStopChan)
  }

  var err error

  readloop: for s.conn != nil {

    // debug: read a chunk and print it
    // b := make([]byte, 128)
    // if _, err := s.conn.Read(b); err != nil {
    //   log.Println(err)
    //   break
    // }
    // fmt.Printf("Read: %v\n", string(b))
    // continue

    // debug: force close with error
    // s.CloseError(ProtocolErrorInvalidMsg)
    // return ErrInvalidMsg

    // Set read timeout
    if hasReadDeadline {
      var ok bool
      if rd, ok = s.conn.(readDeadline); ok {
        readTimeoutAt := time.Now().Add(readTimeout)
        // fmt.Printf("setting read timeout to %v  %v\n", readTimeout, readTimeoutAt)
        if err = rd.SetReadDeadline(readTimeoutAt); err != nil {
          panic("SetReadDeadline failed: " + err.Error())
        }
      }
    }

    // Read next message
    t, id, name, wait, size, err1 := ReadMsg(s.conn)
    err = err1
    if err == nil {
      // fmt.Printf("Read: msg: t=%c  id=%q  name=%q  size=%v\n", byte(t), id, name, size)

      switch t {
        case MsgTypeSingleReq:
          err = s.readBufferReq(limits, id, name, int(size))

        case MsgTypeStreamReq:
          err = s.readStreamReq(limits, id, name, int(size))

        case MsgTypeStreamReqPart:
          err = s.readStreamReqPart(limits, id, int(size))

        case MsgTypeSingleRes, MsgTypeStreamRes, MsgTypeErrorRes, MsgTypeRetryRes:
          err = s.readResponse(t, id, int(wait), int(size))

        case MsgTypeNotification:
          err = s.readNotification(name, int(size))

        case MsgTypeHeartbeat:
          if s.OnHeartbeat != nil {
            s.OnHeartbeat(int(wait), time.Unix(int64(size), 0))
          }

        case MsgTypeProtocolError:
          code := int(size)
          s.closeCode = code
          s.Close()
          err = protocolError(code)
          break readloop

        default:
          s.CloseError(ProtocolErrorInvalidMsg)
          err = ErrInvalidMsg
          break readloop
      }
    }

    if err != nil {
      if err == io.EOF {
        s.Close()
      } else if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
        s.CloseError(ProtocolErrorTimeout)
      } else {
        s.CloseError(ProtocolErrorInvalidMsg)
      }
    }
  }

  if heartbeatStopChan != nil {
    heartbeatStopChan <- true
  }

  return err
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


// Close this socket because of a protocol error
func (s *Sock) CloseError(code int) error {
  if s.conn != nil {
    s.closeCode = code
    s.wmu.Lock()
    s.conn.Write(MakeMsg(MsgTypeProtocolError, "", "", 0, code))
    s.wmu.Unlock()
    return s.Close()
  }
  return nil
}


// Close this socket
func (s *Sock) Close() error {
  if s.conn != nil {
    err := s.conn.Close()
    s.conn = nil
    if s.CloseHandler != nil {
      s.CloseHandler(s, s.closeCode)
    }
    return err
  }
  return nil
}
