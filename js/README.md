# Gotalk in JavaScript

The JavaScript implementation of Gotalk currently only supports connecting over web sockets from modern web browsers. A future version might provide "listening" abilities and support for other environments like Nodejs.

Here's an example of connecting to a web server, providing a "prompt" operation (which the server can invoke at ay time to ask the user some question) and finally invoking a "echo" operation on the server.

```js
gotalk.handle('prompt', function (params, result) {
  var answer = prompt(params.question, params.placeholderAnswer);
  result(answer);
});

gotalk.connect('ws://'+document.location.host+'/gotalk', function (err, s) {
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
connect(addr String, cb function(Error, Sock)) ➝ Sock
```

```go
// Default `Handlers` utilized by the module-level `handle*` functions
defaultHandlers ➝ Handlers
```

```go
// Convenience "shortcuts" to `defaultHandlers`
handle(op String, Handlers.ReqValueHandler)
handleNotification(name String, Handlers.NotValueHandler)
handleBufferRequest(op String, Handlers.ReqBufferHandler)
handleBufferNotification(name String, Handlers.NotBufferHandler)
```

## Sock

A connection over which gotalk can be spoken.

```go
type Sock prototypeof EventEmitter {
  handlers ➝ Handlers    // default: defaultHandlers
  protocol ⇄ ProtocolImp // default: protocol.binary

  // Adopt a Web Socket, which should be in an "OPEN" ready-state. You need to
  // call `handshake` followed by `startReading` after adopting a web socket.
  adoptWebSocket(WebSocket)

  // Perform protocol handshake.
  handshake()

  // Schedule reading from the underlying connection. Should only be called
  // once per connection.
  startReading()

  // Send a request for operation `op` with raw-buffer `buf` as the payload,
  // if any. The type of result depends on the protocol used by the server
  // — a server sending a "text" frame means the result is a String, while a
  // server sending a "binary" frame causes the result to be a Buf.
  bufferRequest(op String,
                buf String|Buf|null,
                cb function(Error, result Buf|String))

  // Send request for operation `op` with `value` as the payload, using JSON
  // for encoding.
  request(op String, value any, cb function(Error, result any))

  // Send notification `name` with raw-buffer `buf` as the payload, if any.
  bufferNotify(name String, buf String|Buf|null)

  // Send notification `name` with `value`, using JSON for encoding.
  notify(name String, value any)

  // Returns a string representing the address to which the socket is connected.
  address() ➝ String | null

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
}
```

```go
// Create a new socket with `handlers`.
Sock(Handlers=defaultHandlers) ➝ Sock
```

## Handlers

Container for mapping requests and notifications to code for handling them.

```go
type Handlers {
  // Callable for producing the results value of an operation.
  interface BufferResult(Buf|String) {
    // Callable for producing an error result
    error(Error)
  }

  // Like BufferResult, but accepts any value which will be encoded as JSON.
  interface ValueResult(any) {
    error(Error)
  }

  // Signature for request handlers dealing with raw data
  interface ReqBufferHandler(Buf|String, BufferResult, op String)

  // Signature for request handlers dealing with JSON-decoded value
  interface ReqValueHandler(any, ValueResult, op String)

  // Signature for notification handlers dealing with raw data
  interface NotBufferHandler(Buf|String, name String)

  // Signature for request handlers dealing with JSON-decoded value
  interface NotValueHandler(any, name String)

  // Register a handler for an operation `op`. If `op` is the empty string the
  // handler will be registered as a "fallback" handler, meaning that if there are
  // no handlers registered for request "x", the fallback handler will be invoked.
  handleRequest(op String, ReqValueHandler)
  handleBufferRequest(op String, ReqBufferHandler)
  
  // Register a handler for notification `name`. Just as with request handlers,
  // registering a handler for the empty string means it's registered as the
  // fallback handler.
  handleNotification(name String, NotValueHandler)
  handleBufferNotification(name String, NotBufferHandler)

  // Find request and notification handlers
  findRequestHandler(op String) ➝ ReqBufferHandler|null
  findNotificationHandler(name String) ➝ NotBufferHandler|null
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
  Version int

  // Message type constants
  MsgTypeSingleReq      int
  MsgTypeStreamReq      int
  MsgTypeStreamReqPart  int
  MsgTypeSingleRes      int
  MsgTypeStreamRes      int
  MsgTypeErrorRes       int
  MsgTypeNotification   int

  // Implements a byte-binary version of the gotalk protocol
  binary ProtocolImp<Buf>

  // Implements a JavaScript text version of the gotalk protocol
  text ProtocolImp<String>
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
  parseMsg(T) ➝ {t:int, id:T, name:String, size:int} | null

  // Create a T representing a message, not including any payload data.
  makeMsg(t int, id T|String, name String, payloadSize int) ➝ T
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
  toString([encoding String[, start int[, end int]]]) ➝ Buf
}
```

```go
// Create a new buffer capable of storing `size` bytes, or by wrapping a native-
// type array.
Buf(size int) ➝ Buf
Buf(ArrayBuffer) ➝ Buf
Buf(Uint8Array) ➝ Buf

// Create a new buffer by encoding a string as `encoding` (defaults to "utf8".)
Buf.fromString(String[, encoding String]) ➝ Buf
```

```go
// Test if something is a Buf
Buf.isBuf(any) ➝ true|false
```
