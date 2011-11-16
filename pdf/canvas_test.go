// Copyright (C) 2011, Ross Light

package pdf

import (
	"testing"
)

const pathExpectedOutput = `12.00000 34.00000 m
-56.00000 78.00000 l
h
`

func TestPath(t *testing.T) {
	path := new(Path)
	path.Move(12, 34)
	path.Line(-56, 78)
	path.Close()

	if path.buf.String() != pathExpectedOutput {
		t.Errorf("Output was %q, expected %q", path.buf.String(), pathExpectedOutput)
	}
}
