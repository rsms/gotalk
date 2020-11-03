package gotalk

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// Returned by (Sock)BufferRequest when a streaming response is recieved
var (
	ErrUnexpectedStreamingRes = errors.New("unexpected streaming response")
	ErrSockClosed             = errors.New("socket closed")
)

type pendingResMap map[string]chan Response
type pendingReqMap map[string]chan []byte

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
	// If this is 0, automatic sending of heartbeats is disabled.
	// Defaults to 20 seconds when created with NewSock.
	HeartbeatInterval time.Duration

	// If not nil, this function is invoked when a heartbeat is recevied
	OnHeartbeat func(load int, t time.Time)

	// -------------------------------------------------------------------------
	// Used by connected sockets
	connmu    sync.RWMutex       // guards writes on conn and conn itself (W)
	conn      io.ReadWriteCloser // non-nil after successful call to Connect or accept
	closex    uint32             // atomic switch for closing conn (see Close())
	closeCode int32              // protocol error (ProtocolErrorXXX = closeCode-1)

	// Used for sending requests:
	nextOpID     uint32
	pendingRes   pendingResMap
	pendingResMu sync.RWMutex

	// Used for handling streaming requests:
	pendingReq   pendingReqMap
	pendingReqMu sync.RWMutex

	// Used for graceful shutdown
	shutdownWg *sync.WaitGroup // non-nil means that the socket has been shut down
}

func NewSock(h *Handlers) *Sock {
	// if you change this, also update WebSocket initialization in websocket.go
	return &Sock{
		Handlers:          h,
		HeartbeatInterval: 20 * time.Second,
	}
}

// Creates two sockets which are connected to eachother without any resource limits.
// If `handlers` is nil, DefaultHandlers are used.
// If `limits` is nil, DefaultLimits are used.
func Pipe(handlers *Handlers, limits *Limits) (*Sock, *Sock, error) {
	if handlers == nil {
		handlers = DefaultHandlers
	}
	c1, c2 := net.Pipe()
	s1 := NewSock(handlers)
	s2 := NewSock(handlers)
	s1.Adopt(c1)
	s2.Adopt(c2)
	// Note: We deliberately ignore performing a handshake
	go s1.Read(limits)
	go s2.Read(limits)
	return s1, s2, nil
}

// Connect to a server via `how` at `addr`.
// Unless there's an error, the returned socket is already reading in a different
// goroutine and is ready to be used.
func Connect(how, addr string) (*Sock, error) {
	s := NewSock(DefaultHandlers)
	return s, s.Connect(how, addr, DefaultLimits)
}

// Connect to a server via `how` at `addr` over TLS.
// Unless there's an error, the returned socket is already reading in a different
// goroutine and is ready to be used.
func ConnectTLS(how, addr string, config *tls.Config) (*Sock, error) {
	s := NewSock(DefaultHandlers)
	return s, s.ConnectTLS(how, addr, DefaultLimits, config)
}

// Adopt an I/O stream, which should already be in a "connected" state.
// After adopting a new connection, you should call Handshake to perform the protocol
// handshake, followed by Read to read messages.
func (s *Sock) Adopt(r io.ReadWriteCloser) {
	// lock to wait for any ongoing writes
	s.connmu.Lock()
	s.conn = r
	atomic.StoreInt32(&s.closeCode, 0)
	atomic.StoreUint32(&s.closex, 0)
	s.connmu.Unlock()
}

// ----------------------------------------------------------------------------------------------

// Connect to a server via `how` at `addr`
func (s *Sock) Connect(how, addr string, limits *Limits) error {
	c, err := net.Dial(how, addr)
	if err != nil {
		return err
	}
	return s.ConnectReader(c, limits)
}

// Connect to a server via `how` at `addr` over TLS.
// tls.Config is optional; passing nil is equivalent to &tls.Config{}
//
func (s *Sock) ConnectTLS(how, addr string, limits *Limits, config *tls.Config) error {
	if config == nil {
		config = &tls.Config{
			RootCAs: TLSCertPool(),
		}
	} else if config.RootCAs == nil {
		config = config.Clone()
		config.RootCAs = TLSCertPool()
	}
	c, err := tls.Dial(how, addr, config)
	if err != nil {
		return err
	}
	return s.ConnectReader(c, limits)
}

// Take control over reader r, perform initial handshake
// and begin communication on a background goroutine.
func (s *Sock) ConnectReader(r io.ReadWriteCloser, limits *Limits) error {
	s.Adopt(r)
	if err := s.Handshake(); err != nil {
		return err
	}
	go s.Read(limits)
	return nil
}

// Access the socket's underlying connection
func (s *Sock) Conn() io.ReadWriteCloser {
	s.connmu.RLock()
	conn := s.conn
	s.connmu.RUnlock()
	return conn
}

// String returns a name that uniquely identifies the socket during its lifetime
func (s *Sock) String() string {
	return fmt.Sprintf("%p", s)
}

// ----------------------------------------------------------------------------------------------

func (s *Sock) registerResChan(ch chan Response) string {
	s.pendingResMu.Lock()
	id := string(FormatRequestID(s.nextOpID))
	s.nextOpID++
	if s.pendingRes == nil {
		s.pendingRes = make(pendingResMap)
	}
	s.pendingRes[id] = ch
	s.pendingResMu.Unlock()
	return id
}

func (s *Sock) forgetResChan(id string) {
	s.pendingResMu.Lock()
	if ch := s.pendingRes[id]; ch != nil {
		delete(s.pendingRes, id)
	}
	s.pendingResMu.Unlock()
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
	delete(s.pendingReq, id)
	s.pendingReqMu.Unlock()
}

// ----------------------------------------------------------------------------------------------

func (s *Sock) writeMsg(t MsgType, id, op string, wait uint32, buf []byte) error {
	s.connmu.Lock()
	var err error
	if s.conn == nil {
		err = ErrSockClosed
	} else {
		msg := MakeMsg(t, id, op, wait, uint32(len(buf)))
		if _, err = s.conn.Write(msg); err == nil && len(buf) != 0 {
			_, err = s.conn.Write(buf)
		}
	}
	s.connmu.Unlock()
	return err
}

// Send a single-buffer request.
// A response should be received from reschan.
func (s *Sock) SendRequest(r *Request, reschan chan Response) error {
	if s.shutdownWg != nil {
		return ErrSockClosed
	}
	id := s.registerResChan(reschan)
	err := s.writeMsg(r.MsgType, id, r.Op, 0, r.Data)
	if err != nil {
		s.forgetResChan(id)
		if closeError := s.checkCloseCode(); closeError != nil {
			err = closeError
		}
	}
	return err
}

func (s *Sock) checkCloseCode() error {
	closeCode := atomic.LoadInt32(&s.closeCode)
	if closeCode == 0 {
		return nil
	}
	return protocolError(closeCode - 1) // -1 since procotol error codes are 0-based
}

// Send a single-buffer request, wait for and return the response.
// Automatically retries the request if needed.
func (s *Sock) BufferRequest(op string, buf []byte) ([]byte, error) {
	reschan := make(chan Response, 1)
	req := NewRequest(op, buf)
	for {
		err := s.SendRequest(req, reschan)
		if err != nil {
			if closeError := s.checkCloseCode(); closeError != nil {
				err = closeError
			}
			return nil, err
		}

		// await response
		res, ok := <-reschan
		if !ok {
			// channel closed
			err = ErrSockClosed
			if closeError := s.checkCloseCode(); closeError != nil {
				err = closeError
			}
			return nil, err
		}

		if res.IsError() {
			if res.Wait > 0 {
				return nil, protocolError(int32(res.Wait))
			}
			return nil, &res
		}

		if res.IsStreaming() {
			return nil, ErrUnexpectedStreamingRes
		}

		if res.IsRetry() {
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
	id := s.registerResChan(reschan)
	return &StreamRequest{sock: s, op: op, id: id}, reschan
}

// Send a single-buffer notification
func (s *Sock) BufferNotify(name string, buf []byte) error {
	if s.shutdownWg != nil {
		return ErrSockClosed
	}
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

func (s *Sock) respondRetry(readz int, id string, wait uint32, msg string) error {
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
type writeDeadline interface {
	SetWriteDeadline(time.Time) error
}

func (s *Sock) readBufferReq(lim *limitsImpl, id, op string, size int) error {
	if lim.incBufferReq() == false {
		return s.respondRetry(size, id, lim.waitBufferReq(), "request rate limit")
	}

	handler := s.Handlers.FindBufferRequestHandler(op)
	if handler == nil {
		err := s.respondError(size, id, "unknown operation \""+op+"\"")
		lim.decBufferReq()
		return err
	}

	// Read complete payload
	inbuf := make([]byte, size)
	if _, err := readn(s.conn, inbuf); err != nil {
		lim.decBufferReq()
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
			lim.decBufferReq()
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
	return z, w.s.writeMsg(MsgTypeStreamRes, w.id, "", 0, []byte(s))
}

func (w *streamWriter) Close() error {
	if !w.wroteEOS {
		w.wroteEOS = true
		return w.s.writeMsg(MsgTypeStreamRes, w.id, "", 0, nil)
	}
	return nil
}

func (s *Sock) readStreamReq(lim *limitsImpl, id, op string, size int) error {
	if lim.incStreamReq() == false {
		if lim.streamReqEnabled() {
			return s.respondRetry(size, id, lim.waitStreamReq(), "request rate limit")
		} else {
			return s.respondError(size, id, "stream requests not supported")
		}
	}

	handler := s.Handlers.FindStreamRequestHandler(op)
	if handler == nil {
		err := s.respondError(size, id, "unknown operation \""+op+"\"")
		lim.decStreamReq()
		return err
	}

	// Read first buff
	inbuf := make([]byte, size)
	if _, err := readn(s.conn, inbuf); err != nil {
		lim.decStreamReq()
		return err
	}

	// Create read chan
	rch := s.allocReqChan(id)
	rch <- inbuf

	// Dispatch handler
	go func() {
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
		lim.decStreamReq()
	}()

	return nil
}

func (s *Sock) readStreamReqPart(lim *limitsImpl, id string, size int) error {
	rch := s.getReqChan(id)
	if rch == nil {
		return errors.New("illegal message") // There was no "start stream" message
	}

	var b []byte = nil

	if size != 0 {
		b = make([]byte, size)
		if _, err := readn(s.conn, b); err != nil {
			lim.decStreamReq()
			return err
		}
	}

	rch <- b
	return nil
}

// -----------------------------------------------------------------------------------------------

func (s *Sock) readResponse(t MsgType, id string, wait, size int) error {
	// read payload
	var buf []byte
	if size != 0 {
		buf = make([]byte, size)
		if _, err := readn(s.conn, buf); err != nil {
			s.forgetResChan(id)
			return err
		}
	}

	// get response channel and hold the lock until we've sent the response
	s.pendingResMu.Lock()
	ch := s.pendingRes[id]
	if ch != nil && t != MsgTypeStreamRes {
		delete(s.pendingRes, id)
	}
	s.pendingResMu.Unlock()

	if ch != nil {
		ch <- Response{t, buf, time.Duration(wait) * time.Millisecond}
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

func protocolError(code int32) error {
	switch code {
	case ProtocolErrorAbnormal:
		return ErrAbnormal
	case ProtocolErrorUnsupported:
		return ErrUnsupported
	case ProtocolErrorInvalidMsg:
		return ErrInvalidMsg
	case ProtocolErrorTimeout:
		return ErrTimeout
	default:
		return errors.New("unknown error")
	}
}

func (s *Sock) sendHeartbeats(stopChan chan bool) {
	// Sleep for a very short amount of time to allow modification of HeartbeatInterval after
	// e.g. a call to Connect
	time.Sleep(time.Millisecond)
	var bufa [16]byte
	buf := bufa[:]
	for {
		// load is just the number of current goroutines. There has to be a more interesting "load"
		// number to convey...
		g := float32(runtime.NumGoroutine()-3) / 100000.0
		if g > 1 {
			g = 1
		} else if g < 0 {
			g = 0
		}
		if err := s.SendHeartbeat(g, buf); err != nil {
			return
		}
		select {
		case <-time.After(s.HeartbeatInterval):
			continue
		case <-stopChan:
			return
		}
	}
}

func (s *Sock) SendHeartbeat(load float32, buf []byte) error {
	msg := MakeHeartbeatMsg(uint16(load*float32(HeartbeatMsgMaxLoad)), buf)
	s.connmu.Lock()
	var err error
	if s.conn != nil {
		_, err = s.conn.Write(msg)
	}
	s.connmu.Unlock()
	return err
}

type netLocalAddressable interface {
	LocalAddr() net.Addr
}

func (s *Sock) setProcolError(protocolErrorCode int32) {
	// procol error codes are 0-based
	atomic.StoreInt32(&s.closeCode, protocolErrorCode+1)
}

// After completing a succesful handshake, call this function to read messages received to this
// socket. Does not return until the socket is closed.
// If HeartbeatInterval > 0 this method also sends automatic heartbeats.
func (s *Sock) Read(limits *Limits) error {
	if s.shutdownWg != nil {
		return ErrSockClosed
	}

	lim := makeLimitsImpl(limits)

	s.connmu.RLock()
	conn := s.conn
	s.connmu.RUnlock()

	hasReadDeadline := lim.readTimeout != time.Duration(0)

	// Pipes doesn't support deadlines
	netaddr, ok := conn.(netLocalAddressable)
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
		heartbeatStopChan = make(chan bool, 1)
		go s.sendHeartbeats(heartbeatStopChan)
	}

	var err error
	readbuf := make([]byte, 128)

readloop:
	for {

		// debug: read a chunk and print it
		// b := make([]byte, 128)
		// if _, err := conn.Read(b); err != nil {
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
			if rd, ok := conn.(readDeadline); ok {
				if err = rd.SetReadDeadline(time.Now().Add(lim.readTimeout)); err != nil {
					// If we failed to set read timeout, close socket immediately and report error.
					// The alternative, to ignore that read deadline could not be set, would be dangerous
					// in case that the user relies on timeouts for resource management and security.
					s.Close()
					return err
				}
			}
		}

		// Read next message
		t, id, name, wait, size, err1 := ReadMsg(conn, readbuf)
		err = err1

		if err == nil {
			// fmt.Printf("Read: msg: t=%c  id=%q  name=%q  size=%v\n", byte(t), id, name, size)

			switch t {
			case MsgTypeSingleReq:
				err = s.readBufferReq(&lim, id, name, int(size))

			case MsgTypeStreamReq:
				err = s.readStreamReq(&lim, id, name, int(size))

			case MsgTypeStreamReqPart:
				err = s.readStreamReqPart(&lim, id, int(size))

			case MsgTypeSingleRes, MsgTypeStreamRes, MsgTypeErrorRes, MsgTypeRetryRes:
				err = s.readResponse(t, id, int(wait), int(size))

			case MsgTypeNotification:
				err = s.readNotification(name, int(size))

			case MsgTypeHeartbeat:
				if s.OnHeartbeat != nil {
					s.OnHeartbeat(int(wait), time.Unix(int64(size), 0))
				}

			case MsgTypeProtocolError:
				code := int32(size)
				s.setProcolError(code)
				if s.shutdownWg == nil {
					s.Close()
				}
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
				if s.shutdownWg == nil {
					s.CloseError(ProtocolErrorTimeout)
				}
			} else {
				// Broken connection (e.g. pipe error, connection reset by peer, etc.)
				err = io.EOF
				s.Close()
			}
			break
		}

	} // readloop

	if s.shutdownWg != nil {
		s.Close()
		s.shutdownWg.Done()
	}

	if heartbeatStopChan != nil {
		heartbeatStopChan <- true
	}

	return err
}

// Address of this socket
func (s *Sock) Addr() string {
	conn := s.Conn()
	if conn != nil {
		if netconn, ok := conn.(net.Conn); ok {
			return netconn.RemoteAddr().String()
		}
	}
	return ""
}

// Close this socket because of a protocol error (ProtocolErrorXXX)
func (s *Sock) CloseError(protocolErrorCode int32) error {
	if protocolErrorCode < 0 {
		panic("negative protocolErrorCode")
	}
	s.connmu.Lock()
	s.setProcolError(protocolErrorCode)
	msg := MakeMsg(MsgTypeProtocolError, "", "", 0, uint32(protocolErrorCode))
	s.conn.Write(msg) // ignore error
	s.connmu.Unlock()
	err := s.Close()
	return err
}

// IsClosed returns true if Close() has been called.
// It is safe for multiple goroutines to call this concurrently.
func (s *Sock) IsClosed() bool {
	return atomic.LoadUint32(&s.closex) > 0
}

// Close this socket.
// It is safe for multiple goroutines to call this concurrently.
func (s *Sock) Close() error {
	if atomic.AddUint32(&s.closex, 1) != 1 {
		// another goroutine won the race or already closed
		return nil
	}

	s.pendingResMu.Lock()
	s.connmu.Lock()

	// close the underlying connection
	err := s.conn.Close()

	// check for close error
	closeCode := atomic.LoadInt32(&s.closeCode)
	if closeCode > 0 {
		err = protocolError(closeCode - 1) // -1 since procotol error codes are 0-based
	}

	// end any pending request-response channels
	var errmsg []byte
	var waitarg time.Duration
	for _, ch := range s.pendingRes {
		if errmsg == nil {
			if err != nil {
				errmsg = []byte(err.Error())
			} else {
				errmsg = []byte(ErrSockClosed.Error())
			}
			if closeCode != 0 {
				waitarg = time.Duration(closeCode - 1)
			}
		}
		select {
		case ch <- Response{
			MsgType: MsgTypeErrorRes,
			Data:    errmsg,
			Wait:    waitarg,
		}:
		default:
		}
	}

	s.pendingRes = nil

	s.connmu.Unlock()
	s.pendingResMu.Unlock()

	// call CloseHandler
	if s.CloseHandler != nil {
		s.CloseHandler(s, int(closeCode))
	}

	return err
}

// Shut down this socket, giving it timeout time to complete any ongoing work.
//
// timeout should be a short duration as its used for I/O read and write timeout; any work in
// handlers does not account to the timeout (and is unlimited.) timeout is ignored if the
// underlying Conn() does not implement SetReadDeadline or SetWriteDeadline.
//
// This method returns immediately. Once all work is complete, calls s.Close() and wg.Done().
//
// This method should not be used with web socket connections. Instead, call Close() from
// your http.Server.RegisterOnShutdown handler.
//
func (s *Sock) Shutdown(wg *sync.WaitGroup, timeout time.Duration) error {
	if s.shutdownWg != nil {
		return ErrSockClosed
	}
	s.shutdownWg = wg

	// give the connection a very short time to complete reads & writes
	deadline := time.Now().Add(timeout)

	var err error
	if rd, ok := s.conn.(readDeadline); ok {
		// give a Read call a short amount of time to complete
		err = rd.SetReadDeadline(deadline)
	}
	if wd, ok := s.conn.(writeDeadline); ok {
		// give a Read call a short amount of time to complete
		err = wd.SetWriteDeadline(deadline)
	}
	return err
}
