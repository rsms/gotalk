package gotalk
import (
  "testing"
)

func assertEq(t *testing.T, actual, expect interface{}) {
  if actual != expect {
    if _, ok := expect.(string); ok {
      t.Errorf("Expected `%q %T` but got `%q %T`\n", expect,expect, actual,actual)
    } else {
      t.Errorf("Expected `%v %T` but got `%v %T`\n", expect,expect, actual,actual)
    }
  }
}
