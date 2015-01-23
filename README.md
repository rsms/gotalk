# gotalk

[Gotalk](https://github.com/rsms/gotalk) exists to make it easy for programs to *talk with one another over the internet*, like a web app coordinating with a web server, or a bunch of programs dividing work amongst eachother.

![A terribly boring amateur comic strip](https://github.com/rsms/gotalk/raw/master/doc/gotalk-comic.png)

Gotalk takes the natural approach of *bidirectional* and *concurrent* communication — any peer have the ability to expose "operations" as well as asking other peers to perform operations. The traditional restrictions of who can request and who can respond usually associated with a client-server model is nowhere to be found in gotalk.

## Gotalk in a nutshell

**Bidirectional** — There's no discrimination on capabilities depending on who connected or who accepted. Both "servers" and "clients" can expose operations as well as send requests to the other side.

**Concurrent** — Requests, results, and notifications all share a single connection without blocking eachother by means of [pipelining](http://en.wikipedia.org/wiki/Protocol_pipelining). There's no serialization on request-result or even for a single large message, as the gotalk protocol is frame-based and multiplexes messages over a single connection. This means you can perform several requests at once without having to think about queueing or blocking.

![Diagram of how Gotalk uses connection pipelining](https://github.com/rsms/gotalk/raw/master/doc/gotalk-pipeline-diagram.png)

**Simple** — Gotalk has a simple and opinionated API with very few components. You expose an operation via "handle" and send requests via "request".

**Debuggable** — The Gotalk protocol's wire format is ASCII-based for easy on-the-wire inspection of data. For example, here's a protocol message representing an operation request: `r001005hello00000005world`. The Gotalk protocol can thus be operated over any reliable byte transport.

**Practical** — Gotalk includes a JavaScript implementation for Web Sockets alongside the full-featured Go implementation, making it easy to build real-time web applications. The Gotalk source code also includes a number of easily-readable examples.


## By example

There are a few examples in the `examples` directory demonstrating Gotalk. But let's explore a simple program right now — here's a little something written in [Go](http://golang.org/) which demonstrates the use of an operation named "greet":

```go
func server() {
  gotalk.Handle("greet", func(in GreetIn) (GreetOut, error) {
    return GreetOut{"Hello " + in.Name}, nil
  })
  if err := gotalk.Serve("tcp", "localhost:1234"); err != nil {
    log.Fatalln(err)
  }
}

func client() {
  s, err := gotalk.Connect("tcp", "localhost:1234")
  if err != nil {
    log.Fatalln(err)
  }
  greeting := &GreetOut{}
  if err := s.Request("greet", GreetIn{"Rasmus"}, greeting); err != nil {
    log.Fatalln(err)
  }
  log.Printf("greeting: %+v\n", greeting)
  s.Close()
}
```

Let's look at the above example in more detail, broken apart to see what's going on.

We begin by importing the gotalk library together with `log` which we use for printing to the console:

```go
package main
import (
  "log"
  "github.com/rsms/gotalk"
)
```

We define two types: Expected input (request parameters) and output (result) for our "greet" operation:

```go
type GreetIn struct {
  Name string `json:"name"`
}
type GreetOut struct {
  Greeting string `json:"greeting"`
}
```

Registers a process-global request handler for an operation called "greet" accepting parameters of type `GreetIn`, returning results of type `GreetOut`:

```go
func server() {
  gotalk.Handle("greet", func(in GreetIn) (GreetOut, error) {
    return GreetOut{"Hello " + in.Name}, nil
  })
```

Finally at the bottom of our `server` function we call `gotalk.Serve`, which starts a local TCP server on port 1234:

```go
  if err := gotalk.Serve("tcp", "localhost:1234"); err != nil {
    log.Fatalln(err)
  }
}
```

In out `client` function we start by connecting to the server:

```go
func client() {
  s, err := gotalk.Connect("tcp", "localhost:1234")
  if err != nil {
    log.Fatalln(err)
  }
```

Finally we send a request for "greet" and print the result:

```go
  greeting := GreetOut{}
  if err := s.Request("greet", GreetIn{"Rasmus"}, &greeting); err != nil {
    log.Fatalln(err)
  }
  log.Printf("greeting: %+v\n", greeting)

  s.Close()
}
```

Output:

```go
greeting: {Greeting:Hello Rasmus}
```

## Gotalk in the web browser

Gotalk is implemented not only in the full-fledged Go package, but also in a JavaScript library. This allows writing web apps talking Gotalk via Web Sockets possible.

```go
// server.go:
package main
import (
  "net/http"
  "github.com/rsms/gotalk"
)
func main() {
  gotalk.HandleBufferRequest("echo", func(in []byte) ([]byte, error) {
    return in, nil
  })
  http.Handle("/gotalk", gotalk.WebSocketHandler(nil, nil))
  http.Handle("/", http.FileServer(http.Dir(".")))
  err := http.ListenAndServe(":1234", nil)
  if err != nil {
    panic("ListenAndServe: " + err.Error())
  }
}
```

In our html document, we begin by registering any operations we can handle:

```html
<!-- index.html -->
<body>
<script type="text/javascript" src="gotalk.js"></script>
<script>
gotalk.handle('greet', function (params, result) {
  result({ greeting: 'Hello ' + params.name });
});
</script>
```

We can't "listen & accept" connections in a web browser, but we can "connect" so we do just that, connecting to "/gotalk" which is where we registered `gotalk.WebSocketHandler` in our server.

```html
<!-- index.html -->
<body>
<script type="text/javascript" src="gotalk.js"></script>
<script>
gotalk.handle('greet', function (params, result) {
  result({ greeting: 'Hello ' + params.name });
});

gotalk.connect('ws://'+document.location.host+'/gotalk', function (err, s) {
  if (err) return console.error(err);
  // s is a gotalk.Sock
});
</script>
```

This is enough for enabling the *server* to do things in the *browser* ...

But you probably want to have the *browser* send requests to the *server*, so let's send a "echo" request just as our connection opens:

```js
gotalk.connect('ws://'+document.location.host+'/gotalk', function (err, s) {
  if (err) return console.error(err);
  s.request("echo", "Hello world", function (err, result) {
    if (err) return console.error('echo failed:', err);
    console.log('echo result:', result);
  });
});
```

We could rewrite our code like this to allow some UI component to send a request:

```js
var s = gotalk.connect('ws://'+document.location.host+'/gotalk', function (err, s) {
  if (err) return console.error(err);
});

button.addEventListener('click', function () {
  s.request("echo", "Hello world", function (err, result) {
    if (err) return console.error('echo failed:', err);
    console.log('echo result:', result);
  });
});
```

The request will fail with an error "socket is closed" if the user clicks our button while the connection isn't open.


## Protocol and wire format

The wire format is designed to be human-readable and flexible; it's byte-based and can be efficiently implemented in a number of environments ranging from HTTP and WebSocket in a web browser to raw TCP in Go or C. The protocol provides only a small set of operations on which more elaborate operations can be modeled by the user.

Here's a complete description of the protocol:

    conversation    = ProtocolVersion Message*
    message         = SingleRequest | StreamRequest
                    | SingleResult | StreamResult
                    | ErrorResult

    ProtocolVersion = <hexdigit> <hexdigit>

    SingleRequest   = "r" requestID operation payload
    StreamRequest   = "s" requestID operation payload StreamReqPart+
    StreamReqPart   = "p" requestID payload
    SingleResult    = "R" requestID payload
    StreamResult    = "S" requestID payload StreamResult*
    ErrorResult     = "E" requestID payload
    Notification    = "n" name payload

    requestID       = <byte> <byte> <byte>

    operation       = text3
    name            = text3

    text3           = text3Size text3Value
    text3Size       = hexUInt3
    text3Value      = <<byte>{text3Size} as utf8 text>

    payload         = payloadSize payloadData?
    payloadSize     = hexUInt8
    payloadData     = <byte>{payloadSize}

    hexUInt3        = <hexdigit> <hexdigit> <hexdigit>
    hexUInt8        = <hexdigit> <hexdigit> <hexdigit> <hexdigit>
                      <hexdigit> <hexdigit> <hexdigit> <hexdigit>

A conversation begins with the protocol version:

```lua
00  -- ProtocolVersion 0
```

If the version of the protocol spoken by the other end is not supported by the reader, the connection is terminated and the conversation never starts. Otherwise, any messages are read and/or written.

This is a "single-payload" request ...

```py
+----------------- SingleRequest
|  +---------------- requestID   "001"
|  |      +--------- operation   "echo"
|  |      |       +- payloadSize 25
|  |      |       |
r001004echo00000019{"message":"Hello World"}
```

... and a corresponding "single-payload" result:

```py
+----------------- SingleResult
|  +---------------- requestID   "001"
|  |       +-------- payloadSize 25
|  |       |
R00100000019{"message":"Hello World"}
```

Each request is identified by exactly three bytes—the `requestID`—which is requestor-specific and has no purpose beyond identity, meaning the value is never interpreted.

These "single" requests & results are the most common protocol messages, and as their names indicates, their payloads follow immediately after the header. For large payloads this can become an issue when dealing with many concurrent requests over a single connection, for which there's a more complicated "streaming" request & result type which we will explore later on.

If an error occurs while handing a request, an "error" is send as the reply instead of a regular result:

```py
+----------------- ErrorResult
|  +---------------- requestID   "001"
|  |       +-------- payloadSize 38
|  |       |
E00100000026{"error":"Unknown operation \"echo\""}
```

As with all messages, what the payload data represents is up to each application and not part of the Gotalk protocol although we use JSON in our examples here.

When there's no expectation on a response, Gotalk provides a "notification" message type:

```py
+---------------------- Notification
|              +--------- name        "chat message"
|              |       +- payloadSize 46
|              |       |
n00cchat message0000002e{"message":"Hi","from":"nthn","room":"gonuts"}
```

Notifications are never replied to nor can they cause "error" results.

For more complicated scenarios there are "streaming-payload" requests and results at our disposal. This allows transmitting of large amounts of data without the need for large buffers. For example this could be used to forward audio data to audio playback hardware, or to transmit a large file off of slow media like a tape drive or hard-disk drive.

Because transmitting a streaming request or result does not occupy "the line" (single-payloads are transmitted serially), they can also be useful when there are many concurrent requests happening over a single connection.

Here's an example of a "streaming-payload" request ...

```py
+----------------- StreamRequest
|  +---------------- requestID   "001"
|  |      +--------- operation   "echo"
|  |      |       +- payloadSize 11
|  |      |       |
s001004echo0000000b{"message":

+----------------- streamReqPart
|  +---------------- requestID   "001"
|  |       +-------- payloadSize 14
|  |       |
p0010000000e"Hello World"}

+----------------- streamReqPart
|  +---------------- requestID   "001"
|  |       +-------- payloadSize 0 (end of stream)
|  |       |
p00100000000
```

... followed by a "streaming-payload" result:

```py
+----------------- StreamResult (1st part)
|  +---------------- requestID   "001"
|  |       +-------- payloadSize 11
|  |       |
S0010000000b{"message":

+----------------- StreamResult (2nd part)
|  +---------------- requestID   "001"
|  |       +-------- payloadSize 14
|  |       |
S0010000000e"Hello World"}

+----------------- StreamResult
|  +---------------- requestID   "001"
|  |       +-------- payloadSize 0 (end of stream)
|  |       |
S00100000000
```

Requests and results does not need to match on the "single" vs "streaming" detail — it's perfectly fine to send a streaming request and read a single response, or send a single response just to receive a streaming result. *The payload type is orthogonal to the message type*, with the exception of an error response which is always a "single-payload" message, carrying any information about the error in its payload. Note however that the current version of the Go package does not provide a high-level API for mixed-kind request-response handling.


## Other implementations

- <https://github.com/gtaylor/python-gotalk>


## MIT license

Copyright (c) 2015 Rasmus Andersson <http://rsms.me/>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
