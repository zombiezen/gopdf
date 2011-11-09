package pdf

import (
	"testing"
)

type marshalTest struct {
	Value    interface{}
	Expected string
}

var marshalTests = []marshalTest{
	{"", "()"},
	{"This is a string", "(This is a string)"},
	{"Strings may contain newlines\nand such.", "(Strings may contain newlines\nand such.)"},
	{"Escape (this).", `(Escape \(this\).)`},
	{int(123), "123"},
	{int(-321), "-321"},
	{float64(-3.141599), "-3.14160"},
	{float64(1e9), "1000000000.00000"},
	{name(""), "/"},
	{name("foo"), "/foo"},
	{[]interface{}{}, `[ ]`},
	{[]string{"foo", "(parens)"}, `[ (foo) (\(parens\)) ]`},
	{map[name]string{}, `<< >>`},
	{map[name]string{name("foo"): "bar"}, `<< /foo (bar) >>`},
}

func TestMarshal(t *testing.T) {
	for i, tt := range marshalTests {
		result, err := Marshal(tt.Value)
		switch {
		case err != nil:
			t.Errorf("%d. Marshal(%#v) error: %v", i, tt.Value, err)
		case string(result) != tt.Expected:
			t.Errorf("%d. Marshal(%#v) != %q (got %q)", i, tt.Value, tt.Expected, result)
		}
	}
}
