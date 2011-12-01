// Copyright (C) 2011, Ross Light

package pdf

import (
	"fmt"
	"strconv"
)

// Name is a PDF name object, which is used as an identifier.
type Name string

func (n Name) String() string {
	return string(n)
}

func (n Name) MarshalPDF() ([]byte, error) {
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

func (obj indirectObject) MarshalPDF() ([]byte, error) {
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

// Reference holds a PDF indirect reference.
type Reference struct {
	Number     uint
	Generation uint
}

func (ref Reference) MarshalPDF() ([]byte, error) {
	return []byte(fmt.Sprintf("%d %d R", ref.Number, ref.Generation)), nil
}
