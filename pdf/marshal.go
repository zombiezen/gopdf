package pdf

import (
	"bytes"
	"fmt"
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

func (n name) MarshalPDF() ([]byte, os.Error) {
	// TODO: escape characters
	return []byte("/" + n), nil
}

var arrayMarshalChars = []byte{'[', ' ', ']'}

func (a array) MarshalPDF() ([]byte, os.Error) {
	marshalled := make([][]byte, len(a)+2)
	marshalled[0] = arrayMarshalChars[0:1]
	marshalled[len(marshalled)-1] = arrayMarshalChars[2:3]

	for i, obj := range a {
		if m, ok := obj.(Marshaler); ok {
			var err os.Error
			marshalled[i], err = m.MarshalPDF()
			if err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("array element %d does not implement Marshaler", i)
		}
	}
	return bytes.Join(marshalled, arrayMarshalChars[1:2]), nil
}

var dictionaryMarshalChars = []byte{'<', '<', ' ', '>', '>'}

func (d dictionary) MarshalPDF() ([]byte, os.Error) {
	marshalled := make([][]byte, 0, len(d)*2+2)
	marshalled = append(marshalled, dictionaryMarshalChars[0:2])

	for k, obj := range d {
		// Marshal key
		mk, err := k.MarshalPDF()
		if err != nil {
			return nil, err
		}

		// Marshal value
		if m, ok := obj.(Marshaler); ok {
			mobj, err := m.MarshalPDF()
			if err != nil {
				return nil, err
			}
			marshalled = append(marshalled, mk, mobj)
		} else {
			return nil, fmt.Errorf("dictionary element %s does not implement Marshaler", k)
		}
	}

	marshalled = append(marshalled, dictionaryMarshalChars[3:5])
	return bytes.Join(marshalled, dictionaryMarshalChars[2:3]), nil
}

const (
	streamBegin = "stream\r\n"
	streamEnd   = "\r\nendstream"
)

func (s stream) MarshalPDF() ([]byte, os.Error) {
	var b bytes.Buffer

	// TODO: Force Length key
	mdict, err := s.Dictionary.MarshalPDF()
	if err != nil {
		return nil, err
	}
	b.Write(mdict)
	b.WriteString(streamBegin)
	b.Write(s.Bytes)
	b.WriteString(streamEnd)
	return b.Bytes(), nil
}

const (
	objectBegin = "obj "
	objectEnd   = " endobj"
)

func (obj indirectObject) MarshalPDF() ([]byte, os.Error) {
	m, ok := obj.Object.(Marshaler)
	if !ok {
		return nil, fmt.Errorf("indirect object %d %d does not implement Marshaler", obj.Number, obj.Generation)
	}
	data, err := m.MarshalPDF()
	if err != nil {
		return nil, err
	}

	mn, mg := strconv.Uitoa(obj.Number), strconv.Uitoa(obj.Generation)
	result := make([]byte, 0, len(mn)+1+len(mg)+len(objectBegin)+len(data)+len(objectEnd))
	result = append(result, mn...)
	result = append(result, ' ')
	result = append(result, mg...)
	result = append(result, objectBegin...)
	result = append(result, data...)
	result = append(result, objectEnd...)
	return result, nil
}
