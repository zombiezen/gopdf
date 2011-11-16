// Copyright (C) 2011, Ross Light

package pdf

import (
	"bytes"
	"fmt"
)

// Text is a PDF text object.
type Text struct {
	buf   bytes.Buffer
	fonts map[Name]bool
}

func (text *Text) Show(s string) {
	fmt.Fprintf(&text.buf, "%s Tj\n", quote(s))
}

func (text *Text) SetFont(name Name, size int) {
	nameData, err := name.MarshalPDF()
	if err != nil {
		// TODO: log error?
		return
	}

	if text.fonts == nil {
		text.fonts = make(map[Name]bool)
	}
	text.fonts[name] = true
	fmt.Fprintf(&text.buf, "%s %d Tf\n", nameData, size)
}

func (text *Text) SetLeading(leading int) {
	fmt.Fprintf(&text.buf, "%d TL\n", leading)
}

func (text *Text) NextLine() {
	fmt.Fprintln(&text.buf, "T*")
}

func (text *Text) NextLineOffset(tx, ty int) {
	fmt.Fprintf(&text.buf, "%d %d Td\n", tx, ty)
}

// The PDF standard 14 fonts
const (
	Courier            Name = "Courier"
	CourierBold        Name = "Courier-Bold"
	CourierOblique     Name = "Courier-Oblique"
	CourierBoldOblique Name = "Courier-BoldOblique"

	Helvetica            Name = "Helvetica"
	HelveticaBold        Name = "Helvetica-Bold"
	HelveticaOblique     Name = "Helvetica-Oblique"
	HelveticaBoldOblique Name = "Helvetica-BoldOblique"

	Symbol Name = "Symbol"

	Times           Name = "Times-Roman"
	TimesBold       Name = "Times-Bold"
	TimesItalic     Name = "Times-Italic"
	TimesBoldItalic Name = "Times-BoldItalic"

	ZapfDingbats Name = "ZapfDingbats"
)
