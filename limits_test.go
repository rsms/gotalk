package gotalk

import (
	"testing"
	"time"
)

func TestLimits(t *testing.T) {
	l := makeLimitsImpl(NoLimits)
	assertEq(t, l.incBufferReq(), true)
	assertEq(t, l.streamReqEnabled(), true)
	assertEq(t, l.incStreamReq(), true)
	assertEq(t, l.readTimeout, time.Duration(0))

	// unlimited buffer requests, no stream requests
	l = makeLimitsImpl(NewLimits(Unlimited, 0))
	assertEq(t, l.incBufferReq(), true)
	assertEq(t, l.streamReqEnabled(), false)
	assertEq(t, l.incStreamReq(), false)

	// 1 buffer request, no stream requests
	l = makeLimitsImpl(NewLimits(1, 0))
	assertEq(t, l.incBufferReq(), true)
	assertEq(t, l.incBufferReq(), false)
	l.decBufferReq()
	assertEq(t, l.incBufferReq(), true)
	assertEq(t, l.streamReqEnabled(), false)
	assertEq(t, l.incStreamReq(), false)

	// unlimited buffer requests, 1 stream request
	l = makeLimitsImpl(NewLimits(Unlimited, 1))
	assertEq(t, l.incBufferReq(), true)
	assertEq(t, l.streamReqEnabled(), true)
	assertEq(t, l.incStreamReq(), true)
	assertEq(t, l.incStreamReq(), false)
	l.decStreamReq()
	assertEq(t, l.incStreamReq(), true)

	// 2 buffer requests, 2 stream request
	l = makeLimitsImpl(NewLimits(2, 2))

	// test limitCounter, bufferLimit
	assertEq(t, l.bufferLimit.count, uint32(0))
	assertEq(t, l.incBufferReq(), true)
	assertEq(t, l.bufferLimit.count, uint32(1))
	assertEq(t, l.incBufferReq(), true)
	assertEq(t, l.bufferLimit.count, uint32(2))
	assertEq(t, l.incBufferReq(), false)
	assertEq(t, l.bufferLimit.count, uint32(2))
	l.decBufferReq()
	assertEq(t, l.incBufferReq(), true)
	assertEq(t, l.bufferLimit.count, uint32(2))
	l.decBufferReq()
	assertEq(t, l.bufferLimit.count, uint32(1))
	l.decBufferReq()
	assertEq(t, l.bufferLimit.count, uint32(0))

	// test limitCounter, streamLimit
	assertEq(t, l.streamReqEnabled(), true)
	assertEq(t, l.streamLimit.count, uint32(0))
	assertEq(t, l.incStreamReq(), true)
	assertEq(t, l.streamLimit.count, uint32(1))
	assertEq(t, l.incStreamReq(), true)
	assertEq(t, l.streamLimit.count, uint32(2))
	assertEq(t, l.incStreamReq(), false)
	assertEq(t, l.streamLimit.count, uint32(2))
	l.decStreamReq()
	assertEq(t, l.incStreamReq(), true)
	l.decStreamReq()
	l.decStreamReq()
	assertEq(t, l.bufferLimit.count, uint32(0))
	assertEq(t, l.streamLimit.count, uint32(0))

}
