package gotalk
import (
  "sync/atomic"
  "math/rand"
  "time"
)

type Limits interface {
  // Maximum amount of time allowed to read a buffer request. 0 = no timeout.
  // Defaults to 30 seconds.
  ReadTimeout() time.Duration
  SetReadTimeout(time.Duration)

  incBufferReq() bool
  decBufferReq()
  streamReqEnabled() bool
  incStreamReq() bool
  decStreamReq()
}

// Create new Limits, limiting request processing.
//
// `streamRequestLimit` limits the amount of stream requests but works together with `requestLimit`
// meaning that we can handle `requestLimit` requests of any type, but no more than
//
// `streamRequestLimit` of the streaming kind. Say `streamRequestLimit=5` and `requestLimit=10`,
// and we are currently processing 5 streaming requests, we can handle an additional 5 buffered
// requests, but no more streaming requests.
//
// - If both `requestLimit` and `streamRequestLimit` is 0, buffer requests are not limited and
//   stream requests are disabled.
// - If `streamRequestLimit` is 0, buffer requests are limited to `requestLimit` and stream
//   requests are disabled.
// - If `requestLimit` is 0, buffer requests aren't limited, but stream requests are limited
//   to `streamRequestLimit`.
//
func NewLimits(requestLimit uint32, streamRequestLimit uint32) Limits {
  if requestLimit == 0 && streamRequestLimit == 0 {
    return &noLimitNoStream{DefaultLimits.ReadTimeout()}

  } else if requestLimit == 0 {
    return &limitStream{limit{streamRequestLimit, 0}, DefaultLimits.ReadTimeout()}

  } else if streamRequestLimit == 0 {
    return &limitSingleNoStream{limit{requestLimit, 0}, DefaultLimits.ReadTimeout()}

  } else {
    if streamRequestLimit > requestLimit {
      panic("streamRequestLimit > requestLimit")
    }
    return &limitSingleAndStream{
      limit{requestLimit, 0},
      limit{streamRequestLimit, 0},
      DefaultLimits.ReadTimeout(),
    }
  }
}

// DefaultLimits does not limit buffer requests, and disables stream requests.
var DefaultLimits Limits = &noLimitNoStream{30 * time.Second}

// NoLimits does not limit buffer requests or stream requests, not does it have a read timeout.
var NoLimits Limits = noLimit(false)

// -----------------------------------------------------------------------------------------------

type noLimit bool
func (l noLimit) incBufferReq() bool { return true }
func (l noLimit) decBufferReq() {}
func (l noLimit) streamReqEnabled() bool { return true }
func (l noLimit) incStreamReq() bool { return true }
func (l noLimit) decStreamReq() {}
func (l noLimit) ReadTimeout() time.Duration { return 0 }
func (l noLimit) SetReadTimeout(_ time.Duration) {}

// -----------------------------------------------------------------------------------------------

type noLimitNoStream struct {
  readTimeout time.Duration
}
func (l *noLimitNoStream) incBufferReq() bool { return true }
func (l *noLimitNoStream) decBufferReq() {}
func (l *noLimitNoStream) streamReqEnabled() bool { return false }
func (l *noLimitNoStream) incStreamReq() bool { return false }
func (l *noLimitNoStream) decStreamReq() {}
func (l *noLimitNoStream) ReadTimeout() time.Duration { return l.readTimeout }
func (l *noLimitNoStream) SetReadTimeout(d time.Duration) { l.readTimeout = d }

// -----------------------------------------------------------------------------------------------

type limitStream struct {
  streamLimit limit
  readTimeout time.Duration
}
func (l *limitStream) incBufferReq() bool { return true }
func (l *limitStream) decBufferReq() {}
func (l *limitStream) streamReqEnabled() bool { return true }
func (l *limitStream) incStreamReq() bool { return l.streamLimit.inc() }
func (l *limitStream) decStreamReq() { l.streamLimit.dec() }
func (l *limitStream) ReadTimeout() time.Duration { return l.readTimeout }
func (l *limitStream) SetReadTimeout(d time.Duration) { l.readTimeout = d }

// -----------------------------------------------------------------------------------------------

type limitSingleNoStream struct {
  singleLimit limit
  readTimeout time.Duration
}
func (l *limitSingleNoStream) incBufferReq() bool { return l.singleLimit.inc() }
func (l *limitSingleNoStream) decBufferReq() { l.singleLimit.dec() }
func (l *limitSingleNoStream) streamReqEnabled() bool { return false }
func (l *limitSingleNoStream) incStreamReq() bool { return false }
func (l *limitSingleNoStream) decStreamReq() {}
func (l *limitSingleNoStream) ReadTimeout() time.Duration { return l.readTimeout }
func (l *limitSingleNoStream) SetReadTimeout(d time.Duration) { l.readTimeout = d }

// -----------------------------------------------------------------------------------------------

type limitSingleAndStream struct {
  bothLimit   limit
  streamLimit limit
  readTimeout time.Duration
}
func (l *limitSingleAndStream) incBufferReq() bool { return l.bothLimit.inc() }
func (l *limitSingleAndStream) decBufferReq() { l.bothLimit.dec() }
func (l *limitSingleAndStream) streamReqEnabled() bool { return true }
func (l *limitSingleAndStream) incStreamReq() bool {
  if l.bothLimit.inc() {
    if l.streamLimit.inc() {
      return true
    }
    l.bothLimit.dec()
  }
  return false
}
func (l *limitSingleAndStream) decStreamReq() {
  l.streamLimit.dec()
  l.bothLimit.dec()
}
func (l *limitSingleAndStream) ReadTimeout() time.Duration { return l.readTimeout }
func (l *limitSingleAndStream) SetReadTimeout(d time.Duration) { l.readTimeout = d }

// -----------------------------------------------------------------------------------------------

type limit struct {
  limit uint32
  count uint32
}

func (l *limit) inc() bool {
  n := atomic.AddUint32(&l.count, 1)
  if n > l.limit {
    l.dec()
    return false
  }
  return true
}

func (l *limit) dec() {
  atomic.AddUint32(&l.count, ^uint32(0))  // see godoc sync/atomic/#AddUint32
}

// -----------------------------------------------------------------------------------------------

func init() {
  rand.Seed(time.Now().UTC().UnixNano())
}

func limitWait(min, max int) int {
  return min + rand.Intn(max - min)
}

func limitWaitStreamReq() int {
  // Time to tell requestor to wait when sending a streaming requests while limit has been reached
  return limitWait(500, 5000)
}

func limitWaitBufferReq() int {
  // Time to tell requestor to wait when sending a buffer requests while limit has been reached
  return limitWait(500, 5000)
}
