package gotalk
import "time"

type Response struct {
  MsgType
  Data []byte
  Wait time.Duration   // only valid when IsRetry()==true
}

// Returns a string describing the error, when IsError()==true
func (r *Response) Error() string {
  return string(r.Data)
}

// True if this response is a requestor error (ErrorResult)
func (r *Response) IsError() bool {
  return r.MsgType == MsgTypeErrorRes
}

// True if response is a "server can't handle it right now, please retry" (RetryResult)
func (r *Response) IsRetry() bool {
  return r.MsgType == MsgTypeRetryRes
}

// True if this is part of a streaming response (StreamResult)
func (r *Response) IsStreaming() bool {
  return r.MsgType == MsgTypeStreamRes
}
