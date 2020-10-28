package gotalk

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"golang.org/x/net/websocket"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// WebSocketConnection is an alias for the websocket connection type, to spare the use
// from having to import golang.org/x/net/websocket
type WebSocketConnection = websocket.Conn

// WebSocket is a type of gotalk.Sock used for web socket connections,
// managed by a WebSocketServer.
type WebSocket struct {
	Sock

	// A function to be called when the socket closes. See Socket.CloseHandler for details.
	CloseHandler func(s *WebSocket, code int)
}

// Conn returns the underlying web socket connection.
//
// Accessing the web socket connection inside a handler function:
// Handler functions can opt in to receive a pointer to a Sock but not a WebSocket.
// This makes handlers more portable, testable and the implementation becomes simpler.
// However, sometimes you might need to access the web socket connection anyhow:
//
//   gotalk.Handle("foo", func(s *gotalk.Sock, m FooMessage) error {
//     ws := s.Conn().(*gotalk.WebSocketConnection)
//     // do something with ws
//     return nil
//   })
//
func (s *WebSocket) Conn() *WebSocketConnection { return s.Sock.conn.(*WebSocketConnection) }

// Request returns the http request upgraded to the WebSocket
func (s *WebSocket) Request() *http.Request { return s.Conn().Request() }

// Context returns the http request's context. The returned context is never nil.
func (s *WebSocket) Context() context.Context { return s.Request().Context() }

// ---------------------------------------------------------------------------------

// WebSocketServer conforms to http.HandlerFunc and is used to serve Gotalk over HTTP or HTTPS
type WebSocketServer struct {
	// Handlers describe what this server is capable of responding to.
	// Initially set to gotalk.DefaultHandlers by NewWebSocketServer().
	//
	// Handler can be assigned a new set of handlers at any time.
	// Whenever a new socket is connected, it references the current value of Handlers, therefore
	// changes to Handlers has an effect for newly connected sockets only.
	*Handlers

	// Limits control resource limits.
	// Initially set to gotalk.DefaultLimits by NewWebSocketServer().
	*Limits

	// OnConnect is an optional handler to be invoked when a new socket is connected.
	// This handler is only called for sockets which passed the protocol handshake.
	// If you want to deny a connection, simply call s.Close() on the socket in this handler.
	//
	// Gotalk checks if Origin header is a valid URL by default but does nothing else in terms of
	// origin validation. You might want to verify s.Conn().Config().Origin in OnConnect.
	OnConnect func(s *WebSocket)

	// HeartbeatInterval is not used directly by WebSocketServer but assigned to every new socket
	// that is connected. The default initial value (0) means "no automatic heartbeats" (disabled.)
	//
	// Note that automatic heartbeats are usually not a good idea for web sockets for two reasons:
	//
	//   a) You usually want to keep as few connections open as possible; letting them time out is
	//      often desired (heartbeats prevent connection timeout.)
	//
	//   b) Automatic timeout uses more resources
	//
	HeartbeatInterval time.Duration

	// OnHeartbeat is an optional callback for heartbeat confirmation messages.
	// Not used directly by WebSocketServer but assigned to every new socket that is connected.
	OnHeartbeat func(load int, t time.Time)

	// Underlying websocket server (will become a function in gotalk 2)
	Server *websocket.Server

	// DEPRECATED use OnConnect instead
	OnAccept SockHandler

	// storage for underlying web socket server.
	// Server is a pointer to this for legacy reasons.
	// Gotalk <=1.1.5 allocated websocket.Server separately on the heap and assigned it to Server.
	server websocket.Server
}

// NewWebSocketServer creates a web socket server which is a http.Handler
func NewWebSocketServer() *WebSocketServer {
	s := &WebSocketServer{
		Handlers: DefaultHandlers,
		Limits:   DefaultLimits,
		server: websocket.Server{
			Handshake: checkOrigin,
		},
	}
	s.Server = &s.server // legacy API
	s.server.Handler = s.onAccept
	return s
}

// DEPREACTED use NewWebSocketServer
func WebSocketHandler() *WebSocketServer {
	return NewWebSocketServer()
}

func (s *WebSocketServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, gotalkJSSuffix) {
		if s.maybeReplyNotModified(w, r) {
			return
		}
		contentLength := jslibLen
		body := jslibBody
		header := w.Header()
		acceptEncoding := parseCommaStrSet(r.Header.Get("Accept-Encoding"))
		if _, ok := acceptEncoding["gzip"]; ok {
			header["Content-Encoding"] = []string{"gzip"}
			contentLength = jslibLenGzip
			body = jslibBodyGzip
		}
		header["Content-Length"] = contentLength
		for k, v := range jslibHeader {
			header[k] = v
		}
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	} else {
		// upgrade request connection to web socket protocol
		s.server.ServeHTTP(w, r)
	}
}

// ---------------------------------------------------------------------------------
// internal

// onAccept is called for new web socket connections
func (server *WebSocketServer) onAccept(ws *WebSocketConnection) {
	// Set the frame payload type of the web socket
	ws.PayloadType = websocket.BinaryFrame

	// Create a new gotalk socket of the WebSocket flavor
	sock := &WebSocket{
		Sock: Sock{
			Handlers:          server.Handlers,
			HeartbeatInterval: server.HeartbeatInterval,
			OnHeartbeat:       server.OnHeartbeat,
			conn:              ws,
		},
	}
	sock.Sock.CloseHandler = func(_ *Sock, code int) {
		if sock.CloseHandler != nil {
			sock.CloseHandler(sock, code)
		}
	}

	// Adopt the web socket
	sock.Adopt(ws)

	// perform protocol handshake
	if err := sock.Handshake(); err != nil {
		sock.Close()
		return
	}

	// Call optional OnConnect handler
	if server.OnConnect != nil {
		server.OnConnect(sock)
	} else if server.OnAccept != nil {
		// legacy deprecated callback
		server.OnAccept(&sock.Sock)
	}

	// If the OnConnect handler closed the connection, stop here
	if sock.conn == nil {
		return
	}

	// enter read loop
	sock.Read(server.Limits)
}

func checkOrigin(config *websocket.Config, req *http.Request) (err error) {
	config.Origin, err = websocket.Origin(config, req)
	if err == nil && config.Origin == nil {
		return fmt.Errorf("null origin")
	}
	return err
}

const gotalkJSSuffix = "/gotalk.js"
const gotalkJSMapSuffix = "/gotalk.js.map"

// js lib cache
var (
	jslibInitOnce sync.Once
	jslibETag     string
	jslibHeader   http.Header // alias map[string][]string
	jslibLen      []string
	jslibBody     []byte
	jslibLenGzip  []string
	jslibBodyGzip []byte
)

func init() {
	jslibETag = "\"" + JSLibSHA1Base64 + "\""
	jslibHeader = map[string][]string{
		"Content-Type":  {"application/javascript; charset=utf-8"},
		"Cache-Control": {"public,max-age=300"}, // 5min
		"ETag":          {jslibETag},
	}
	// Note on Cache-Control:
	// max-age is the max time a browser can hold on to a copy of gotalk.js without
	// checking back with an etag. I.e. it does not limit the age of the cached resource
	// but limits how often the browser revalidates. max-age should thus be relatively low
	// as a revalidation request will in most cases end early with a HTTP 304 Not Modified
	// response. 5min seems like a decent value.
	// TODO: Consider making max-age configurable.

	// uncompressed
	jslibBody = []byte(JSLibString)
	jslibLen = []string{strconv.FormatInt(int64(len(jslibBody)), 10)}

	// gzip
	var zbuf bytes.Buffer
	zw, _ := gzip.NewWriterLevel(&zbuf, gzip.BestCompression)
	_, err := zw.Write(jslibBody)
	if err2 := zw.Close(); err2 != nil && err == nil {
		err = err2
	}
	if err != nil {
		panic(err)
	}
	jslibBodyGzip = zbuf.Bytes()
	jslibLenGzip = []string{strconv.FormatInt(int64(len(jslibBodyGzip)), 10)}
}

func (s *WebSocketServer) maybeReplyNotModified(w http.ResponseWriter, r *http.Request) bool {
	reqETag := r.Header["If-None-Match"]
	if len(reqETag) != 0 && reqETag[0] == jslibETag {
		w.WriteHeader(http.StatusNotModified)
		return true
	}
	return false
}

func parseCommaStrSet(s string) map[string]struct{} {
	m := make(map[string]struct{})
	if len(s) == 0 {
		return m
	}
	var start int = 0
	var end int = -1
	for i, b := range s {
		if start == -1 {
			if b != ' ' {
				start = i
			}
		} else {
			switch b {
			case ' ':
				end = i
			case ',':
				if end == -1 {
					end = i
				}
				m[s[start:end]] = struct{}{}
				end = -1
				start = -1
				break
			}
		}
	}
	if start != -1 {
		end = len(s)
		m[s[start:end]] = struct{}{}
	}
	return m
}
