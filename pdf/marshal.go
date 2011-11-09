package pdf

import (
	"bytes"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type Marshaler interface {
	MarshalPDF() ([]byte, os.Error)
}

const marshalFloatPrec = 5

// Marshal returns the PDF encoding of v.
//
// If the value implements the Marshaler interface, then its MarshalPDF method
// is called.  ints, strings, and floats will be marshalled according to the PDF
// standard.
func Marshal(v interface{}) ([]byte, os.Error) {
	return marshalValue(reflect.ValueOf(v))
}

func marshalValue(v reflect.Value) ([]byte, os.Error) {
	if m, ok := v.Interface().(Marshaler); ok {
		return m.MarshalPDF()
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return []byte(strconv.Itoa64(v.Int())), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return []byte(strconv.Uitoa64(v.Uint())), nil
	case reflect.Float32, reflect.Float64:
		return []byte(strconv.Ftoa64(v.Float(), 'f', marshalFloatPrec)), nil
	case reflect.String:
		return []byte(quote(v.String())), nil
	case reflect.Ptr:
		return marshalValue(v.Elem())
	case reflect.Array, reflect.Slice:
		return marshalSlice(v)
	case reflect.Map:
		return marshalDictionary(v)
	}

	return nil, os.NewError("pdf: unsupported type: " + v.Type().String())
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

var arrayMarshalChars = []byte{'[', ' ', ']'}

func marshalSlice(v reflect.Value) ([]byte, os.Error) {
	marshalled := make([][]byte, 0, v.Len()+2)
	marshalled = append(marshalled, arrayMarshalChars[0:1])
	for i := 0; i < v.Len(); i++ {
		m, err := marshalValue(v.Index(i))
		if err != nil {
			return nil, err
		}
		marshalled = append(marshalled, m)
	}
	marshalled = append(marshalled, arrayMarshalChars[2:3])
	return bytes.Join(marshalled, arrayMarshalChars[1:2]), nil
}

var dictionaryMarshalChars = []byte{'<', '<', ' ', '>', '>'}

func marshalDictionary(v reflect.Value) ([]byte, os.Error) {
	if v.Type().Key() != reflect.TypeOf(name("")) {
		return nil, os.NewError("pdf: cannot marshal dictionary with key type: " + v.Type().Key().String())
	}

	marshalled := make([][]byte, 0, v.Len()*2+2)
	marshalled = append(marshalled, dictionaryMarshalChars[0:2])

	for _, k := range v.MapKeys() {
		// Marshal key
		mk, err := k.Interface().(name).MarshalPDF()
		if err != nil {
			return nil, err
		}

		// Marshal value
		mobj, err := marshalValue(v.MapIndex(k))
		if err != nil {
			return nil, err
		}

		marshalled = append(marshalled, mk, mobj)
	}

	marshalled = append(marshalled, dictionaryMarshalChars[3:5])
	return bytes.Join(marshalled, dictionaryMarshalChars[2:3]), nil
}
