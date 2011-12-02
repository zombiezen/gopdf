// Copyright (C) 2011, Ross Light

package pdf

import (
	"fmt"
	"os"
	"strconv"
)

// Name is a PDF name object, which is used as an identifier.
type Name string

func (n Name) String() string {
	return string(n)
}

func (n Name) marshalPDF() ([]byte, os.Error) {
	// TODO: escape characters
	return []byte("/" + n), nil
}

type indirectObject struct {
	Reference
	Object interface{}
}

const (
	objectBegin = " obj "
	objectEnd   = " endobj"
)

func (obj indirectObject) marshalPDF() ([]byte, os.Error) {
	data, err := marshal(obj.Object)
	if err != nil {
		return nil, err
	}

	mn, mg := strconv.Uitoa(obj.Number), strconv.Uitoa(obj.Generation)
	result := make([]byte, 0, len(mn)+1+len(mg)+len(objectBegin)+len(data)+len(objectEnd))
	result = append(result, []byte(mn)...)
	result = append(result, ' ')
	result = append(result, []byte(mg)...)
	result = append(result, []byte(objectBegin)...)
	result = append(result, data...)
	result = append(result, []byte(objectEnd)...)
	return result, nil
}

// Reference holds a PDF indirect reference.
type Reference struct {
	Number     uint
	Generation uint
}

func (ref Reference) marshalPDF() ([]byte, os.Error) {
	return []byte(fmt.Sprintf("%d %d R", ref.Number, ref.Generation)), nil
}
