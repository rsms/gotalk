package gotalk

import (
  "io"
  "strconv"
  "errors"
)

const (
  // Version of this protocol
  ProtocolVersion      = uint8(0)

  // Protocol message types
  MsgTypeSingleReq     = MsgType(byte('r'))
  MsgTypeStreamReq     = MsgType(byte('s'))
  MsgTypeStreamReqPart = MsgType(byte('p'))
  MsgTypeSingleRes     = MsgType(byte('R'))
  MsgTypeStreamRes     = MsgType(byte('S'))
  MsgTypeErrorRes      = MsgType(byte('E'))
  MsgTypeNotification  = MsgType(byte('n'))
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
  if err := readn(s, b); err != nil {
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


// Create a slice of bytes representing a message (w/o any payload.)
func MakeMsg(t MsgType, id, name3 string, size int) []byte {
  bz := 9  // e.g. "n00000005"
  if id != "" {
    if len(id) != 3 { panic("len(id) != 3") }
    bz += 3  // msg with id e.g. "R00100000005"
  }

  name3z := len(name3)
  if name3z != 0 {
    bz += 3 + name3z  // msg w/ name3 e.g. "r001004echo00000005"
  }

  b := make([]byte, bz)
  b[0] = byte(t)  // type e.g. "R"
  z := 1

  if id != "" {
    b[1] = id[0]
    b[2] = id[1]
    b[3] = id[2]  // id e.g. "abc"
    z += 3
  }

  if name3z != 0 {
    copyFixnum(b[z:z+3], 3, uint64(name3z), 16) // name3 size e.g. "004"
    z += 3
    copy(b[z:], []byte(name3))
    z += name3z
  }

  if size == 0 {
    copy(b[z:], zeroes[:8])
  } else {
    copyFixnum(b[z:z+8], 8, uint64(size), 16)  // payload size e.g. "0000005"
  }

  return b[:z+8]
}


// Read a message from `s`
func ReadMsg(s io.Reader) (t MsgType, id, name3 string, size uint32, err error) {
  // "r001004echo00000005" => ('r', "001", "echo", 5, nil)
  // "R00100000005"        => ('R', "001", "", 5, nil)
  for {
    b := make([]byte, 128)

    // A message has a minimum size of 12, so read first 12 bytes
    if err = readn(s, b[:12]); err != nil {
      break
    }

    t = MsgType(b[0])
    z := 1

    if t != MsgTypeNotification {
      id = string(b[z:z+3])
      z += 3
    }

    if t == MsgTypeSingleReq || t == MsgTypeStreamReq || t == MsgTypeNotification {
      name3z, e := strconv.ParseUint(string(b[z:z+3]), 16, 16)
      z += 3
      if e != nil {
        err = e
        break
      }

      // Read remainder of message
      newz := z + int(name3z) + 8  // 8 = payload size

      if cap(b) < newz {
        // Grow buffer (only happens with really long name3)
        newb := make([]byte, newz)
        copy(newb, b)
        b = newb
      }

      if newz > 12 {
        if err = readn(s, b[12:newz]); err != nil {
          break
        }
      }

      name3 = string(b[z:z+int(name3z)])
      z += int(name3z)
      b = b[z:]
    } else {
      b = b[z:]
    }

    pz, e := strconv.ParseUint(string(b[:8]), 16, 32)
    if e != nil {
      err = e
      break
    }
    size = uint32(pz)
    break
  }

  return t, id, name3, size, err
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
// TODO: is there already a function like this in the io package?
func readn(s io.Reader, b []byte) error {
  p := 0
  n := len(b)
  for p < n {
    z, err := s.Read(b[p:])
    if err != nil {
      return err
    }
    p += z
  }
  return nil
}
