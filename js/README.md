# Gotalk in JavaScript

The JavaScript implementation of Gotalk currently only supports connecting over web sockets from modern web browsers. A future version might provide "listening" abilities and support for other environments like Nodejs.

Here's an example of connecting to a web server, providing a "prompt" operation (which the server can invoke at ay time to ask the user some question) and finally invoking a "echo" operation on the server.

```js
gotalk.handle('prompt', function (params, result) {
  var answer = prompt(params.question, params.placeholderAnswer);
  result(answer);
});

gotalk.connection().on('open', function (err, s) {
  if (err) return console.error(err);
  s.request('echo', function(err, result) {
    alert('echo returned:', result);
  });
});
```

# API

When using the `gotalk.js` file (e.g. in a web browser), the API is exposed as `window.gotalk`. If the `gotalk` directory containing the Gotalk CommonJS module is used, the API is returned as an object from `require('./gotalk')`.

## gotalk

```go
// Connect to `addr`, invoking `cb` when either the connection is open and
// ready to be used, or with an error if connection failed.
connect(addr string, cb function(Error, Sock)) ➝ Sock
```

```go
// Default `Handlers` utilized by the module-level `handle*` functions
defaultHandlers ➝ Handlers

// Default address to connect to. This is falsey if the JS library isn't served
// by gotalk.
defaultResponderAddress ➝ string
```

```go
// Open a connection to a gotalk responder.
// If `addr` is not provided, `defaultResponderAddress` is used.
// Equivalent to `Sock(defaultHandlers).open(addr, cb)`
open([addr string], [cb function(Error, Sock)]) ➝ Sock

// Start a persistent (keep-alive) connection to a gotalk responder.
// If `addr` is not provided, `defaultResponderAddress` is used.
// Equivalent to `Sock(defaultHandlers).openKeepAlive(addr)`
connection([addr string]) ➝ Sock
```

```go
// Convenience "shortcuts" to `defaultHandlers`
handle(op string, Handlers.ReqValueHandler)
handleNotification(name string, Handlers.NotValueHandler)
handleBufferRequest(op string, Handlers.ReqBufferHandler)
handleBufferNotification(name string, Handlers.NotBufferHandler)
```


## Sock

A connection over which gotalk can be spoken.

```go
type Sock prototypeof EventEmitter {
  handlers ➝ Handlers    // default: defaultHandlers
  protocol ⇄ ProtocolImp // default: protocol.binary

  // Open a connection to a gotalk responder.
  // If `addr` is not provided, `defaultResponderAddress` is used.
  open([addr string], [cb function(Error, Sock)]) ➝ Sock

  // Start a persistent (keep-alive) connection to a gotalk responder.
  // If `addr` is not provided, `defaultResponderAddress` is used.
  // Because the "open" step is abstracted away, this function does not accept
  // any "open callback". You should listen to the "open" and "close" events
  // instead.
  // The Sock will stay connected, and reconnect as needed, until you call `end()`.
  openKeepAlive([addr string]) ➝ Sock


  // Send request for operation `op` with `value` as the payload, using JSON
  // for encoding.
  request(op string, value any, cb function(Error, result any))

  // Send a request for operation `op` with raw-buffer `buf` as the payload,
  // if any. The type of result depends on the protocol used by the server
  // — a server sending a "text" frame means the result is a string, while a
  // server sending a "binary" frame causes the result to be a Buf.
  bufferRequest(op string,
                buf Buf|string|null,
                cb function(Error, result Buf|string))

  // Create a StreamRequest for operation `op` which is ready to be used.
  // Note that calling this method does not send any data — sending the request
  // and reading the response is performed by using the returned object.
  streamRequest(op string) ➝ StreamRequest

  // Send notification `name` with raw-buffer `buf` as the payload, if any.
  bufferNotify(name string, buf Buf|string|null)

  // Send notification `name` with `value`, using JSON for encoding.
  notify(name string, value any)

  // Send a heartbeat message with `load` which should be in the range [0-1]
  sendHeartbeat(load float)


  // Returns a string representing the address to which the socket is connected.
  address() ➝ string|null


  // Adopt a connection capable of being received from, written to and closed.
  // It should be in an "OPEN" ready-state.
  // You need to call `handshake` followed by `startReading` after adopting a previosuly
  // unadopted connection.
  // Throws an error if the provided connection type is not supported.
  // Currently only supports WebSocket.
  adopt(c Conn)

  // Perform protocol handshake.
  handshake()

  // Schedule reading from the underlying connection. Should only be called
  // once per connection.
  startReading()


  // Close the socket. If there are any outstanding responses from pending
  // requests, the socket will close when all pending requests has finished.
  // If you call this function a second time, the socket will close immediately,
  // even if there are outstanding responses.
  end()

  // Event emitted when the connection has opened.
  event "open" ()

  // Event emitted when the connection has closed. If it closed because of an
  // error, the error argument is non-falsey.
  event "close" (Error)

  event "heartbeat" ({time: Date, load: float})
}
```

```go
// Create a new socket with `handlers`.
Sock(Handlers=defaultHandlers) ➝ Sock
```

## StreamRequest

Represents a streaming request.

Response(s) arrive by the "data"(buf) event. When the response is complete,
a "end"(error) event is emitted, where error is non-empty if the request failed.

```go
type StreamRequest {
  op ➝ string  // Operation name
  id ➝ string  // Request ID

  // Write a request chunk. Writing an empty `buf` or null causes the request to end,
  // meaning no more chunks can be written. Calling `write()` or `end()` after the
  // request has finished has no effect.
  write(buf Buf|string)

  // End the request, indicating to the responder that it will not receive more payloads.
  end()

  // Event emitted when a response chunk is received.
  // Depending on the underlying transport, the argument is either a Buf or a string
  event "data" (Buf|string)

  // Event emitted when the response has ended.
  // If it ended because of an error, the argument is non-falsy.
  event "close" (Error)
}
```

```go
// Create a StreamRequest operating on a certain socket `s`.
// You should probably use `Sock.streamRequest()` instead, as it sets up response
// tracking, generates a request ID, etc.
StreamRequest(s Sock, op string, id string) ➝ StreamRequest
```

## Handlers

Container for mapping requests and notifications to code for handling them.

```go
type Handlers {
  // Callable for producing the results value of an operation.
  interface BufferResult(Buf|string) {
    // Callable for producing an error result
    error(Error)
  }

  // Like BufferResult, but accepts any value which will be encoded as JSON.
  interface ValueResult(any) {
    error(Error)
  }

  // Signature for request handlers dealing with raw data
  interface ReqBufferHandler(Buf|string, BufferResult, op string)

  // Signature for request handlers dealing with JSON-decoded value
  interface ReqValueHandler(any, ValueResult, op string)

  // Signature for notification handlers dealing with raw data
  interface NotBufferHandler(Buf|string, name string)

  // Signature for request handlers dealing with JSON-decoded value
  interface NotValueHandler(any, name string)

  // Register a handler for an operation `op`. If `op` is the empty string the
  // handler will be registered as a "fallback" handler, meaning that if there are
  // no handlers registered for request "x", the fallback handler will be invoked.
  handleRequest(op string, ReqValueHandler)
  handleBufferRequest(op string, ReqBufferHandler)
  
  // Register a handler for notification `name`. Just as with request handlers,
  // registering a handler for the empty string means it's registered as the
  // fallback handler.
  handleNotification(name string, NotValueHandler)
  handleBufferNotification(name string, NotBufferHandler)

  // Find request and notification handlers
  findRequestHandler(op string) ➝ ReqBufferHandler|null
  findNotificationHandler(name string) ➝ NotBufferHandler|null
}
```

```go
// Create a new Handlers object
Handlers() ➝ Handlers
```

## protocol

Describes the gotalk protocol and provides functionality for encoding and decoding messages.

```go
protocol = {
  // The version of the protocol implementation
  Version = 1

  // Message type constants
  MsgTypeSingleReq     = byte('r')
  MsgTypeStreamReq     = byte('s')
  MsgTypeStreamReqPart = byte('p')
  MsgTypeSingleRes     = byte('R')
  MsgTypeStreamRes     = byte('S')
  MsgTypeErrorRes      = byte('E')
  MsgTypeRetryRes      = byte('e')
  MsgTypeNotification  = byte('n')
  MsgTypeHeartbeat     = byte('h')
  MsgTypeProtocolError = byte('f')

  // ProtocolError codes
  ErrorAbnormal    = 0
  ErrorUnsupported = 1
  ErrorInvalidMsg  = 2
  ErrorTimeout     = 3

  // Maximum value of a heartbeat's "load"
  HeartbeatMsgMaxLoad = 0xffff

  // Implements a byte-binary version of the gotalk protocol
  binary ProtocolImp<Buf>

  // Implements a JavaScript text version of the gotalk protocol
  text ProtocolImp<string>
}
```

Generic interface implemented by `protocol.binary` and `protocol.text`:

```go
type ProtocolImp<T> {
  // Produce a fixed-digit number for integer `n`
  makeFixnum(n int, digits int) ➝ T

  // protocol.Version as a T
  versionBuf T

  // Parse value as protocol version which is expected to have a length of 2.
  parseVersion(T) ➝ int

  // Parses a message from a T, which must not including any payload data.
  parseMsg(T) ➝ {t:int, id:T, name:string, size:int} | null

  // Create a T representing a message, not including any payload data.
  makeMsg(t int, id T|string, name string, payloadSize int) ➝ T
}
```

## Buf

Represents a fixed-length, mutable sequence of bytes. This type is used as the payload value when operating the "binary" (byte-wise) protocol. In a web browser, this is implemented as a native `UInt8Array` backed by an `ArrayBuffer`, while in Nodejs it's implemented in the standard-library `Buffer`.

```go
type Buf {
  // Access or assign byte value at `index`
  [index int] ⇄ uint8

  // Returns a new view into the same buffer
  slice(start int[, end int]) ➝ Buf

  // Copies data from a region of this buffer to a region in the target buffer.
  copy(target Buf[, targetStart[, sourceStart[, sourceEnd]]])

  // Returns a text version of the buffer by interpreting the buffer's bytes as
  // `encoding`. Providing optional `start` and `end` have the same effects as
  // calling `b.slice(start, end).toString()`. `encoding` defaults to "utf8".
  toString([encoding string[, start int[, end int]]]) ➝ string
}
```

```go
// Create a new buffer capable of storing `size` bytes, or by wrapping a native-
// type array.
Buf(size int) ➝ Buf
Buf(ArrayBuffer) ➝ Buf
Buf(Uint8Array) ➝ Buf

// Create a new buffer by encoding a string as `encoding` (defaults to "utf8".)
Buf.fromString(string[, encoding string]) ➝ Buf
```

```go
// Test if something is a Buf
Buf.isBuf(any) ➝ true|false
```
