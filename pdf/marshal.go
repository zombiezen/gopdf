// Copyright (C) 2011, Ross Light

package pdf

import (
	"bytes"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// A Marshaler can produce a PDF object.
type marshaler interface {
	marshalPDF() ([]byte, os.Error)
}

// marshal returns the PDF encoding of v.
//
// If the value implements the marshaler interface, then its marshalPDF method
// is called.  ints, strings, and floats will be marshalled according to the PDF
// standard.
func marshal(v interface{}) ([]byte, os.Error) {
	state := new(marshalState)
	if err := state.marshalValue(reflect.ValueOf(v)); err != nil {
		return nil, err
	}
	return state.Bytes(), nil
}

type marshalState struct {
	bytes.Buffer
}

const marshalFloatPrec = 5

func (state *marshalState) marshalValue(v reflect.Value) os.Error {
	if !v.IsValid() {
		state.WriteString("null")
		return nil
	}

	if m, ok := v.Interface().(marshaler); ok {
		b, err := m.marshalPDF()
		if err != nil {
			return err
		}
		state.Write(b)
		return nil
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		state.WriteString(strconv.Itoa64(v.Int()))
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		state.WriteString(strconv.Uitoa64(v.Uint()))
		return nil
	case reflect.Float32, reflect.Float64:
		state.WriteString(strconv.Ftoa64(v.Float(), 'f', marshalFloatPrec))
		return nil
	case reflect.String:
		state.WriteString(quote(v.String()))
		return nil
	case reflect.Ptr, reflect.Interface:
		return state.marshalValue(v.Elem())
	case reflect.Array, reflect.Slice:
		return state.marshalSlice(v)
	case reflect.Map:
		return state.marshalDictionary(v)
	case reflect.Struct:
		return state.marshalStruct(v)
	}

	return os.NewError("pdf: unsupported type: " + v.Type().String())
}

// quote escapes a string and returns a PDF string literal.
func quote(s string) string {
	r := strings.NewReplacer(
		"\r", `\r`,
		"\t", `\t`,
		"\b", `\b`,
		"\f", `\f`,
		"(", `\(`,
		")", `\)`,
		`\`, `\\`,
	)
	return "(" + r.Replace(s) + ")"
}

func (state *marshalState) marshalSlice(v reflect.Value) os.Error {
	state.WriteString("[ ")
	for i := 0; i < v.Len(); i++ {
		if err := state.marshalValue(v.Index(i)); err != nil {
			return err
		}
		state.WriteByte(' ')
	}
	state.WriteString("]")
	return nil
}

func (state *marshalState) marshalDictionary(v reflect.Value) os.Error {
	if v.Type().Key() != reflect.TypeOf(name("")) {
		return os.NewError("pdf: cannot marshal dictionary with key type: " + v.Type().Key().String())
	}

	state.WriteString("<< ")
	for _, k := range v.MapKeys() {
		// Marshal key
		mk, err := k.Interface().(name).marshalPDF()
		if err != nil {
			return err
		}
		state.Write(mk)
		state.WriteByte(' ')

		// Marshal value
		if err := state.marshalValue(v.MapIndex(k)); err != nil {
			return err
		}
		state.WriteByte(' ')
	}
	state.WriteString(">>")
	return nil
}

func (state *marshalState) marshalStruct(v reflect.Value) os.Error {
	state.WriteString("<< ")
	t := v.Type()
	n := v.NumField()
	for i := 0; i < n; i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}

		tag, omitEmpty := f.Name, false
		if tv := f.Tag.Get("pdf"); tv != "" {
			if tv == "-" {
				continue
			}

			name, options := parseTag(tv)
			if name != "" {
				tag = name
			}
			omitEmpty = options.Contains("omitempty")
		}

		fieldValue := v.Field(i)
		if omitEmpty && isEmptyValue(fieldValue) {
			continue
		}

		// Marshal key
		mk, err := name(tag).marshalPDF()
		if err != nil {
			return err
		}
		state.Write(mk)
		state.WriteByte(' ')

		// Marshal value
		if err := state.marshalValue(fieldValue); err != nil {
			return err
		}
		state.WriteByte(' ')
	}
	state.WriteString(">>")
	return nil
}

type tagOptions []string

func parseTag(tag string) (name string, options tagOptions) {
	result := strings.Split(tag, ",")
	return result[0], tagOptions(result[1:])
}

func (options tagOptions) Contains(opt string) bool {
	for _, o := range options {
		if opt == o {
			return true
		}
	}
	return false
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}
