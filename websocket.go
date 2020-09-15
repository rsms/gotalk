package gotalk

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"golang.org/x/net/websocket"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type WebSocketServer struct {
	Limits
	Handlers *Handlers
	OnAccept SockHandler

	// Template value for accepted sockets. Defaults to 0 (no automatic heartbeats)
	HeartbeatInterval time.Duration

	// Template value for accepted sockets. Defaults to nil
	OnHeartbeat func(load int, t time.Time)

	Server *websocket.Server
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
		"Content-Type":  []string{"application/javascript; charset=utf-8"},
		"Cache-Control": []string{"public,max-age=300"}, // 5min
		"ETag":          []string{jslibETag},
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
		s.Server.ServeHTTP(w, r)
	}
}

// Handler that can be used with the http package
func WebSocketHandler() *WebSocketServer {
	server := &WebSocketServer{
		Limits:   DefaultLimits,
		Handlers: DefaultHandlers,
	}

	handler := func(ws *websocket.Conn) {
		s := NewSock(server.Handlers)
		ws.PayloadType = websocket.BinaryFrame // websocket.TextFrame;
		s.Adopt(ws)
		if err := s.Handshake(); err != nil {
			s.Close()
		} else {
			if server.OnAccept != nil {
				server.OnAccept(s)
			}
			s.HeartbeatInterval = server.HeartbeatInterval
			s.OnHeartbeat = server.OnHeartbeat
			s.Read(server.Limits)
		}
	}

	server.Server = &websocket.Server{Handler: handler, Handshake: checkOrigin}

	return server
}

func checkOrigin(config *websocket.Config, req *http.Request) (err error) {
	config.Origin, err = websocket.Origin(config, req)
	if err == nil && config.Origin == nil {
		return fmt.Errorf("null origin")
	}
	return err
}
