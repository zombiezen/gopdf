// Copyright (C) 2011, Ross Light

package pdf

import (
	"bytes"
)

// Text is a PDF text object.  The zero value is an empty text object.
type Text struct {
	buf   bytes.Buffer
	fonts map[Name]bool

	x, y        float32
	currFont    Name
	currSize    float32
	currLeading float32
}

// Text adds a string to the text object.
func (text *Text) Text(s string) {
	writeCommand(&text.buf, "Tj", s)
	if widths := getFontWidths(text.currFont); widths != nil {
		text.x += computeStringWidth(s, widths, text.currSize)
	}
}

// SetFont changes the current font to either a standard font or a font
// declared in the canvas.
func (text *Text) SetFont(name Name, size float32) {
	if text.fonts == nil {
		text.fonts = make(map[Name]bool)
	}
	text.fonts[name] = true
	text.currFont, text.currSize = name, size
	writeCommand(&text.buf, "Tf", name, size)
}

// SetLeading changes the amount of space between lines.
func (text *Text) SetLeading(leading float32) {
	writeCommand(&text.buf, "TL", leading)
	text.currLeading = leading
}

// NextLine advances the current text position to the next line, based on the
// current leading.
func (text *Text) NextLine() {
	writeCommand(&text.buf, "T*")
	text.x = 0
	text.y -= text.currLeading
}

// NextLineOffset advances the current text position by the given offset (in
// typographical points).
func (text *Text) NextLineOffset(tx, ty float32) {
	writeCommand(&text.buf, "Td", tx, ty)
	text.x = tx
	text.y -= ty
}

// X returns the current x position of the text cursor.
func (text *Text) X() float32 {
	return text.x
}

// Y returns the current y position of the text cursor.
func (text *Text) Y() float32 {
	return text.y
}

// Standard 14 fonts
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

func getFontWidths(fontName Name) []uint16 {
	switch fontName {
	case Courier:
		return courierWidths
	case CourierBold:
		return courierBoldWidths
	case CourierOblique:
		return courierObliqueWidths
	case CourierBoldOblique:
		return courierBoldObliqueWidths
	case Helvetica:
		return helveticaWidths
	case HelveticaBold:
		return helveticaBoldWidths
	case HelveticaOblique:
		return helveticaObliqueWidths
	case HelveticaBoldOblique:
		return helveticaBoldObliqueWidths
	case Symbol:
		return symbolWidths
	case Times:
		return timesRomanWidths
	case TimesBold:
		return timesBoldWidths
	case TimesItalic:
		return timesItalicWidths
	case TimesBoldItalic:
		return timesBoldItalicWidths
	case ZapfDingbats:
		return zapfDingbatsWidths
	}
	return nil
}

func computeStringWidth(s string, widths []uint16, fontSize float32) float32 {
	width := float32(0)
	for _, r := range s {
		if r < len(widths) {
			width += float32(widths[r])
		}
	}
	return width * fontSize / 1000
}
