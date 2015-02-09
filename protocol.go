package gotalk

import (
  "io"
  "strconv"
  "errors"
  "encoding/binary"
  "bytes"
  "time"
)

const (
  // Version of this protocol
  ProtocolVersion      = uint8(1)

  // Protocol message types
  MsgTypeSingleReq     = MsgType(byte('r'))
  MsgTypeStreamReq     = MsgType(byte('s'))
  MsgTypeStreamReqPart = MsgType(byte('p'))
  MsgTypeSingleRes     = MsgType(byte('R'))
  MsgTypeStreamRes     = MsgType(byte('S'))
  MsgTypeErrorRes      = MsgType(byte('E'))
  MsgTypeRetryRes      = MsgType(byte('e'))
  MsgTypeNotification  = MsgType(byte('n'))
  MsgTypeHeartbeat     = MsgType(byte('h'))
  MsgTypeProtocolError = MsgType(byte('f'))
)

const (
  ProtocolErrorAbnormal    = 0
  ProtocolErrorUnsupported = 1
  ProtocolErrorInvalidMsg  = 2
  ProtocolErrorTimeout     = 3
)

// Protocol message type
type MsgType byte

// Write the version this protocol implements to `s`
func WriteVersion(s io.Writer) (int, error) {
  return s.Write(protocolVersionBuf[:])
}

// Read the version the other end implements. Returns an error if this side's protocol
// is incompatible with the other side's version.
func ReadVersion(s io.Reader) (uint8, error) {
  b := make([]byte, 2)
  if _, err := readn(s, b); err != nil {
    return 0, err
  }
  n, err := strconv.ParseUint(string(b), 16, 8)
  if err != nil {
    return 0, err
  }
  if n != uint64(ProtocolVersion) {
    return 0, errors.New("unsupported protocol version \"" + string(b) + "\"")
  }
  return uint8(n), nil
}


// Maximum value of a heartbeat's "load"
var HeartbeatMsgMaxLoad = 0xffff


// Create a slice of bytes representing a heartbeat message
func MakeHeartbeatMsg(load uint16) []byte {
  b := []byte{byte(MsgTypeHeartbeat),0,0,0,0,0,0,0,0,0,0,0,0}
  z := 1
  copyFixnum(b[z:z+4], 4, uint64(load), 16)
  z += 4
  copyFixnum(b[z:z+8], 8, uint64(time.Now().UTC().Unix()), 16)
  return b
}


// Create a slice of bytes representing a message (w/o any payload)
func MakeMsg(t MsgType, id, name3 string, wait, size int) []byte {
  // calculate buffer size
  bz := 9  // minimum size, fitting type and payload size
  name3z := 0

  if t == MsgTypeRetryRes {
    bz = 21  // e.g. "e00010000000100000001"
  } else {
    if id != "" {
      bz += 4  // msg with id e.g. "R000100000005"
    }
    name3z = len(name3)
    if name3z != 0 {
      bz += 3 + name3z  // msg w/ name3 e.g. "r0001004echo00000005"
    }
  }

  b := make([]byte, bz)
  b[0] = byte(t)  // type e.g. "R"
  z := 1

  if id != "" {
    b[1] = id[0]
    b[2] = id[1]
    b[3] = id[2]
    b[4] = id[3]  // id e.g. "abcd"
    z += 4
  }

  if name3z != 0 {
    if len(name3) == 0 {
      panic("empty name")
    }
    copyFixnum(b[z:z+3], 3, uint64(name3z), 16) // name3 size e.g. "004"
    z += 3
    copy(b[z:], []byte(name3))
    z += name3z
  }

  if t == MsgTypeRetryRes {
    if wait == 0 {
      copy(b[z:], zeroes[:8])
    } else {
      copyFixnum(b[z:z+8], 8, uint64(wait), 16)
    }
    z += 8
  }

  if size == 0 {
    copy(b[z:], zeroes[:8])
  } else {
    copyFixnum(b[z:z+8], 8, uint64(size), 16)  // payload size e.g. "0000005"
  }

  return b[:z+8]
}


// Read a message from `s`
// If t is MsgTypeHeartbeat, wait==load, size==time
func ReadMsg(s io.Reader) (t MsgType, id, name3 string, wait, size uint32, err error) {
  // "r0001004echo00000005"  => ('r', "0001", "echo", 0, 5, nil)
  // "R000100000005"         => ('R', "0001", "", 0, 5, nil)
  // "e00010000138800000014" => ('e', "0001", "", 5000, 20, nil)
  b := make([]byte, 128)

  // A message has a minimum size of 13, so read first 13 bytes
  // e.g. "n001a00000000" = <notification> <short name> <no payload>
  readz := 13
  readz, err = readn(s, b[:readz])
  if err != nil {
    if err == io.EOF && readz >= 9 && b[0] == byte(MsgTypeProtocolError) {
      // OK to read until EOF for MsgTypeProtocolError as they are shorter than other messages
      err = nil
    } else {
      return
    }
  }

  // type
  t = MsgType(b[0])
  z := 1

  if t == MsgTypeHeartbeat {
    // load
    var n uint64
    n, err = strconv.ParseUint(string(b[z:z+4]), 16, 16)
    z += 4
    if err != nil {
      return
    }
    wait = uint32(n)

  } else if t != MsgTypeNotification && t != MsgTypeProtocolError {
    // requestID
    id = string(b[z:z+4])
    z += 4
  }

  if t == MsgTypeSingleReq || t == MsgTypeStreamReq || t == MsgTypeNotification {
    // name
    // text3Size
    name3z, e := strconv.ParseUint(string(b[z:z+3]), 16, 16)
    z += 3
    if e != nil {
      err = e
      return
    }

    // Read remainder of message
    newz := z + int(name3z) + 8  // 8 = payload size

    if cap(b) < newz {
      // Grow buffer (only happens with really long name3)
      newb := make([]byte, newz)
      copy(newb, b)
      b = newb
    }

    if newz > readz {
      if _, err = readn(s, b[readz:newz]); err != nil {
        return
      }
    }

    // text3Value
    name3 = string(b[z:z+int(name3z)])
    z += int(name3z)

  } else if t == MsgTypeRetryRes {
    // wait
    n, e := strconv.ParseUint(string(b[z:z+8]), 16, 32)
    if e != nil {
      err = e
      return
    }
    wait = uint32(n)
    z += 8
    // read remainding 8 bytes of the message
    if _, err = readn(s, b[z:z+8]); err != nil {
      return
    }
  }

  // payloadSize (or time if t==MsgTypeHeartbeat)
  n, e := strconv.ParseUint(string(b[z:z+8]), 16, 32)
  if e != nil {
    err = e
    return
  }
  size = uint32(n)

  return
}


// Returns a 4-byte representation of a 32-bit integer, suitable an integer-based request ID.
func FormatRequestID(n uint32) []byte {
  buf := bytes.NewBuffer(make([]byte,4)[:0])
  err := binary.Write(buf, binary.LittleEndian, n)
  if err != nil {
    panic(err)
  }
  return buf.Bytes()
}


// =============================================================================

var protocolVersionBuf [2]byte  // version as fixnum2 ([H,H] where H is any of 0-9a-fA-F)
var zeroes = [...]byte{48,48,48,48,48,48,48,48}

func init() {
  copyFixnum(protocolVersionBuf[:0], 2, uint64(ProtocolVersion), 16)
}


func copyFixnum(buf []byte, ndigits int, n uint64, base int) {
  z := len(strconv.AppendUint(buf[:0], n, base))
  rShiftSlice(buf[:ndigits], ndigits-z, byte(48))
}


func makeFixnumBuf(ndigits int, n uint64, base int) []byte {
  zb := make([]byte, ndigits)
  z := len(strconv.AppendUint(zb[:0], n, base))
  rShiftSlice(zb, ndigits-z, byte(48))
  return zb
}


func rShiftSlice(b []byte, n int, padb byte) {
  if n != 0 {
    bz := len(b)
    wi := bz-1
    for ; wi >= n; wi-- {
      b[wi] = b[wi-n]
      b[wi-n] = padb
    }
    zn := bz-n-1
    for ; wi > zn; wi-- {
      b[wi] = padb
    }
  }
}


// Read exactly len(b) bytes from s, blocking if needed
func readn(s io.Reader, b []byte) (int, error) {
  // behaves similar to io.ReadFull, but simpler and allowing EOF<len(b)
  p := 0
  n := len(b)
  for p < n {
    z, err := s.Read(b[p:])
    p += z
    if err != nil {
      return p, err
    }
  }
  return p, nil
}
