package gotalk

type Request struct {
  MsgType
  Op   string
  Data []byte
}

// Creates a new single request
func NewRequest(op string, buf []byte) *Request {
  return &Request{MsgTypeSingleReq, op, buf}
}

type StreamRequest struct {
  sock    *Sock
  op      string
  id      string
  started bool // request started?
}

func (r *StreamRequest) Write(b []byte) error {
  if r.started == false {
    r.started = true
    if err := r.sock.writeMsg(MsgTypeStreamReq, r.id, r.op, 0, b); err != nil {
      r.finalize()
      return err
    }
  } else {
    if err := r.sock.writeMsg(MsgTypeStreamReqPart, r.id, "", 0, b); err != nil {
      r.finalize()
      return err
    }
  }
  return nil
}

func (r *StreamRequest) End() error {
  err := r.sock.writeMsg(MsgTypeStreamReqPart, r.id, "", 0, nil)
  if err != nil {
    r.finalize()
  }
  return err
}

func (r *StreamRequest) finalize() {
  if r.id != "" {
    r.sock.deallocResChan(r.id)
    r.id = ""
  }
}
