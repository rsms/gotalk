// A minimal gotalk web app
package main

import (
	"github.com/rsms/gotalk"
	"net/http"
)

func main() {
	gotalk.Handle("echo", func(in string) (string, error) {
		return in, nil
	})
	http.Handle("/gotalk/", gotalk.WebSocketHandler())
	http.Handle("/", http.FileServer(http.Dir(".")))
	println("Listening on http://localhost:1234/")
	panic(http.ListenAndServe("localhost:1234", nil))
}
