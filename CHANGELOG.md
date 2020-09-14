## v1.0.0
- Changelog start

## v1.0.1
- Adds Version constant (`gotalk.Version string`)

## v1.1.0
- Exposes net.Listener on Server.Listener (previously private)
- Adds ListenTLS and ListenTLSCustom functions for starting servers with TLS
- Adds ConnectTLS and Sock.ConnectTLS functions for connecting with TLS
- Adds Sock.ConnectReader function for connecting over any io.ReadWriteCloser
- Adds Server.EnableUnixSocketGC convenience helper function
- Adds TLSCertPool and TLSAddRootCerts functions for customizing default TLS config
- Semantic change to UNIX socket transport: gotalk no longer installs a signal handler
  when listening over a UNIX socket using the function `Listen("unix", ...)`.
  If you want to retain that functionality, explicitly call `Server.EnableUnixSocketGC`
  on the server returned by `Listen("unix", ...)`.
