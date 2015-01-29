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
  assertMsgEqual(t, MakeMsg(MsgTypeSingleReq, "abc", "echo", 0, 0),  []byte("rabc004echo00000000"))
  assertMsgEqual(t, MakeMsg(MsgTypeSingleReq, "abc", "echo", 0, 3),  []byte("rabc004echo00000003"))
  assertMsgEqual(t, MakeMsg(MsgTypeStreamReq, "abc", "echo", 0, 3),  []byte("sabc004echo00000003"))
  assertMsgEqual(t, MakeMsg(MsgTypeStreamReqPart, "abc", "", 0, 3),  []byte("pabc00000003"))
  assertMsgEqual(t, MakeMsg(MsgTypeSingleRes, "abc", "", 0, 3),      []byte("Rabc00000003"))
  assertMsgEqual(t, MakeMsg(MsgTypeStreamRes, "abc", "", 0, 3),      []byte("Sabc00000003"))
  assertMsgEqual(t, MakeMsg(MsgTypeErrorRes, "abc", "", 0, 3),       []byte("Eabc00000003"))
  assertMsgEqual(t, MakeMsg(MsgTypeRetryRes, "abc", "", 6, 3),       []byte("eabc0000000600000003"))
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
  for i := 0; i != m.size; i++ {
    s.WriteByte(kPayloadByte)
  }
  //t.Logf("s.String() => '%s'\n", s.String())
}


func assertReadMsg(t *testing.T, s *bytes.Buffer, m *testMsg) {
  ty, id, name, wait, size, err := ReadMsg(s)
  if err       != nil {    t.Errorf("ReadMsg() failed: %v", err.Error()) }
  if ty        != m.t {    t.Errorf("ReadMsg() ty = \"%v\", expected \"%v\"", ty, m.t) }
  if id        != m.id {   t.Errorf("ReadMsg() id = \"%v\", expected \"%v\"", id, m.id) }
  if name      != m.name { t.Errorf("ReadMsg() name = \"%v\", expected \"%v\"", name, m.name) }
  if int(wait) != m.wait { t.Errorf("ReadMsg() wait = %v, expected %v", wait, m.wait) }
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
  assertReadMsg(t,s, &testMsg{MsgTypeNotification, "", "", 0, 0})

  m := []*testMsg{
    &testMsg{MsgTypeSingleReq,     "abc", "echo",    0, 0},
    &testMsg{MsgTypeSingleReq,     "zzz", "lolcats", 0, 3},
    &testMsg{MsgTypeStreamReq,     "abc", "echo",    0, 4},
    &testMsg{MsgTypeStreamReqPart, "abc", "",        0, 5},
    &testMsg{MsgTypeSingleRes,     "abc", "",        0, 6},
    &testMsg{MsgTypeStreamRes,     "abc", "",        0, 7},
    &testMsg{MsgTypeErrorRes,      "abc", "",        0, 8},
    &testMsg{MsgTypeRetryRes,      "abc", "",        6, 8},
    &testMsg{MsgTypeNotification,  "",    "hello",   0, 9},
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
