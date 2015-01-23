package main
import (
  "encoding/json"
  "net/http"
  "time"
  "github.com/rsms/gotalk"
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


func acceptor(port string) {

  // Handle streaming request & result
  gotalk.HandleStreamRequest(
    "joke",
    func(s *gotalk.Sock, op string, in chan []byte, result gotalk.StreamWriter) error {
      // In a real-world application accepting connections from anyone on the internet, we
      // might want to schedule a timeout here, in case we never receive an "end stream" message. 

      // Read streaming request payloads
      for b := <-in; b != nil; b = <-in {
        println("joke request payload: \"" + string(b) + "\"")
      }

      // Send three jokes as three separate stream payloads
      for i := 0; i < 3; i++ {
        // Wait a little while in between sending jokes
        if i > 0 { time.Sleep(time.Second) }
        joke, err := fetchRandomJoke()
        if err != nil {
          return err
        }
        if err := result([]byte(joke)); err != nil {
          return err
        }
      }

      return nil
    })

  // Accept connections
  s, err := gotalk.Listen("tcp", "localhost:"+port)
  if err != nil {
    panic(err)
  }
  s.StreamReqLimit = 3  // Enable streaming requests
  println("listening at", s.Addr())
  go s.Accept()
}
