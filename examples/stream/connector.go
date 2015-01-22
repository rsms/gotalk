package main
import (
  "fmt"
  "log"
  "github.com/rsms/gotalk"
)

func connector(port string) {
  s, err := gotalk.Connect("tcp", "localhost:"+port)
  if err != nil {
    log.Fatalln(err)
  }
  println("connected to", s.Addr())

  // Send a request with a streaming payload
  req := s.StreamRequest("joke")
  if err != nil { log.Fatalln(err) }
  if err := req.Write([]byte("tell me")); err != nil { log.Fatalln(err) }
  if err := req.Write([]byte(" a joke")); err != nil { log.Fatalln(err) }
  if err := req.Write([]byte(" or two")); err != nil { log.Fatalln(err) }
  if err := req.End(); err != nil { log.Fatalln(err) }
  // s000004joke00000007tell me
  // p00000000007 a joke
  // p00000000007 or two
  // p00000000000

  // Read streaming result
  for {
    outbuf, err := req.Read()
    if err != nil { log.Fatalln(err) }
    if outbuf == nil { break }  // end of stream
    fmt.Printf("joke: \"%s\"\n", string(outbuf))
  }

  s.Close()
}
