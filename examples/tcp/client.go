package main
import (
  "fmt"
  "github.com/rsms/gotalk"
)

func client(port string) {
  gotalk.HandleBufferRequest("ping", func(_ *gotalk.Sock, _ string, b []byte) ([]byte, error) {
    fmt.Printf("client: handling 'ping' request: %q\n", string(b))
    return []byte("pong"), nil
  })

  // Connect to the server
  s, err := gotalk.Connect("tcp", "localhost:"+port)
  if err != nil {
    panic(err)
  }
  fmt.Printf("client: connected to %s\n", s.Addr())

  // Send a notification as JSON
  fmt.Printf("client: sending 'msg' notification\n")
  if err := s.Notify("msg", struct{Msg string}{"World"}); err != nil {
    fmt.Printf("client: notification: %v\n", err.Error())
  }

  // Send a notification as byte string
  fmt.Printf("client: sending 'msg' notification\n")
  if err := s.BufferNotify("msg", []byte("Hello")); err != nil {
    fmt.Printf("client: notification: error %v\n", err.Error())
  }

  // Send a request & read result via JSON-encoded go values
  fmt.Printf("client: sending 'greet' request\n")
  greeting := GreetOut{}
  if err := s.Request("greet", GreetIn{"Rasmus"}, &greeting); err != nil {
    fmt.Printf("client: greet: error %v\n", err.Error())
  } else {
    fmt.Printf("client: greet: %+v\n", greeting)
  }

  // Send a request & read result as byte strings
  fmt.Printf("client: sending 'echo' request\n")
  b, err := s.BufferRequest("echo", []byte("abc"))
  if err != nil {
    fmt.Printf("client: echo: error %v\n", err.Error())
  } else {
    fmt.Printf("client: echo: %v\n", string(b))
  }

  s.Close()
}
