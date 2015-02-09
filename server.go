package gotalk
import (
  "net"
  "os"
  "os/signal"
  "syscall"
  "time"
)


type SockHandler func(*Sock)


// Accepts socket connections
type Server struct {
  // Handlers associated with this listener. Accepted sockets inherit the value.
  Handlers *Handlers

  // Limits. Accepted sockets are subject to the same limits.
  Limits Limits

  // Function to be invoked just after a new socket connection has been accepted and
  // protocol handshake has sucessfully completed. At this point the socket is ready
  // to be used. However the function will be called in the socket's "read" goroutine,
  // meaning no messages will be received on the socket until this function returns.
  AcceptHandler SockHandler

  // Template value for accepted sockets. Defaults to 0 (no automatic heartbeats)
  HeartbeatInterval time.Duration

  // Template value for accepted sockets. Defaults to nil
  OnHeartbeat func(load int, t time.Time)

  listener net.Listener
}


// Create a new server already listening on `l`
func NewServer(h *Handlers, limits Limits, l net.Listener) *Server {
  return &Server{Handlers:h, Limits:limits, listener:l}
}


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


// Start a `how` server listening for connections at `addr`. You need to call Accept() on the
// returned socket to start accepting connections. `how` and `addr` are passed to `net.Listen()`
// and thus any values accepted by net.Listen are valid.
// The returned server has Handlers=DefaultHandlers and Limits=DefaultLimits set.
func Listen(how, addr string) (*Server, error) {
  l, err := net.Listen(how, addr)
  if err != nil {
    return nil, err
  }

  if tcpl, ok := l.(*net.TCPListener); ok {
    // Wrap TCP listener to enable TCP keep-alive
    l = &tcpKeepAliveListener{tcpl}
  }

  s := NewServer(DefaultHandlers, DefaultLimits, l)

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
    c, e := s.listener.Accept()
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
  if s.listener != nil {
    return s.listener.Addr().String()
  }
  return ""
}


// Stop listening for and accepting connections
func (s *Server) Close() error {
  if s.listener != nil {
    err := s.listener.Close()
    s.listener = nil
    return err
  }
  return nil
}
