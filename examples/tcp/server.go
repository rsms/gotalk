package main
import (
  "encoding/json"
  "log"
  "net/http"
  "time"
  . "github.com/rsms/gotalk"
)

// Fecth a random joke from an external API
func FetchRandomJoke() (string, error) {
  r, err := http.Get("http://api.icndb.com/jokes/random")
  if err != nil {
    return "", err
  }
  defer r.Body.Close()
  var m struct {
    Value struct {
      Joke string `json:"joke"`
    } `json:"value"`
  }
  json.NewDecoder(r.Body).Decode(&m)
  return m.Value.Joke, nil
}

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

  // Handle streaming request & result
  HandleStreamRequest("joke", func(s Sock, op string, in chan []byte, write StreamWriter) error {
    println("in joke read handler")

    // Read any request payloads
    for b := <-in; b != nil; b = <-in {
      println("joke request payload: \"" + string(b) + "\"")
    }

    // Send three jokes as three separate stream payloads
    for i := 0; i < 3; i++ {
      // Wait a little while in between sending jokes
      if i > 0 { time.Sleep(time.Second) }
      joke, err := FetchRandomJoke()
      if err != nil {
        return err
      }
      if err := write([]byte(joke)); err != nil {
        return err
      }
    }

    return nil
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
  println("listening at", s.Addr().String())
  s.Accept(nil)
}
