# Gotalk in JavaScript

The JavaScript implementation of Gotalk currently only supports connecting over web sockets from modern web browsers. A future version might provide "listening" abilities and support for other environments like Nodejs.

Here's an example of connecting to a web server, providing a "prompt" operation (which the server can invoke at ay time to ask the user some question) and finally invoking a "echo" operation on the server.

```js
gotalk.handle('prompt', function (params, result) {
  var answer = prompt(params.question, params.placeholderAnswer);
  result(answer);
})

// open a connection
const c = gotalk.connection()

// do stuff when the connection state changes
c.on("open",  () => console.log("connection opened"))
c.on("close", () => console.log("connection closed"))

// make a request
let pong = await c.request("ping")
console.log("ping reply:", pong)
```

When using the `gotalk.js` file (e.g. in a web browser), the API is exposed as `window.gotalk`.
If the `gotalk` directory containing the Gotalk CommonJS module is used, the API is returned as
an object from `require('./gotalk')`.

See [`gotalk.d.ts`](gotalk.d.ts) for documentation.
