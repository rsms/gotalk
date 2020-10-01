/*
Gotalk is a complete muli-peer real-time messaging library.
See https://github.com/rsms/gotalk#readme for a more in-depth explanation of what Gotalk is
and what it can do for you.

Most commonly Gotalk is used for rich web app development, as an alternative to HTTP APIs, when
the web app mainly runs client-side rather than uses a traditional "request a new page" style.


WebSocket example


Here is an example of a minimal but fully functional web server with Gotalk over websocket:

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
		http.Handle("/gotalk/", gotalk.NewWebSocketServer())
		// mount a file server to handle all other requests
		http.Handle("/", http.FileServer(http.Dir(".")))
		panic(http.ListenAndServe("localhost:1234", nil))
	}

Here is a matching HTML document; a very basic web app:

	<!DOCTYPE HTML>
	<html lang="en">
		<head>
			<meta charset="utf-8">
			<!-- load the built-in JS library -->
			<script type="text/javascript" src="/gotalk/gotalk.js"></script>
		</head>
		<body style="white-space:pre;font-family:monospace"><button>Send request</button>
		<script>
		// create a connection (automatically reconnects as needed)
		let c = gotalk.connection()
			.on('open', async() => log(`connection opened\n`))
			.on('close', reason => log(`connection closed (reason: ${reason})\n`))
		// make out button send a request
		document.body.firstChild.onclick = async () => {
			let res = await c.requestp('test/message', 'hello ' + new Date())
			log(`reply: ${JSON.stringify(res, null, 2)}\n`)
		}
		function log(message) {
			document.body.appendChild(document.createTextNode(message))
		}
		</script>
		</body>
	</html>


API layers


Gotalk can be thought of as being composed by four layers:

	4. Request-response API with automatic Go data encoding/decoding
	3. Request management, connection management
	2. Transport (TCP, pipes, unix sockets, HTTP, WebSocket, etc)
	1. Framed & streaming byte-based message protocol

You can make use of only some parts. For example you could write and read structured data in files
using the message protocol and basic file I/O, or use the high-level request-response API with
some custom transport.


*/
package gotalk
