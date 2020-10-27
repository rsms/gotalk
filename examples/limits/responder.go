package main

import (
	"fmt"
	"time"

	"github.com/rsms/gotalk"
)

func responder(port string) {
	// A simple echo operation with a 500ms response delay
	gotalk.HandleBufferRequest("echo", func(s *gotalk.Sock, op string, buf []byte) ([]byte, error) {
		fmt.Printf("responder: handling request\n")
		time.Sleep(time.Millisecond * 10)
		return buf, nil
	})

	// Start a server
	s, err := gotalk.Listen("tcp", "localhost:"+port)
	if err != nil {
		panic(err)
	}

	// Limit this server to 5 concurrent requests and set really low wait times
	// to that this demo doesn't take forever to run.
	s.Limits = &gotalk.Limits{
		BufferRequests: 5,
		BufferMinWait:  10 * time.Millisecond,
		BufferMaxWait:  100 * time.Millisecond,
	}

	// Accept connections
	fmt.Printf("listening at %q\n", s.Addr())
	go s.Accept()
}
