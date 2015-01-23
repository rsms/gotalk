package main
import (
  "fmt"
  "log"
  "github.com/rsms/gotalk"
)

type GreetIn struct {
  Name string `json:"name"`
}
type GreetOut struct {
  Greeting string `json:"greeting"`
}

func main() {
  gotalk.HandleBufferRequest("ping", func(_ *gotalk.Sock, _ string, b []byte) ([]byte, error) {
    println("in ping handler: b=", string(b))
    return []byte("pong"), nil
  })

  s, err := gotalk.Connect("tcp", "localhost:1234")
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
  b, err := s.BufferRequest("echo", []byte("abc"))
  if err != nil {
    fmt.Printf("echo: %v\n", err.Error())
  } else {
    fmt.Printf("echo: %v\n", string(b))
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
