package main

import (
	"fmt"
	"github.com/rsms/gotalk"
	"time"
)

func responder(port string) {
	// A simple echo operation with a 500ms response delay
	gotalk.HandleBufferRequest("echo", func(s *gotalk.Sock, op string, buf []byte) ([]byte, error) {
		fmt.Printf("responder: handling request\n")
		time.Sleep(time.Millisecond * 400)
		return buf, nil
	})

	// Start a server
	s, err := gotalk.Listen("tcp", "localhost:"+port)
	if err != nil {
		panic(err)
	}

	// Limit this server to 5 concurrent requests (and disable streaming messages)
	s.Limits = gotalk.NewLimits(5, 0)

	// Accept connections
	fmt.Printf("listening at %q\n", s.Addr())
	go s.Accept()
}
