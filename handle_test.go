package gotalk
import (
  "testing"
  "bytes"
  "runtime/debug"
)


func recoverAsFail(t *testing.T) {
  if v := recover(); v != nil {
    t.Log(v)
    t.Log(string(debug.Stack()))
    t.Fail()
  }
}


func checkReqHandler(t *testing.T, s *Sock, h *Handlers, name, input, expectedOutput string) {
  if a := h.FindBufferRequestHandler(name); a == nil {
    t.Errorf("handler '%s' not found", name)
  } else if outbuf, err := a(s, name, []byte(input)); err != nil {
    t.Errorf("handler '%s' returned an error: %s", name, err.Error())
  } else if bytes.Equal(outbuf, []byte(expectedOutput)) == false {
    t.Errorf("handler '%s' returned '%s', expected '%s'", name, string(outbuf), expectedOutput)
  }
}


func TestRequestFuncHandlers(t *testing.T) {
  h := NewHandlers()

  // Handlers.Handle panic() when the signature is incorrect, so treat panic as test failure
  defer recoverAsFail(t)

  invocationCount := 0

  // All possible handler func permutations (with `int` value types)
  h.Handle("a", func(s *Sock, op string, p int) (int, error) {
    if op != "a" {
      t.Errorf("expected op='a' but got '%s'", op)
    }
    invocationCount++
    return p+1, nil
  })
  h.Handle("b", func(s *Sock, p int) (int, error) {
    invocationCount++
    return p+1, nil
  })
  h.Handle("c", func(p int) (int, error) {
    invocationCount++
    return p+1, nil
  })
  h.Handle("d", func(s *Sock) (int, error) {
    invocationCount++
    return 1, nil
  })
  h.Handle("e", func() (int, error) {
    invocationCount++
    return 1, nil
  })
  h.Handle("f", func(s *Sock, op string, p int) error {
    if op != "f" {
      t.Errorf("expected op='f' but got '%s'", op)
    }
    invocationCount++
    return nil
  })
  h.Handle("g", func(s *Sock, p int) error {
    invocationCount++
    return nil
  })
  h.Handle("h", func(p int) error {
    invocationCount++
    return nil
  })
  h.Handle("i", func(s *Sock) error {
    invocationCount++
    return nil
  })
  h.Handle("j", func() error {
    invocationCount++
    return nil
  })
  h.Handle("", func(s *Sock, op string, p int) error {
    if op != "fallback1" && op != "fallback2" {
      t.Errorf("expected op='fallback1'||'fallback2' but got '%s'", op)
    }
    invocationCount++
    return nil
  })

  s := NewSock(h)

  checkReqHandler(t,s,h, "a", "1", "2")
  checkReqHandler(t,s,h, "b", "1", "2")
  checkReqHandler(t,s,h, "c", "1", "2")
  checkReqHandler(t,s,h, "d", "", "1")
  checkReqHandler(t,s,h, "e", "", "1")
  checkReqHandler(t,s,h, "f", "1", "")
  checkReqHandler(t,s,h, "g", "1", "")
  checkReqHandler(t,s,h, "h", "1", "")
  checkReqHandler(t,s,h, "i", "", "")
  checkReqHandler(t,s,h, "j", "", "")
  checkReqHandler(t,s,h, "fallback1", "1", "")
  checkReqHandler(t,s,h, "fallback2", "1", "")

  if invocationCount != 12 {
    t.Error("not all handlers were invoked")
  }
}


func checkNotHandler(t *testing.T, s *Sock, h *Handlers, name, input string) {
  if a := h.FindNotificationHandler(name); h == nil {
    t.Errorf("handler '%s' not found", name)
  } else {
    a(s, name, []byte(input))
  }
}


func TestNotificationFuncHandlers(t *testing.T) {
  h := NewHandlers()
  defer recoverAsFail(t)
  invocationCount := 0

  // All possible handler func permutations (with `int` value types)
  h.HandleNotification("a", func(s *Sock, name string, p int) {
    if name != "a" { t.Errorf("expected name='a' but got '%s'", name) }
    if p != 1 { t.Errorf("expected p='1' but got '%v'", p) }
    invocationCount++
  })
  h.HandleNotification("b", func(name string, p int) {
    if name != "b" { t.Errorf("expected name='b' but got '%s'", name) }
    if p != 2 { t.Errorf("expected p='2' but got '%v'", p) }
    invocationCount++
  })
  h.HandleNotification("c", func(p int) {
    if p != 3 { t.Errorf("expected p='3' but got '%v'", p) }
    invocationCount++
  })
  h.HandleNotification("", func(name string, p int) {
    if name != "fallback1" && name != "fallback2" {
      t.Errorf("expected name='fallback1'||'fallback2' but got '%s'", name)
    }
    if p != 4 { t.Errorf("expected p='4' but got '%v'", p) }
    invocationCount++
  })

  s := NewSock(h)

  checkNotHandler(t,s,h, "a", "1")
  checkNotHandler(t,s,h, "b", "2")
  checkNotHandler(t,s,h, "c", "3")
  checkNotHandler(t,s,h, "fallback1", "4")
  checkNotHandler(t,s,h, "fallback2", "4")

  if invocationCount != 5 {
    t.Error("not all handlers were invoked")
  }
}

