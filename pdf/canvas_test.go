// Copyright (C) 2011, Ross Light

package pdf

import (
	"testing"
)

const pathExpectedOutput = `12 34 m
-56 78 l
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

const textExpectedOutput = `/Helvetica 12 Tf
14 TL
(Hello, World!) Tj
T*
(This is SPARTA!!1!) Tj
`

func TestText(t *testing.T) {
	text := new(Text)
	text.SetFont(Helvetica, 12)
	text.SetLeading(14)
	text.Text("Hello, World!")
	text.NextLine()
	text.Text("This is SPARTA!!1!")

	if text.buf.String() != textExpectedOutput {
		t.Errorf("Output was %q, expected %q", text.buf.String(), textExpectedOutput)
	}

	if len(text.fonts) == 1 {
		if !text.fonts[Helvetica] {
			t.Error("Helvetica missing from fonts")
		}
	} else {
		t.Errorf("Got %d fonts, expected %d", len(text.fonts), 1)
	}
}
