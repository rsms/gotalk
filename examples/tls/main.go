package main
import "os"

// Describes the request parameter and operation result types for our "greet" operation
type GreetIn struct {
  Name string `json:"name"`
}

type GreetOut struct {
  Greeting string `json:"greeting"`
}

func main() {
	mode := "demo"
	if len(os.Args) > 1 {
		mode = os.Args[1]
		if len(os.Args) > 2 || (mode != "client" && mode != "server") {
			println("usage:")
			println("  " + os.Args[0] + "         Run full demo and exit")
			println("  " + os.Args[0] + " server  Run demo server")
			println("  " + os.Args[0] + " client  Run demo client and exit")
			println("  " + os.Args[0] + " -h      Print help and exit")
			os.Exit(1)
		}
	}

  port := "1234"

  if mode == "server" || mode == "demo" {
	  server(port)
	}

  if mode == "client" || mode == "demo" {
  	client(port)
  } else {
	  <- make(chan bool)  // wait forever, until ^C
	}
}
