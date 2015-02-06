package main
import (
  "net/http"
  "github.com/rsms/gotalk"
)

func main() {
  gotalk.Handle("echo", func(in string) (string, error) {
    return in, nil
  })
  http.Handle("/gotalk/", gotalk.WebSocketHandler())
  http.Handle("/", http.FileServer(http.Dir(".")))
  err := http.ListenAndServe("localhost:1234", nil)
  if err != nil {
    panic(err)
  }
}
