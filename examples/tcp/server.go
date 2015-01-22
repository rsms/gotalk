package main
import (
  "log"
  . "github.com/rsms/gotalk"
)

// Describes the request parameter and operation result types for our "greet" operation
type GreetIn struct {
  Name string `json:"name"`
}

type GreetOut struct {
  Greeting string `json:"greeting"`
}

func main() {
  // Handle JSON-encoded request & result
  Handle("greet", func(in GreetIn) (GreetOut, error) {
    println("in greet handler: in.Name=", in.Name)
    return GreetOut{"Hello " + in.Name}, nil
  })

  // Handle buffered request & result
  HandleBufferRequest("echo", func(s Sock, op string, inbuf []byte) ([]byte, error) {
    println("in echo handler: inbuf=", string(inbuf))
    return inbuf, nil
  })

  // Handle all notifications
  HandleBufferNotification("", func(s Sock, name string, inbuf []byte) {
    log.Printf("got notification: \"%s\" => \"%v\"\n", name, string(inbuf))
  })

  // Accept connections
  s, err := Listen("tcp", "localhost:1234")
  if err != nil {
    log.Fatalln(err)
  }
  println("listening at", s.Addr())
  s.Accept(nil)
}
