package main
import (
  "time"
  "github.com/rsms/gotalk"
)


func responder(port string) {
  // A simple echo operation with a 500ms response delay
  gotalk.HandleBufferRequest("echo", func(s *gotalk.Sock, op string, buf []byte) ([]byte, error) {
    println("responder: handling request")
    time.Sleep(time.Millisecond * 400)
    return buf, nil
  })

  // Start a server
  s, err := gotalk.Listen("tcp", "localhost:"+port)
  if err != nil {
    panic(err)
  }

  // Limit this server to 5 concurrent requests (and disable streaming messages)
  s.Limits = gotalk.NewLimits(5, 0)

  // Accept connections
  println("listening at", s.Addr())
  go s.Accept()
}
