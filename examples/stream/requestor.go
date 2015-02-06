// Demonstrates 
//
package main
import (
  "fmt"
  "log"
  "github.com/rsms/gotalk"
)

func requestor(port string) {
  s, err := gotalk.Connect("tcp", "localhost:"+port)
  if err != nil {
    log.Fatalln(err)
  }
  println("requestor: connected to", s.Addr())

  // Send a request with a streaming payload
  req, res := s.StreamRequest("joke")
  if err := req.Write([]byte("tell me")); err != nil { log.Fatalln(err) }
  if err := req.Write([]byte(" a joke")); err != nil { log.Fatalln(err) }
  if err := req.Write([]byte(" or two")); err != nil { log.Fatalln(err) }
  if err := req.End(); err != nil { log.Fatalln(err) }

  // Read streaming result
  for {
    r := <- res
    if r.IsError() {
      log.Fatalln(r)
    }
    if r.Data == nil {
      break
    }
    fmt.Printf("requestor: received response payload: %q\n", string(r.Data))
    if !r.IsStreaming() || r.Data == nil {
      break
    }
  }

  s.Close()
}
