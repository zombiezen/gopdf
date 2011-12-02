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

func (n name) marshalPDF(dst []byte) ([]byte, os.Error) {
	// TODO: escape characters
	dst = append(dst, '/')
	return append(dst, []byte(n)...), nil
}

type indirectObject struct {
	Reference
	Object interface{}
}

const (
	objectBegin = " obj "
	objectEnd   = " endobj"
)

func (obj indirectObject) marshalPDF(dst []byte) ([]byte, os.Error) {
	var err os.Error
	mn, mg := strconv.Uitoa(obj.Number), strconv.Uitoa(obj.Generation)
	dst = append(dst, []byte(mn)...)
	dst = append(dst, ' ')
	dst = append(dst, []byte(mg)...)
	dst = append(dst, []byte(objectBegin)...)
	if dst, err = marshal(dst, obj.Object); err != nil {
		return nil, err
	}
	dst = append(dst, []byte(objectEnd)...)
	return dst, nil
}

// Reference holds a PDF indirect reference.
type Reference struct {
	Number     uint
	Generation uint
}

func (ref Reference) marshalPDF() ([]byte, os.Error) {
	return append(dst, []byte(fmt.Sprintf("%d %d R", ref.Number, ref.Generation))...), nil
}
