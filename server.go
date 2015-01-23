package gotalk
import (
  "net"
  "os"
  "os/signal"
  "syscall"
)


type SockHandler func(*Sock)


// Accepts socket connections
type Server struct {
  // Handlers associated with this listener. Accepted sockets inherit the value.
  Handlers *Handlers

  // Default value for accepted sockets' StreamReqLimit
  StreamReqLimit int

  // Function to be invoked just after a new socket connection has been accepted and
  // protocol handshake has sucessfully completed. At this point the socket is ready
  // to be used. However the function will be called in the socket's "read" goroutine,
  // meaning no messages will be received on the socket until this function returns.
  AcceptHandler SockHandler

  listener net.Listener
}


func NewServer(h *Handlers, l net.Listener) *Server {
  return &Server{Handlers:h, listener:l}
}


// Start a `how` server listening for connections at `addr`. You need to call Accept() on the
// returned socket to start accepting connections.
func Listen(how, addr string) (*Server, error) {
  l, err := net.Listen(how, addr)
  if err != nil {
    return nil, err
  }

  s := NewServer(DefaultHandlers, l)

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
  for {
    c, err := s.listener.Accept()
    if err != nil {
      return err
    }
    go s.accept(c)
  }
}

func (s *Server) accept(c net.Conn) {
  s2 := NewSock(s.Handlers)
  s2.StreamReqLimit = s.StreamReqLimit
  s2.Adopt(c)
  if err := s2.Handshake(); err == nil {
    if s.AcceptHandler != nil {
      s.AcceptHandler(s2)
    }
    s2.Read()
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
