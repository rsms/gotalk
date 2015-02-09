package gotalk
import (
  "testing"
  "time"
)

func verifyLimitsReadTimeout(t *testing.T, l Limits) {
  assertEq(t, l.ReadTimeout(), 30 * time.Second)
  l.SetReadTimeout(10 * time.Second)
  assertEq(t, l.ReadTimeout(), 10 * time.Second)
}

func TestLimits(t *testing.T) {
  l := NoLimits
  assertEq(t, l.incBufferReq(), true)
  assertEq(t, l.streamReqEnabled(), true)
  assertEq(t, l.incStreamReq(), true)
  assertEq(t, l.ReadTimeout(), time.Duration(0))
  l.SetReadTimeout(10 * time.Second)
  assertEq(t, l.ReadTimeout(), time.Duration(0)) // NoLimits is immutable

  // noLimitNoStream
  l = NewLimits(0, 0)
  assertEq(t, l.incBufferReq(), true)
  assertEq(t, l.streamReqEnabled(), false)
  assertEq(t, l.incStreamReq(), false)
  verifyLimitsReadTimeout(t, l)

  // limitStream
  l = NewLimits(0, 1)
  assertEq(t, l.incBufferReq(), true)
  assertEq(t, l.streamReqEnabled(), true)
  assertEq(t, l.incStreamReq(), true)
  assertEq(t, l.incStreamReq(), false)
  l.decStreamReq()
  assertEq(t, l.incStreamReq(), true)
  verifyLimitsReadTimeout(t, l)

  // limitSingleNoStream
  l = NewLimits(1, 0)
  assertEq(t, l.incBufferReq(), true)
  assertEq(t, l.incBufferReq(), false)
  l.decBufferReq()
  assertEq(t, l.incBufferReq(), true)
  assertEq(t, l.streamReqEnabled(), false)
  assertEq(t, l.incStreamReq(), false)
  verifyLimitsReadTimeout(t, l)


  // limitSingleAndStream
  // Note: incBufferReq effectively decrements both requests and streams
  l = NewLimits(2, 2)
  lss, ok := l.(*limitSingleAndStream)
  if !ok {
    t.Fatalf("l.(*limitSingleAndStream) failed\n")
  }

  assertEq(t, l.incBufferReq(), true)
  assertEq(t, l.incBufferReq(), true)
  assertEq(t, l.incBufferReq(), false)
  l.decBufferReq()
  assertEq(t, l.incBufferReq(), true)
  l.decBufferReq()
  l.decBufferReq()
  assertEq(t, lss.bothLimit.count, uint32(0))
  assertEq(t, lss.streamLimit.count, uint32(0))

  assertEq(t, l.streamReqEnabled(), true)

  assertEq(t, l.incStreamReq(), true)
  assertEq(t, lss.bothLimit.count, uint32(1))
  assertEq(t, lss.streamLimit.count, uint32(1))

  assertEq(t, l.incStreamReq(), true)
  assertEq(t, l.incStreamReq(), false)
  l.decStreamReq()
  assertEq(t, l.incStreamReq(), true)
  l.decStreamReq()
  l.decStreamReq()
  assertEq(t, lss.bothLimit.count, uint32(0))
  assertEq(t, lss.streamLimit.count, uint32(0))

  verifyLimitsReadTimeout(t, l)
}

