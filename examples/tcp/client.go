package main
import (
  "fmt"
  "log"
  . "github.com/rsms/gotalk"
)

type GreetIn struct {
  Name string `json:"name"`
}
type GreetOut struct {
  Greeting string `json:"greeting"`
}

func main() {
  HandleBufferRequest("ping", func(s Sock, op string, inbuf []byte) ([]byte, error) {
    println("in ping handler: inbuf=", string(inbuf))
    return []byte("pong"), nil
  })

  s, err := Connect("tcp", "localhost:1234")
  if err != nil {
    log.Fatalln(err)
  }
  println("connected to", s.Addr())

  // Send a request & read result via JSON-encoded go values
  greeting := GreetOut{}
  if err := s.Request("greet", GreetIn{"Rasmus"}, &greeting); err != nil {
    fmt.Printf("greet: %v\n", err.Error())
  } else {
    fmt.Printf("greet: %+v\n", greeting)
  }

  // Send a request & read result as byte strings
  outbuf, err := s.BufferRequest("echo", []byte("abc"))
  if err != nil {
    fmt.Printf("echo: %v\n", err.Error())
  } else {
    fmt.Printf("echo: %v\n", string(outbuf))
  }

  // Send a notification as JSON
  if err := s.Notify("msg", "World"); err != nil {
    fmt.Printf("echo: %v\n", err.Error())
  }

  // Send a notification as byte string
  if err := s.BufferNotify("msg", []byte("Hello")); err != nil {
    fmt.Printf("echo: %v\n", err.Error())
  }

  s.Close()
}
