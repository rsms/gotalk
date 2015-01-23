package main
import (
  "log"
  "github.com/rsms/gotalk"
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
  gotalk.Handle("greet", func(in GreetIn) (GreetOut, error) {
    println("in greet handler: in.Name=", in.Name)
    return GreetOut{"Hello " + in.Name}, nil
  })

  // Handle buffered request & result
  gotalk.HandleBufferRequest("echo", func(_ *gotalk.Sock, _ string, b []byte) ([]byte, error) {
    println("in echo handler: inbuf=", string(b))
    return b, nil
  })

  // Handle all notifications
  gotalk.HandleBufferNotification("", func(_ *gotalk.Sock, name string, b []byte) {
    log.Printf("got notification: \"%s\" => \"%v\"\n", name, string(b))
  })

  // Accept connections
  s, err := gotalk.Listen("tcp", "localhost:1234")
  if err != nil {
    log.Fatalln(err)
  }
  println("listening at", s.Addr())
  s.Accept()
}
