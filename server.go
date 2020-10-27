package gotalk

import (
	"crypto/tls"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type SockHandler func(*Sock)

// Accepts socket connections
type Server struct {
	// Handlers associated with this server. Accepted sockets inherit the value.
	*Handlers

	// Limits. Accepted sockets are subject to the same limits.
	*Limits

	// Function to be invoked just after a new socket connection has been accepted and
	// protocol handshake has sucessfully completed. At this point the socket is ready
	// to be used. However the function will be called in the socket's "read" goroutine,
	// meaning no messages will be received on the socket until this function returns.
	AcceptHandler SockHandler

	// Template value for accepted sockets. Defaults to 0 (no automatic heartbeats)
	HeartbeatInterval time.Duration

	// Template value for accepted sockets. Defaults to nil
	OnHeartbeat func(load int, t time.Time)

	// Transport
	Listener net.Listener
}

// Create a new server already listening on `l`
func NewServer(h *Handlers, limits *Limits, l net.Listener) *Server {
	return &Server{Handlers: h, Limits: limits, Listener: l}
}

// Start a `how` server listening for connections at `addr`.
// You need to call Accept() on the returned socket to start accepting connections.
// `how` and `addr` are passed to `net.Listen()` and thus any values accepted by
// net.Listen are valid.
// The returned server has Handlers=DefaultHandlers and Limits=DefaultLimits set,
// which you can change if you want.
func Listen(how, addr string) (*Server, error) {
	l, err := net.Listen(how, addr)
	if err != nil {
		return nil, err
	}
	l = wrapListener(l)
	s := NewServer(DefaultHandlers, DefaultLimits, l)
	return s, nil
}

// Start a `how` server listening for connections at `addr` with TLS certificates.
// You need to call Accept() on the returned socket to start accepting connections.
// `how` and `addr` are passed to `net.Listen()` and thus any values accepted by
// net.Listen are valid.
// The returned server has Handlers=DefaultHandlers and Limits=DefaultLimits set,
// which you can change if you want.
func ListenTLS(how, addr string, certFile, keyFile string) (*Server, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	return ListenTLSCustom(how, addr, &tls.Config{
		RootCAs:      TLSCertPool(),
		Certificates: []tls.Certificate{cert},
	})
}

// Start a `how` server listening for connections at `addr` with custom TLS configuration.
// You need to call Accept() on the returned socket to start accepting connections.
// `how` and `addr` are passed to `net.Listen()` and thus any values accepted by
// net.Listen are valid.
// The returned server has Handlers=DefaultHandlers and Limits=DefaultLimits set,
// which you can change if you want.
func ListenTLSCustom(how, addr string, config *tls.Config) (*Server, error) {
	l, err := net.Listen(how, addr)
	if err != nil {
		return nil, err
	}
	// must call wrapListener _before_ wrapping with tls.NewListener
	l = tls.NewListener(wrapListener(l), config)
	s := NewServer(DefaultHandlers, DefaultLimits, l)
	return s, nil
}

// Unix sockets must be unlink()ed before being reused again.
// If you don't manage this yourself already, this function provides a limited but
// quick way to deal with cleanup by installing a signal handler.
func (s *Server) EnableUnixSocketGC() {
	// Handle common process-killing signals so we can gracefully shut down.
	if _, ok := s.Listener.(*net.UnixListener); ok {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, os.Interrupt, os.Kill, syscall.SIGTERM)
		go func(c chan os.Signal) {
			<-c       // Wait for a signal
			s.Close() // Stop listening and unlink the socket
			os.Exit(0)
		}(sigc)
	}
}

// Start a `how` server accepting connections at `addr`
func Serve(how, addr string, acceptHandler SockHandler) error {
	s, err := Listen(how, addr)
	if err != nil {
		return err
	}
	s.AcceptHandler = acceptHandler
	return s.Accept()
}

// Accept connections. Blocks until Close() is called or an error occurs.
func (s *Server) Accept() error {
	var tempDelay time.Duration // how long to sleep on accept failure
	for {
		c, e := s.Listener.Accept()
		if e != nil {
			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				time.Sleep(tempDelay)
				continue
			}
			return e
		}
		go s.accept(c)
	}
}

func (s *Server) accept(c net.Conn) {
	s2 := NewSock(s.Handlers)
	s2.Adopt(c)
	if err := s2.Handshake(); err == nil {
		if s.AcceptHandler != nil {
			s.AcceptHandler(s2)
		}
		s2.HeartbeatInterval = s.HeartbeatInterval
		s2.OnHeartbeat = s.OnHeartbeat
		s2.Read(s.Limits)
	}
}

// Address this server is listening at
func (s *Server) Addr() string {
	if s.Listener != nil {
		return s.Listener.Addr().String()
	}
	return ""
}

// Stop listening for and accepting connections
func (s *Server) Close() error {
	if s.Listener != nil {
		err := s.Listener.Close()
		s.Listener = nil
		return err
	}
	return nil
}

// --------------------------------------------------------------
// internals

type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(30 * time.Second)
	return tc, nil
}

// called by NewServer. Possibly wrap a listener.
// Safe to call multiple times on the result.
// I.e. wrapListener(l) == wrapListener(wrapListener(l))
func wrapListener(l net.Listener) net.Listener {
	if tcpl, ok := l.(*net.TCPListener); ok {
		// Wrap TCP listener to enable TCP keep-alive
		return &tcpKeepAliveListener{tcpl}
	}
	return l
}
