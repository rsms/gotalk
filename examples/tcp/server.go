package main
import (
  "fmt"
  "github.com/rsms/gotalk"
)

func server(port string) {
  // Use a separate set of handlers for the server, as we are running the client in the same
  // program and thus we would have both server and client register handlers on DefaultHandlers.
  handlers := gotalk.NewHandlers()

  // Handle JSON-encoded request & result
  handlers.Handle("greet", func(in GreetIn) (GreetOut, error) {
    fmt.Printf("server: handling 'greet' request: %+v\n", in)
    return GreetOut{"Hello " + in.Name}, nil
  })

  // Handle buffered request & result
  handlers.HandleBufferRequest("echo", func(_ *gotalk.Sock, _ string, b []byte) ([]byte, error) {
    fmt.Printf("server: handling 'echo' request: %q\n", string(b))
    return b, nil
  })

  // Handle all notifications
  handlers.HandleBufferNotification("", func(s *gotalk.Sock, name string, b []byte) {
    fmt.Printf("server: received notification: %q => %q\n", name, string(b))

    // Send a request to the other end.
    // Note that we must do this in a goroutine as we would otherwise block this function from
    // returning the response, meaning our client() function would block indefinitely on waiting
    // for a response.
    go func() {
      fmt.Printf("server: sending 'ping' request\n")
      reply, err := s.BufferRequest("ping", []byte("abc"))
      if err != nil {
        fmt.Printf("server: ping: error %v\n", err.Error())
      } else {
        fmt.Printf("server: ping: %v\n", string(reply))
      }
    }()
  })

  // Accept connections
  s, err := gotalk.Listen("tcp", "localhost:"+port)
  if err != nil {
    panic(err)
  }
  s.Handlers = handlers
  fmt.Printf("server: listening at %s\n", s.Addr())
  go s.Accept()
}
