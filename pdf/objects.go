// Copyright (C) 2011, Ross Light

package pdf

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
)

// Name is a PDF Name object.
type Name string

func (n Name) String() string {
	return string(n)
}

func (n Name) MarshalPDF() ([]byte, os.Error) {
	// TODO: escape characters
	return []byte("/" + n), nil
}

// stream is a blob of data.
type stream struct {
	Dictionary map[Name]interface{}
	Bytes      []byte
}

const (
	streamBegin = "stream\r\n"
	streamEnd   = "\r\nendstream"
)

func (s stream) MarshalPDF() ([]byte, os.Error) {
	var b bytes.Buffer

	// TODO: Force Length key
	mdict, err := Marshal(s.Dictionary)
	if err != nil {
		return nil, err
	}
	b.Write(mdict)
	b.WriteString(streamBegin)
	b.Write(s.Bytes)
	b.WriteString(streamEnd)
	return b.Bytes(), nil
}

type indirectObject struct {
	Reference
	Object interface{}
}

const (
	objectBegin = " obj "
	objectEnd   = " endobj"
)

func (obj indirectObject) MarshalPDF() ([]byte, os.Error) {
	data, err := Marshal(obj.Object)
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

// Reference represents a PDF indirect reference.
type Reference struct {
	Number     uint
	Generation uint
}

func (ref Reference) MarshalPDF() ([]byte, os.Error) {
	return []byte(fmt.Sprintf("%d %d R", ref.Number, ref.Generation)), nil
}
