// A simple example of two connected sockets communicating with eachother
package main
import (
  "log"
  "github.com/rsms/gotalk"
)


func requestGreet(s *gotalk.Sock, name string) string {
  r := ""
  if err := s.Request("greet", name, &r); err != nil {
    panic("request error: " + err.Error())
  }
  return r
}


func main() {
  // Create two connected sockets with default handlers and limits
  s1, s2, err := gotalk.Pipe(nil, nil)
  if err != nil {
    panic(err.Error())
  }

  // Give the sockets names so we can include it in the greetings
  s1.UserData = "socket#1"
  s2.UserData = "socket#2"

  // Handle greetings
  gotalk.Handle("greet", func(s *gotalk.Sock, name string) (string, error) {
    sockname, _ := s.UserData.(string)
    return "Hello " + name + " from " + sockname, nil
  })

  // Send a "greet" request to each socket, making the opposite side respond.
  log.Printf("greet(s1, \"Bob\")  => %+v\n", requestGreet(s1, "Bob"))
  log.Printf("greet(s2, \"Lisa\") => %+v\n", requestGreet(s2, "Lisa"))

  // Output
  //   greet(s1, "Bob")  => {Greeting:Hello Bob from socket#2}
  //   greet(s2, "Lisa") => {Greeting:Hello Lisa from socket#1}

  s1.Close()
  s2.Close()
}
