package gotalk

import (
  "golang.org/x/net/websocket"
)

// Handler that can be used with the http package
func WebSocketHandler(h *Handlers, handler SockHandler) websocket.Handler {
  if h == nil {
    h = DefaultHandlers
  }
  return websocket.Handler(
    func (ws *websocket.Conn) {
      s := NewSock(h)
      ws.PayloadType = websocket.BinaryFrame; // websocket.TextFrame;
      s.Adopt(ws)
      if err := s.Handshake(); err != nil {
        s.Close()
      } else {
        if handler != nil {
          handler(s)
        }
        s.Read()
      }
    })
}
