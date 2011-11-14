// Copyright (C) 2011, Ross Light

package pdf

import (
	"testing"
)

type marshalTest struct {
	Value    interface{}
	Expected string
}

type fooStruct struct {
	Size    int64
	Params  map[Name]string
	NotHere string  `pdf:"-"`
	Rename  float64 `pdf:"Pi"`
}

var marshalTests = []marshalTest{
	{nil, "null"},
	{"", "()"},
	{"This is a string", "(This is a string)"},
	{"Strings may contain newlines\nand such.", "(Strings may contain newlines\nand such.)"},
	{"Escape (this).", `(Escape \(this\).)`},
	{int(123), "123"},
	{int(-321), "-321"},
	{float64(-3.141599), "-3.14160"},
	{float64(1e9), "1000000000.00000"},
	{Name(""), "/"},
	{Name("foo"), "/foo"},
	{[]interface{}{}, `[ ]`},
	{[]string{"foo", "(parens)"}, `[ (foo) (\(parens\)) ]`},
	{map[Name]string{}, `<< >>`},
	{map[Name]string{Name("foo"): "bar"}, `<< /foo (bar) >>`},
	{indirectObject{Reference{42, 0}, "foo"}, `42 0 obj (foo) endobj`},
	{Reference{42, 0}, `42 0 R`},
	{
		fooStruct{
			Size:    42,
			Params:  map[Name]string{Name("this"): "that"},
			NotHere: "XXX",
			Rename:  3.141592,
		},
		`<< /Size 42 /Params << /this (that) >> /Pi 3.14159 >>`,
	},
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
