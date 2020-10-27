package gotalk

import (
	"bytes"
	"fmt"
	"regexp"
	"runtime/debug"
	"testing"
)

// recoverAsFail catches a panic and converts it into a test failure.
// Example:
//   func TestThing(t *testing.T) {
//     defer recoverAsFail(t)
//     somethingThatMayPanic()
//   }
func recoverAsFail(t *testing.T) {
	if v := recover(); v != nil {
		t.Log(v)
		t.Log(string(debug.Stack()))
		t.Fail()
	}
}

func assertPanic(t *testing.T, expectedPanicRegExp string, f func()) {
	// Note: (?i) makes it case-insensitive
	t.Helper()
	expected := regexp.MustCompile("(?i)" + expectedPanicRegExp)
	defer func() {
		if v := recover(); v != nil {
			panicMsg := fmt.Sprint(v)
			if !expected.MatchString(panicMsg) {
				t.Log(string(debug.Stack()))
				t.Errorf("expected panic to match %q but got %q", expectedPanicRegExp, panicMsg)
			}
		} else {
			t.Log(string(debug.Stack()))
			t.Errorf("expected panic (but there was no panic)")
		}
	}()
	f()
}

func assertError(t *testing.T, expectedErrorRegExp string, e error) {
	t.Helper()
	expected := regexp.MustCompile("(?i)" + expectedErrorRegExp)
	if e == nil {
		t.Errorf("expected error (but error is nil)")
	} else if !expected.MatchString(e.Error()) {
		t.Errorf("expected error to match %q but got %q", expectedErrorRegExp, e.Error())
	}
}

func assertNotNil(t *testing.T, v interface{}) {
	if v == nil {
		t.Helper()
		t.Errorf("nil")
	}
}

func assertBytes(t *testing.T, expect, actual []byte) {
	if !bytes.Equal(actual, expect) {
		t.Helper()
		t.Errorf("expected bytes %q but got %q", expect, actual)
	}
}

func reprValue(v interface{}) string {
	switch v.(type) {
	case []byte, string:
		return fmt.Sprintf("%q", v)
	}
	return fmt.Sprintf("%#v", v)
}

func assertEq(t *testing.T, expect, actual interface{}) {
	if actual != expect {
		t.Helper()
		t.Errorf("expected %s (%T) but got %s (%T)",
			reprValue(expect), expect, reprValue(actual), actual)
	}
}
