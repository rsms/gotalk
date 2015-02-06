package gotalk
import (
  "testing"
  "bytes"
  "io"
)

func assertMsgEqual(t *testing.T, msg []byte, expect []byte) {
  if !bytes.Equal(msg, expect) {
    t.Errorf("got msg \"%s\", expected \"%s\"", string(msg), string(expect))
  }
}


func TestMakeMsg(t *testing.T) {
  assertMsgEqual(t, MakeMsg(MsgTypeSingleReq, "abcd", "echo", 0, 0),  []byte("rabcd004echo00000000"))
  assertMsgEqual(t, MakeMsg(MsgTypeSingleReq, "abcd", "echo", 0, 3),  []byte("rabcd004echo00000003"))
  assertMsgEqual(t, MakeMsg(MsgTypeStreamReq, "abcd", "echo", 0, 3),  []byte("sabcd004echo00000003"))
  assertMsgEqual(t, MakeMsg(MsgTypeStreamReqPart, "abcd", "", 0, 3),  []byte("pabcd00000003"))
  assertMsgEqual(t, MakeMsg(MsgTypeSingleRes, "abcd", "", 0, 3),      []byte("Rabcd00000003"))
  assertMsgEqual(t, MakeMsg(MsgTypeStreamRes, "abcd", "", 0, 3),      []byte("Sabcd00000003"))
  assertMsgEqual(t, MakeMsg(MsgTypeErrorRes, "abcd", "", 0, 3),       []byte("Eabcd00000003"))
  assertMsgEqual(t, MakeMsg(MsgTypeRetryRes, "abcd", "", 6, 3),       []byte("eabcd0000000600000003"))
  assertMsgEqual(t, MakeMsg(MsgTypeNotification, "", "hello", 0, 3), []byte("n005hello00000003"))
}


func TestWriteReadVersion(t *testing.T) {
  s := new(bytes.Buffer)

  if n, err := WriteVersion(s); err != nil {
    t.Errorf("WriteVersion() failed: %v", err.Error())
  } else if n != 2 {
    t.Errorf("WriteVersion() => (%v, _), expected (2, _)", n)
  }

  if ProtocolVersion != 1 { t.Errorf("ProtocolVersion = %v, expected 1", ProtocolVersion) }
  if s.String() != "01" { t.Errorf("s.String() = %s, expected \"01\"", s.String()) }

  if u, err := ReadVersion(s); err != nil {
    t.Errorf("ReadVersion() failed: %v", err.Error())
  } else if u != ProtocolVersion {
    t.Errorf("ReadVersion() => (%v, _), expected (%v, _)", u, ProtocolVersion)
  }
}


type testMsg struct {
  t          MsgType
  id, name   string
  wait, size int
}

const kPayloadByte = byte('~')


func assertWriteMsg(t *testing.T, s *bytes.Buffer, m *testMsg) {
  msg := MakeMsg(m.t, m.id, m.name, m.wait, m.size)
  if _, err := s.Write(msg); err != nil {
    t.Errorf("s.Write() failed: %v", err.Error())
  }
  // Write ayload data
  if m.t != MsgTypeProtocolError {
    for i := 0; i != m.size; i++ {
      s.WriteByte(kPayloadByte)
    }
  }
  // t.Logf("s.String() => '%s'\n", s.String())
}


func assertReadMsg(t *testing.T, s *bytes.Buffer, m *testMsg) {
  ty, id, name, wait, size, err := ReadMsg(s)
  if err != nil && err != io.EOF {
    t.Fatalf("ReadMsg() failed: %v", err.Error())
  }
  if ty        != m.t {    t.Fatalf("ReadMsg() ty = \"%v\", expected \"%v\"", ty, m.t) }
  if id        != m.id {   t.Fatalf("ReadMsg() id = \"%v\", expected \"%v\"", id, m.id) }
  if name      != m.name { t.Fatalf("ReadMsg() name = \"%v\", expected \"%v\"", name, m.name) }
  if int(wait) != m.wait { t.Fatalf("ReadMsg() wait = %v, expected %v", wait, m.wait) }
  if int(size) != m.size { t.Fatalf("ReadMsg() size = %v, expected %v", size, m.size) }
  // Read payload
  if m.t != MsgTypeProtocolError {
    for i := 0; i != m.size; i++ {
      if b, err := s.ReadByte(); err != nil {
        t.Fatalf("ReadMsg() failed to read payload: %v", err.Error())
      } else if b != kPayloadByte {
        t.Fatalf("ReadMsg() => '%c', expected '%c'", b, kPayloadByte)
      }
    }
  }
}


func assertWriteReadMsg(t *testing.T, s *bytes.Buffer, m *testMsg) {
  assertWriteMsg(t,s,m)
  // println(s.String())
  assertReadMsg(t,s,m)
}


func TestWriteReadMsg(t *testing.T) {
  s := new(bytes.Buffer)

  // Make sure ReadMsg behaves correctly
  s.Write([]byte("n001a00000000"))
  assertReadMsg(t,s, &testMsg{MsgTypeNotification, "", "a", 0, 0})

  m := []*testMsg{
    &testMsg{MsgTypeSingleReq,     "abcd", "echo",    0, 0},
    &testMsg{MsgTypeSingleReq,     "zzzz", "lolcats", 0, 3},
    &testMsg{MsgTypeStreamReq,     "abcd", "echo",    0, 4},
    &testMsg{MsgTypeStreamReqPart, "abcd", "",        0, 5},
    &testMsg{MsgTypeSingleRes,     "abcd", "",        0, 6},
    &testMsg{MsgTypeStreamRes,     "abcd", "",        0, 7},
    &testMsg{MsgTypeErrorRes,      "abcd", "",        0, 8},
    &testMsg{MsgTypeRetryRes,      "abcd", "",        6, 8},
    &testMsg{MsgTypeNotification,  "",    "hello",    0, 9},
    &testMsg{MsgTypeProtocolError, "",    "",         0, int(ProtocolErrorInvalidMsg)},
  }

  // Serially (read, write, read, write, ...)
  for _, msg := range m {
    assertWriteReadMsg(t, s, msg)
  }

  // Interleaved (write, write, ..., read, read, ...)
  for _, msg := range m {
    assertWriteMsg(t, s, msg)
  }
  for _, msg := range m {
    assertReadMsg(t, s, msg)
  }
}
