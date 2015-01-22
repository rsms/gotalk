This example operates over raw TCP connections.

The two programs `server.go` and `client.go` are supposed to be used together:

Terminal 1/2:

    go build -o server server.go && ./server

Terminal 2/2:

    go build -o client client.go && ./client
