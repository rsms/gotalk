This example demonstrates how to use gotalk over raw TCP connections:

    go build && ./tcp

Running this program outputs something like this:

    server: listening at 127.0.0.1:1234
    client: connected to 127.0.0.1:1234
    client: sending 'msg' notification
    client: sending 'msg' notification
    client: sending 'greet' request
    server: received notification: "msg" => "{\"Msg\":\"World\"}"
    server: received notification: "msg" => "Hello"
    server: sending 'ping' request
    server: sending 'ping' request
    server: handling 'greet' request: {Name:Rasmus}
    client: handling 'ping' request: "abc"
    client: handling 'ping' request: "abc"
    client: greet: {Greeting:Hello Rasmus}
    client: sending 'echo' request
    server: ping: pong
    server: ping: pong
    server: handling 'echo' request: "abc"
    client: echo: abc
