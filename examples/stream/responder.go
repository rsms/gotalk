package main
import (
  "encoding/json"
  "net/http"
  "time"
  "github.com/rsms/gotalk"
  "fmt"
  "io"
  "errors"
)

// Fecth a random joke from an external API
func fetchRandomJoke() (string, error) {
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


func handleJoke(s *gotalk.Sock, op string, in chan []byte, out io.WriteCloser) error {
  // We give the entire request a total of 30 seconds to complete
  requestTimeout := time.After(time.Second * 30)

  // Read streaming request payloads with a read timeout of one second
  readloop: for {
    select {
    case b := <-in:
      if b == nil {
        // End of request stream
        break readloop
      }
      fmt.Printf("responder: received request payload: %q\n", string(b))
    case <-time.After(time.Second):
      return errors.New("read timeout")
    case <-requestTimeout:
      return errors.New("request timeout")
    }
  }

  // Send three jokes as three separate stream payloads
  for i := 0; i < 3; i++ {
    // Wait a little while in between sending jokes
    if i > 0 { time.Sleep(time.Second) }
    joke, err := fetchRandomJoke()
    if err != nil {
      return err
    }
    if _, err := io.WriteString(out, joke); err != nil {
      return err
    }
  }

  return nil
}


func responder(port string) {
  // Handle streaming request & result
  gotalk.HandleStreamRequest("joke", handleJoke)

  // Accept connections
  s, err := gotalk.Listen("tcp", "localhost:"+port)
  if err != nil {
    panic(err)
  }
  s.Limits = gotalk.NoLimits  // Remove any limits on streaming requests
  fmt.Printf("responder: listening at %s\n", s.Addr())
  go s.Accept()
}
