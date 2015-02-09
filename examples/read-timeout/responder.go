package main
import (
  "github.com/rsms/gotalk"
  "time"
  "fmt"
)

func responder(port string) {
  // A simple echo operation with a 500ms response delay
  gotalk.HandleBufferRequest("echo", func(s *gotalk.Sock, op string, buf []byte) ([]byte, error) {
    println("responder: handling request")
    return buf, nil
  })

  // Start a server
  s, err := gotalk.Listen("tcp", "localhost:"+port)
  if err != nil {
    panic(err)
  }

  // Print when we receive heartbeats
  s.OnHeartbeat = func(load int, t time.Time) {
    fmt.Printf("responder: received heartbeat: load=%v, time=%v\n", load, t)
  }

  // Configure limits with a read timeout of one second
  s.Limits = gotalk.NewLimits(0, 0)
  s.Limits.SetReadTimeout(time.Second)

  // Accept connections
  println("responder: listening at", s.Addr())
  go s.Accept()
}
