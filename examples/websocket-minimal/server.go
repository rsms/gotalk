// A minimal gotalk web app
package main

import (
	"github.com/rsms/gotalk"
	"net/http"
)

type Message struct {
	Author string
	Body   string
}

func main() {
	// This function handles requests for "test/message".
	gotalk.Handle("test/message", func(input string) (*Message, error) {
		// It can return any Go type. Here we return a structure and no error.
		return &Message{Author: "Bob", Body: input}, nil
	})

	// mount Gotalk at "/gotalk/"
	http.Handle("/gotalk/", gotalk.WebSocketHandler())

	// mount a file server to handle all other requests
	http.Handle("/", http.FileServer(http.Dir(".")))

	println("Listening on http://localhost:1234/")
	panic(http.ListenAndServe("localhost:1234", nil))
}
