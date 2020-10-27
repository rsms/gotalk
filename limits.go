package gotalk

import (
	"math/rand"
	"sync/atomic"
	"time"
)

// DefaultLimits does not limit buffer requests, and disables stream requests.
var DefaultLimits = &Limits{
	ReadTimeout:    30 * time.Second,
	BufferRequests: Unlimited,
	StreamRequests: 0, // disallow stream requests
	BufferMinWait:  500 * time.Millisecond,
	BufferMaxWait:  5000 * time.Millisecond,
	StreamMinWait:  500 * time.Millisecond,
	StreamMaxWait:  5000 * time.Millisecond,
}

// NoLimits does not limit buffer requests or stream requests, nor does it have a read timeout.
var NoLimits = &Limits{
	BufferRequests: Unlimited,
	StreamRequests: Unlimited,
}

// Unlimited can be used with Limits.BufferRequests and Limits.StreamRequests
const Unlimited = uint32(0xFFFFFFFF)

type Limits struct {
	ReadTimeout time.Duration // timeout for reading messages from the network (0=no limit)

	BufferRequests uint32 // max number of concurrent buffer requests
	StreamRequests uint32 // max number of concurrent buffer requests

	BufferMinWait time.Duration // minimum time to wait when BufferRequests has been reached
	BufferMaxWait time.Duration // max time to wait when BufferRequests has been reached

	StreamMinWait time.Duration // minimum time to wait when StreamRequests has been reached
	StreamMaxWait time.Duration // max time to wait when StreamRequests has been reached
}

// Create new Limits based on DefaultLimits
// It's usually easier to just construct Limits{} manually. This function is here mainly for
// backwards API compatibility with earlier Gotalk.
func NewLimits(bufferRequestLimit uint32, streamRequestLimit uint32) *Limits {
	l := *DefaultLimits
	l.BufferRequests = bufferRequestLimit
	l.StreamRequests = streamRequestLimit
	return &l
}

// -----------------------------------------------------------------------------------------------

type limitCounter struct {
	limit uint32
	count uint32
}

func (l *limitCounter) inc() bool {
	n := atomic.AddUint32(&l.count, 1)
	if n > l.limit {
		l.dec()
		return false
	}
	return true
}

func (l *limitCounter) dec() {
	atomic.AddUint32(&l.count, ^uint32(0)) // see godoc sync/atomic/#AddUint32
}

// -----------------------------------------------------------------------------------------------

type limitsImpl struct {
	readTimeout time.Duration // message reading timeout
	bufferLimit limitCounter
	streamLimit limitCounter

	bufferMinWait, bufferMaxWait uint32
	streamMinWait, streamMaxWait uint32
}

func makeLimitsImpl(limits *Limits) limitsImpl {
	if limits == nil {
		limits = DefaultLimits
	}
	bufferMinWait, bufferMaxWait := limits.BufferMinWait, limits.BufferMaxWait
	streamMinWait, streamMaxWait := limits.StreamMinWait, limits.StreamMaxWait

	if bufferMinWait <= 0 {
		bufferMinWait = DefaultLimits.BufferMinWait
	}
	if bufferMaxWait <= 0 {
		bufferMaxWait = DefaultLimits.BufferMaxWait
	} else if bufferMaxWait < bufferMinWait {
		bufferMaxWait = bufferMinWait
	}

	if streamMinWait <= 0 {
		streamMinWait = DefaultLimits.StreamMinWait
	}
	if streamMaxWait <= 0 {
		streamMaxWait = DefaultLimits.StreamMaxWait
	} else if streamMaxWait < streamMinWait {
		streamMaxWait = streamMinWait
	}

	return limitsImpl{
		readTimeout:   limits.ReadTimeout,
		bufferLimit:   limitCounter{limit: limits.BufferRequests},
		streamLimit:   limitCounter{limit: limits.StreamRequests},
		bufferMinWait: uint32(bufferMinWait / time.Millisecond),
		bufferMaxWait: uint32(bufferMaxWait / time.Millisecond),
		streamMinWait: uint32(streamMinWait / time.Millisecond),
		streamMaxWait: uint32(streamMaxWait / time.Millisecond),
	}
}

func (l *limitsImpl) incBufferReq() bool {
	if l.bufferLimit.limit == Unlimited {
		return true
	}
	return l.bufferLimit.inc()
}

func (l *limitsImpl) decBufferReq() {
	if l.bufferLimit.limit != Unlimited {
		l.bufferLimit.dec()
	}
}

func (l *limitsImpl) streamReqEnabled() bool {
	return l.streamLimit.limit > 0
}

func (l *limitsImpl) incStreamReq() bool {
	if l.streamLimit.limit == Unlimited {
		return true
	}
	return l.streamLimit.inc()
}

func (l *limitsImpl) decStreamReq() {
	if l.streamLimit.limit != Unlimited {
		l.streamLimit.dec()
	}
}

func (l *limitsImpl) waitBufferReq() uint32 {
	// Time to tell requestor to wait when sending a buffer requests while limit has been reached
	return randUint32(l.bufferMinWait, l.bufferMaxWait)
}

func (l *limitsImpl) waitStreamReq() uint32 {
	// Time to tell requestor to wait when sending a streaming requests while limit has been reached
	return randUint32(l.streamMinWait, l.streamMaxWait)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randUint32(min, max uint32) uint32 {
	return min + uint32(rand.Intn(int(max-min)))
}
