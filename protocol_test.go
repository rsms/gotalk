package gotalk
import (
  "testing"
  "bytes"
)

func assertMsgEqual(t *testing.T, msg []byte, expect []byte) {
  if !bytes.Equal(msg, expect) {
    t.Errorf("got msg \"%s\", expected \"%s\"", string(msg), string(expect))
  }
}


func TestMakeMsg(t *testing.T) {
  assertMsgEqual(t, MakeMsg(MsgTypeSingleReq, "abc", "echo", 0),  []byte("rabc004echo00000000"))
  assertMsgEqual(t, MakeMsg(MsgTypeSingleReq, "abc", "echo", 3),  []byte("rabc004echo00000003"))
  assertMsgEqual(t, MakeMsg(MsgTypeStreamReq, "abc", "echo", 3),  []byte("sabc004echo00000003"))
  assertMsgEqual(t, MakeMsg(MsgTypeStreamReqPart, "abc", "", 3),  []byte("pabc00000003"))
  assertMsgEqual(t, MakeMsg(MsgTypeSingleRes, "abc", "", 3),      []byte("Rabc00000003"))
  assertMsgEqual(t, MakeMsg(MsgTypeStreamRes, "abc", "", 3),      []byte("Sabc00000003"))
  assertMsgEqual(t, MakeMsg(MsgTypeErrorRes, "abc", "", 3),       []byte("Eabc00000003"))
  assertMsgEqual(t, MakeMsg(MsgTypeNotification, "", "hello", 3), []byte("n005hello00000003"))
}


func TestWriteReadVersion(t *testing.T) {
  s := new(bytes.Buffer)

  if n, err := WriteVersion(s); err != nil {
    t.Errorf("WriteVersion() failed: %v", err.Error())
  } else if n != 2 {
    t.Errorf("WriteVersion() => (%v, _), expected (2, _)", n)
  }

  if ProtocolVersion != 0 { t.Errorf("ProtocolVersion = %v, expected 0", ProtocolVersion) }
  if s.String() != "00" { t.Errorf("s.String() = %s, expected \"00\"", s.String()) }

  if u, err := ReadVersion(s); err != nil {
    t.Errorf("ReadVersion() failed: %v", err.Error())
  } else if u != ProtocolVersion {
    t.Errorf("ReadVersion() => (%v, _), expected (%v, _)", u, ProtocolVersion)
  }
}


type testMsg struct {
  t        MsgType
  id, name string
  size     int
}

const kPayloadByte = byte('~')


func assertWriteMsg(t *testing.T, s *bytes.Buffer, m *testMsg) {
  msg := MakeMsg(m.t, m.id, m.name, m.size)
  if _, err := s.Write(msg); err != nil {
    t.Errorf("s.Write() failed: %v", err.Error())
  }
  for i := 0; i != m.size; i++ {
    s.WriteByte(kPayloadByte)
  }
  //t.Logf("s.String() => '%s'\n", s.String())
}


func assertReadMsg(t *testing.T, s *bytes.Buffer, m *testMsg) {
  ty, id, name, size, err := ReadMsg(s)
  if err       != nil {    t.Errorf("ReadMsg() failed: %v", err.Error()) }
  if ty        != m.t {    t.Errorf("ReadMsg() ty = \"%v\", expected \"%v\"", ty, m.t) }
  if id        != m.id {   t.Errorf("ReadMsg() id = \"%v\", expected \"%v\"", id, m.id) }
  if name      != m.name { t.Errorf("ReadMsg() name = \"%v\", expected \"%v\"", name, m.name) }
  if int(size) != m.size { t.Errorf("ReadMsg() size = %v, expected %v", size, m.size) }
  for i := 0; i != m.size; i++ {
    if b, err := s.ReadByte(); err != nil {
      t.Errorf("ReadMsg() failed to read paylad: %v", err.Error())
    } else if b != kPayloadByte {
      t.Errorf("ReadMsg() => '%c', expected '%c'", b, kPayloadByte)
    }
  }
}


func assertWriteReadMsg(t *testing.T, s *bytes.Buffer, m *testMsg) {
  assertWriteMsg(t,s,m)
  assertReadMsg(t,s,m)
}


func TestWriteReadMsg(t *testing.T) {
  s := new(bytes.Buffer)

  // Make sure ReadMsg behaves correctly for an (invalid) empty notification
  s.Write([]byte("n00000000000"))
  assertReadMsg(t,s, &testMsg{MsgTypeNotification, "", "", 0})

  m := []*testMsg{
    &testMsg{MsgTypeSingleReq,     "abc", "echo",    0},
    &testMsg{MsgTypeSingleReq,     "zzz", "lolcats", 3},
    &testMsg{MsgTypeStreamReq,     "abc", "echo",    4},
    &testMsg{MsgTypeStreamReqPart, "abc", "",        5},
    &testMsg{MsgTypeSingleRes,     "abc", "",        6},
    &testMsg{MsgTypeStreamRes,     "abc", "",        7},
    &testMsg{MsgTypeErrorRes,      "abc", "",        8},
    &testMsg{MsgTypeNotification,  "",    "hello",   9},
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
