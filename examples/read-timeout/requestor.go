package main
import (
  "github.com/rsms/gotalk"
  "fmt"
  "time"
  "io"
)

type slowWriter struct {
  c     io.ReadWriteCloser
  delay time.Duration
}
func (rwc *slowWriter) Read(p []byte) (n int, err error) {
  n, err = rwc.c.Read(p)
  if err == nil {
    // Note: This is a completely unreliable way of detecting messages. It works because
    // we are not sending any payloads starting with the byte 'e'.
    if p[0] == byte(gotalk.MsgTypeRetryRes) {
      println("requestor: Server asked us to retry the request")
    }
  }
  return
}
func (rwc *slowWriter) Write(p []byte) (n int, err error) {
  // Delay anything but writing single-request headers
  if err == nil && ( len(p) < 13 || p[0] != byte(gotalk.MsgTypeSingleReq) ) {
    time.Sleep(rwc.delay)
  }
  return rwc.c.Write(p)
}
func (rwc *slowWriter) Close() error {
  return rwc.c.Close()
}


func sendRequest(s *gotalk.Sock) {
  fmt.Printf("requestor: sending 'echo' request\n")
  b, err := s.BufferRequest("echo", []byte("Hello"))
  if err == gotalk.ErrTimeout {
    fmt.Printf("requestor: timed out\n")
  } else if err != nil {
    fmt.Printf("requestor: error %v\n", err.Error())
  } else {
    fmt.Printf("requestor: success: %v\n", string(b))
  }
}


func timeoutRequest(port string) {
  s, err := gotalk.Connect("tcp", "localhost:"+port)
  if err != nil { panic(err) }
  println("requestor: connected to", s.Addr())

  // Wrap the connection for slow writing to simulate a poor connection
  s.Adopt(&slowWriter{s.Conn(), 2 * time.Second})

  // Send a request -- it will take too long and time out
  sendRequest(s)

  s.Close()
}


func heartbeatKeepAlive(port string) {
  s, err := gotalk.Connect("tcp", "localhost:"+port)
  if err != nil { panic(err) }
  println("requestor: connected to", s.Addr())

  // As the responder has a one second timeout, set our heartbeat interval to half that time
  s.HeartbeatInterval = 500 * time.Millisecond

  // Sleep for 3 seconds
  time.Sleep(3 * time.Second)

  // Send a request, which will work since we have kept the connection alive with heartbeats
  sendRequest(s)

  s.Close()
}


func requestor(port string) {
  timeoutRequest(port)
  heartbeatKeepAlive(port)
}
