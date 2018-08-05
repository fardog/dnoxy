package cmd

import "testing"

func TestValues(t *testing.T) {
	v := make(Values, 0)
	v.Set("one")
	v.Set("two")

	if l := len(v); l != 2 {
		t.Fatalf("expected len == 2, was %d", l)
	}

	for i, e := range []string{"one", "two"} {
		if a := v[i]; a != e {
			t.Errorf("expected value at position %d to be %v, was %v", i, e, a)
		}
	}

	if s := v.String(); s != "one,two" {
		t.Errorf("expected string value to be one,two; was %v", s)
	}
}
