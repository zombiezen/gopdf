// Copyright (C) 2011, Ross Light

package pdf

import (
	"fmt"
	"os"
	"strconv"
)

// name is a PDF name object, which is used as an identifier.
type name string

func (n name) String() string {
	return string(n)
}

func (n name) marshalPDF() ([]byte, os.Error) {
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
	result = append(result, mn...)
	result = append(result, ' ')
	result = append(result, mg...)
	result = append(result, objectBegin...)
	result = append(result, data...)
	result = append(result, objectEnd...)
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
