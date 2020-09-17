package gotalk

import (
	"bytes"
	"io"
	"testing"
)

func TestWriteReadVersion(t *testing.T) {
	// this test assumes version 1
	if ProtocolVersion != 1 {
		t.Errorf("ProtocolVersion = %v, expected 1", ProtocolVersion)
	}

	// read from insufficient input should be an error
	_, err := ReadVersion(&bytes.Buffer{})
	assertError(t, "EOF", err)

	_, err = ReadVersion(bytes.NewBufferString("0"))
	assertError(t, "EOF", err)

	// reading invalid, non-hexadecimal input should fail
	_, err = ReadVersion(bytes.NewBufferString("xx"))
	assertError(t, "invalid syntax", err)

	// unsupported protocol version
	_, err = ReadVersion(bytes.NewBufferString("09"))
	assertError(t, "unsupported protocol version", err)

	// read valid version
	v, _ := ReadVersion(bytes.NewBufferString("01"))
	assertEq(t, uint8(1), v)

	// write version
	s := new(bytes.Buffer)
	if n, err := WriteVersion(s); err != nil {
		t.Errorf("WriteVersion() failed: %v", err.Error())
	} else if n != 2 {
		t.Errorf("WriteVersion() => (%v, _), expected (2, _)", n)
	} else {
		assertBytes(t, s.Bytes(), []byte("01"))
	}
}

type testMsg struct {
	t       MsgType
	id      string
	name    string
	u1      uint32 // depents on t. "wait" for most, "load" for heartbeat
	u2      uint32 // depents on t. "size" for most, "time" for heartbeat
	encoded []byte // known encoded form, for testing
}

var sampleMessages []testMsg

func init() {
	sampleMessages = []testMsg{
		{MsgTypeSingleReq, "idid", "echo", 0, 0, []byte("ridid004echo00000000")},
		{MsgTypeSingleReq, "idid", "echo", 0, 3, []byte("ridid004echo00000003")},
		{MsgTypeStreamReq, "idid", "echo", 0, 3, []byte("sidid004echo00000003")},
		{MsgTypeStreamReqPart, "idid", "", 0, 3, []byte("pidid00000003")},
		{MsgTypeSingleRes, "idid", "", 0, 3, []byte("Ridid00000003")},
		{MsgTypeStreamRes, "idid", "", 0, 3, []byte("Sidid00000003")},
		{MsgTypeErrorRes, "idid", "", 0, 3, []byte("Eidid00000003")},
		{MsgTypeRetryRes, "idid", "", 6, 3, []byte("eidid0000000600000003")},
		{MsgTypeRetryRes, "idid", "", 0, 3, []byte("eidid0000000000000003")},
		{MsgTypeNotification, "", "hello", 0, 3, []byte("n005hello00000003")},
		// {MsgTypeHeartbeat, "", "", 2, 0x5f63ee48, []byte("h00025f63ee48")}, MakeMsg can't handle it
	}
}

func TestMakeMsg(t *testing.T) {
	for _, m := range sampleMessages {
		assertBytes(t, m.encoded, MakeMsg(m.t, m.id, m.name, m.u1, m.u2))
	}

	// heartbeat is "h" <load uint16> <time uint64>
	hb := MakeHeartbeatMsg(2, make([]byte, 16))
	assertBytes(t, []byte("h0002"), hb[:5])
	assertEq(t, 1+4+8, len(hb))

	// only first 4 bytes of id is used
	assertBytes(t, []byte("rABCD004echo00000000"), MakeMsg(MsgTypeSingleReq, "ABCDEF", "echo", 0, 0))

	// id must be at least 4 bytes
	assertPanic(t, "index out of range", func() {
		MakeMsg(MsgTypeSingleReq, "aaa", "echo", 0, 0)
	})
}

func TestReadMsg(t *testing.T) {
	tmpbuf := make([]byte, 128)

	// successfully parse a message
	ty, id, name, wait, size, err := ReadMsg(bytes.NewBufferString("ridid004echo00000003"), tmpbuf)
	assertEq(t, MsgTypeSingleReq, ty)
	assertEq(t, "idid", id)
	assertEq(t, "echo", name)
	assertEq(t, uint32(0), wait)
	assertEq(t, uint32(3), size)

	// read from insufficient input should cause an error
	_, _, _, _, _, err = ReadMsg(bytes.NewBufferString(""), tmpbuf)
	assertError(t, "EOF", err)

	// read valid heartbeat message
	ty, id, name, load, time, err := ReadMsg(bytes.NewBufferString("h00025f63ee48"), tmpbuf)
	assertEq(t, MsgTypeHeartbeat, ty)
	assertEq(t, "", id)
	assertEq(t, "", name)
	assertEq(t, uint32(2), load)
	assertEq(t, uint32(0x5f63ee48), time)

	// read heartbeat with invalid load data
	_, _, _, _, _, err = ReadMsg(bytes.NewBufferString("hXXXX5f63ee48"), tmpbuf)
	assertError(t, "invalid syntax", err)
}

func TestWriteMsg(t *testing.T) {
	s := new(bytes.Buffer)

	// Make sure ReadMsg behaves correctly
	s.Write([]byte("n001a00000000"))
	assertReadMsg(t, s, &testMsg{MsgTypeNotification, "", "a", 0, 0, []byte{}})

	m := []*testMsg{
		{MsgTypeSingleReq, "abcd", "echo", 0, 0, []byte{}},
		{MsgTypeSingleReq, "zzzz", "lolcats", 0, 3, []byte{}},
		{MsgTypeStreamReq, "abcd", "echo", 0, 4, []byte{}},
		{MsgTypeStreamReqPart, "abcd", "", 0, 5, []byte{}},
		{MsgTypeSingleRes, "abcd", "", 0, 6, []byte{}},
		{MsgTypeStreamRes, "abcd", "", 0, 7, []byte{}},
		{MsgTypeErrorRes, "abcd", "", 0, 8, []byte{}},
		{MsgTypeRetryRes, "abcd", "", 6, 8, []byte{}},
		{MsgTypeNotification, "", "hello", 0, 9, []byte{}},
		{MsgTypeProtocolError, "", "", 0, ProtocolErrorInvalidMsg, []byte{}},
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

// ————————————————————————————————————————————————————————————————————————————————————
// helpers

const kPayloadByte = byte('~')

func assertWriteMsg(t *testing.T, s *bytes.Buffer, m *testMsg) {
	msg := MakeMsg(m.t, m.id, m.name, m.u1, m.u2)
	if _, err := s.Write(msg); err != nil {
		t.Errorf("s.Write() failed: %v", err.Error())
	}
	// Write ayload data
	if m.t != MsgTypeProtocolError && m.t != MsgTypeHeartbeat {
		for i := uint32(0); i != m.u2; i++ {
			s.WriteByte(kPayloadByte)
		}
	}
	// t.Logf("s.String() => '%s'\n", s.String())
}

func assertReadMsg(t *testing.T, s *bytes.Buffer, m *testMsg) {
	tmpbuf := make([]byte, 128)
	ty, id, name, u1, u2, err := ReadMsg(s, tmpbuf)
	if err != nil && err != io.EOF {
		t.Fatalf("ReadMsg() failed: %v", err.Error())
	}
	if ty != m.t {
		t.Fatalf("ReadMsg() ty = \"%v\", expected \"%v\"", ty, m.t)
	}
	if id != m.id {
		t.Fatalf("ReadMsg() id = \"%v\", expected \"%v\"", id, m.id)
	}
	if name != m.name {
		t.Fatalf("ReadMsg() name = \"%v\", expected \"%v\"", name, m.name)
	}
	if u1 != m.u1 {
		t.Fatalf("ReadMsg() wait|load = %v, expected %v", u1, m.u1)
	}
	if u2 != m.u2 {
		t.Fatalf("ReadMsg() size|time = %v, expected %v", u2, m.u2)
	}
	// Read payload
	if m.t != MsgTypeProtocolError && m.t != MsgTypeHeartbeat {
		for i := uint32(0); i != m.u2; i++ {
			if b, err := s.ReadByte(); err != nil {
				t.Fatalf("ReadMsg() failed to read payload: %v", err.Error())
			} else if b != kPayloadByte {
				t.Fatalf("ReadMsg() => '%c', expected '%c'", b, kPayloadByte)
			}
		}
	}
}

func assertWriteReadMsg(t *testing.T, s *bytes.Buffer, m *testMsg) {
	assertWriteMsg(t, s, m)
	// println(s.String())
	assertReadMsg(t, s, m)
}
