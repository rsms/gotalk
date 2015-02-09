package main

import (
  "errors"
  "sync"
  "net/http"
  "encoding/json"
  "math/rand"
  "time"
  "io/ioutil"
  "github.com/rsms/gotalk"
)

type Room struct {
  Name     string `json:"name"`
  mu       sync.RWMutex
  messages []*Message
}

func (room *Room) appendMessage(m *Message) {
  room.mu.Lock()
  defer room.mu.Unlock()
  room.messages = append(room.messages, m)
}

type Message struct {
  Author string `json:"author"`
  Body   string `json:"body"`
}

type NewMessage struct {
  Room    string  `json:"room"`
  Message Message `json:"message"`
}

type RoomMap map[string]*Room

var (
  rooms   RoomMap
  roomsmu sync.RWMutex
  socks   map[*gotalk.Sock]int
  socksmu sync.RWMutex
)

func onAccept(s *gotalk.Sock) {
  // Keep track of connected sockets
  socksmu.Lock()
  defer socksmu.Unlock()
  socks[s] = 1

  s.CloseHandler = func (s *gotalk.Sock, _ int) {
    socksmu.Lock()
    defer socksmu.Unlock()
    delete(socks, s)
  }

  // Send list of rooms
  roomsmu.RLock()
  defer roomsmu.RUnlock()
  s.Notify("rooms", rooms)

  // Assign the socket a random username
  username := randomName()
  s.UserData = username
  s.Notify("username", username)
}

func broadcast(name string, in interface{}) {
  socksmu.RLock()
  defer socksmu.RUnlock()
  for s, _ := range socks {
    s.Notify(name, in)
  }
}

func findRoom(name string) *Room {
  roomsmu.RLock()
  defer roomsmu.RUnlock()
  return rooms[name]
}

func createRoom(name string) *Room {
  roomsmu.Lock()
  defer roomsmu.Unlock()
  room := rooms[name]
  if room == nil {
    room = &Room{Name:name}
    rooms[name] = room
    broadcast("rooms", rooms)
  }
  return room
}

// Instead of asking the user for her/his name, we randomly assign one
var names struct {
  First []string
  Last  []string
}

func randomName() string {
  first := names.First[rand.Intn(len(names.First))]
  return first
  // last := names.Last[rand.Intn(len(names.Last))][:1]
  // return first + " " + last
}

func main() {
  socks = make(map[*gotalk.Sock]int)
  rooms = make(RoomMap)

  // Load names data
  if namesjson, err := ioutil.ReadFile("names.json"); err != nil {
    panic("failed to read names.json: " + err.Error())
  } else if err := json.Unmarshal(namesjson, &names); err != nil {
    panic("failed to read names.json: " + err.Error())
  }
  rand.Seed(time.Now().UTC().UnixNano())

  // Add some sample rooms and messages
  createRoom("animals").appendMessage(
    &Message{randomName(),"I like cats"})
  createRoom("jokes").appendMessage(
    &Message{randomName(),"Two tomatoes walked across the street ..."})
  createRoom("golang").appendMessage(
    &Message{randomName(),"func(func(func(func())func()))func()"})

  // Register our handlers
  gotalk.Handle("list-messages", func(roomName string) ([]*Message, error) {
    room := findRoom(roomName)
    if room == nil {
      return nil, errors.New("no such room")
    }
    return room.messages, nil
  })

  gotalk.Handle("send-message", func(s *gotalk.Sock, r NewMessage) error {
    if len(r.Message.Body) == 0 {
      return errors.New("empty message")
    }
    username, _ := s.UserData.(string)
    room := findRoom(r.Room)
    room.appendMessage(&Message{username, r.Message.Body})
    r.Message.Author = username
    broadcast("newmsg", &r)
    return nil
  })

  gotalk.Handle("create-room", func(name string) (*Room, error) {
    if len(name) == 0 {
      return nil, errors.New("empty name")
    }
    return createRoom(name), nil
  })

  // Serve gotalk at "/gotalk/"
  gotalkws := gotalk.WebSocketHandler()
  gotalkws.OnAccept = onAccept
  http.Handle("/gotalk/", gotalkws)

  http.Handle("/", http.FileServer(http.Dir(".")))
  err := http.ListenAndServe(":1235", nil)
  if err != nil {
    panic("ListenAndServe: " + err.Error())
  }
}
