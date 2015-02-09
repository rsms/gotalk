package main
import (
  "fmt"
  "net/http"
  "github.com/rsms/gotalk"
)

type GreetIn struct {
  Name string `json:"name"`
}

type GreetOut struct {
  Greeting string `json:"greeting"`
}

func onAccept(s *gotalk.Sock) {
  s.Notify("hello", "world")
  go func(){
    // Send a request & read result via JSON-encoded go values.
    greeting := GreetOut{}
    if err := s.Request("greet", GreetIn{"Rasmus"}, &greeting); err != nil {
      fmt.Printf("greet request failed: " + err.Error())
    } else {
      fmt.Printf("greet: %+v\n", greeting)
    }
  }()
}

func main() {
  gotalk.Handle("greet", func(in GreetIn) (GreetOut, error) {
    println("in greet handler: in.Name=", in.Name)
    return GreetOut{"Hello " + in.Name}, nil
  })

  gotalk.HandleBufferRequest("echó", func(s *gotalk.Sock, op string, b []byte) ([]byte, error) {
    return b, nil
  })

  ws := gotalk.WebSocketHandler()
  ws.OnAccept = onAccept
  http.Handle("/gotalk/", ws)
  http.Handle("/", http.FileServer(http.Dir(".")))
  err := http.ListenAndServe(":1234", nil)
  if err != nil {
    panic(err)
  }
}
