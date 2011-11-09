package pdf

import (
	"testing"
)

func TestQuote(t *testing.T) {
	tests := []struct{
		Input, Expected string
	}{
		{"", "()"},
		{"This is a string", "(This is a string)"},
		{"Strings may contain newlines\nand such.", "(Strings may contain newlines\nand such.)"},
		{"Escape (this).", `(Escape \(this\).)`},
	}
	for i, tt := range tests {
		result := quote(tt.Input)
		if result != tt.Expected {
			t.Errorf("%d. quote(%q) != %q (got %q)", i, tt.Input, tt.Expected, result)
		}
	}
}

type marshalTest struct {
	Marshaler
	Expected string
}

var marshalTests = []marshalTest{
	{name(""), "/"},
	{name("foo"), "/foo"},
}

func TestMarshal(t *testing.T) {
	for i, tt := range marshalTests {
		result, err := tt.Marshaler.MarshalPDF()
		switch {
		case err != nil:
			t.Errorf("%d. %#v.MarshalPDF() error: %v", i, tt.Marshaler, err)
		case string(result) != tt.Expected:
			t.Errorf("%d. %#v.MarshalPDF() != %q (got %q)", i, tt.Marshaler, tt.Expected, result)
		}
	}
}
