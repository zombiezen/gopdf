// Copyright (C) 2011, Ross Light

package pdf

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"math"
	"os"
)

// Common page sizes (in typographical points)
const (
	USLetterWidth  = 612
	USLetterHeight = 792

	A4Width  = 11.690
	A4Height = 8.268
)

// Canvas is a two-dimensional drawing region on a single page.  You can obtain
// a canvas once you have created a document.
type Canvas struct {
	doc          *Document
	page         *pageDict
	ref          Reference
	contents     *stream
	imageCounter uint
}

// Document returns the document the canvas is attached to.
func (canvas *Canvas) Document() *Document {
	return canvas.doc
}

// Close flushes the page's stream to the document.  This must be called once
// drawing has completed or else the document will be inconsistent.
func (canvas *Canvas) Close() os.Error {
	return canvas.contents.Close()
}

// SetSize changes the page's media box (the size of the physical medium).
func (canvas *Canvas) SetSize(width, height float32) {
	canvas.page.MediaBox = Rectangle{0, 0, width, height}
}

// SetCrop changes the page's crop box.
func (canvas *Canvas) SetCrop(width, height float32) {
	canvas.page.CropBox = Rectangle{0, 0, width, height}
}

// FillStroke fills then strokes the given path.  This operation has the same
// effect as performing a fill then a stroke, but does not repeat the path in
// the file.
func (canvas *Canvas) FillStroke(p *Path) {
	io.Copy(canvas.contents, &p.buf)
	fmt.Fprintf(canvas.contents, "B\n")
}

// Fill paints the area enclosed by the given path using the current fill color.
func (canvas *Canvas) Fill(p *Path) {
	io.Copy(canvas.contents, &p.buf)
	fmt.Fprintf(canvas.contents, "f\n")
}

// Stroke paints a line along the given path using the current stroke color.
func (canvas *Canvas) Stroke(p *Path) {
	io.Copy(canvas.contents, &p.buf)
	fmt.Fprintf(canvas.contents, "S\n")
}

func (canvas *Canvas) SetLineWidth(w float32) {
	fmt.Fprintf(canvas.contents, "%.*f w\n", marshalFloatPrec, w)
}

const colorFloatPrec = 2

// SetColor changes the current fill color to the given RGB triple (in device
// RGB space).
func (canvas *Canvas) SetColor(r, g, b float32) {
	fmt.Fprintf(canvas.contents, "%.*f %.*f %.*f rg\n", colorFloatPrec, r, colorFloatPrec, g, colorFloatPrec, b)
}

// SetStrokeColor changes the current stroke color to the given RGB triple (in
// device RGB space).
func (canvas *Canvas) SetStrokeColor(r, g, b float32) {
	fmt.Fprintf(canvas.contents, "%.*f %.*f %.*f RG\n", colorFloatPrec, r, colorFloatPrec, g, colorFloatPrec, b)
}

// Push saves a copy of the current graphics state.  The state can later be
// restored using Pop.
func (canvas *Canvas) Push() {
	fmt.Fprintln(canvas.contents, "q")
}

// Pop restores the most recently saved graphics state by popping it from the
// stack.
func (canvas *Canvas) Pop() {
	fmt.Fprintln(canvas.contents, "Q")
}

// Translate moves the canvas's coordinates system by the given offset (in
// typographical points).
func (canvas *Canvas) Translate(x, y float32) {
	fmt.Fprintf(canvas.contents, "1 0 0 1 %.*f %.*f cm\n", marshalFloatPrec, x, marshalFloatPrec, y)
}

// Rotate rotates the canvas's coordinate system by a given angle (in radians).
func (canvas *Canvas) Rotate(theta float32) {
	s, c := math.Sin(float64(theta)), math.Cos(float64(theta))
	fmt.Fprintf(canvas.contents, "%.*f %.*f %.*f %.*f 0 0 cm\n", marshalFloatPrec, c, marshalFloatPrec, s, marshalFloatPrec, -s, marshalFloatPrec, c)
}

// Scale multiplies the canvas's coordinate system by the given scalars.
func (canvas *Canvas) Scale(x, y float32) {
	fmt.Fprintf(canvas.contents, "%.*f 0 0 %.*f 0 0 cm\n", marshalFloatPrec, x, marshalFloatPrec, y)
}

// Transform concatenates a 3x3 matrix with the current transformation matrix.
// The arguments map to values in the matrix as shown below:
//
//  / a b 0 \
//  | c d 0 |
//  \ e f 1 /
//
// For more information, see Section 8.3.4 of ISO 32000-1.
func (canvas *Canvas) Transform(a, b, c, d, e, f float32) {
	fmt.Fprintf(canvas.contents, "%.*f %.*f %.*f %.*f %.*f %.*f cm\n", marshalFloatPrec, a, marshalFloatPrec, b, marshalFloatPrec, c, marshalFloatPrec, d, marshalFloatPrec, e, marshalFloatPrec, f)
}

// DrawText paints a text object onto the canvas.
func (canvas *Canvas) DrawText(text *Text) {
	for fontName := range text.fonts {
		if _, ok := canvas.page.Resources.Font[fontName]; !ok {
			canvas.page.Resources.Font[fontName] = canvas.doc.StandardFont(fontName)
		}
	}
	fmt.Fprintln(canvas.contents, "BT")
	io.Copy(canvas.contents, &text.buf)
	fmt.Fprintln(canvas.contents, "ET")
}

// DrawImage paints a raster image at the given location and scaled to the
// given dimensions (in typographical points).  If you want to render the same
// image multiple times in the same document, use DrawImageReference.
func (canvas *Canvas) DrawImage(img image.Image, x, y, w, h float32) {
	canvas.DrawImageReference(canvas.doc.AddImage(img), x, y, w, h)
}

// DrawImageReference paints the raster image referenced in the document at the
// given location and scaled to the given dimensions (in typographical points).
func (canvas *Canvas) DrawImageReference(ref Reference, x, y, w, h float32) {
	name := canvas.nextImageName()
	canvas.page.Resources.XObject[name] = ref
	marshalledName, _ := name.MarshalPDF()

	canvas.Push()
	canvas.Transform(w, 0, 0, h, x, y)
	fmt.Fprintf(canvas.contents, "%s Do\n", marshalledName)
	canvas.Pop()
}

const anonymousImageFormat = "__image%d__"

func (canvas *Canvas) nextImageName() Name {
	var name Name
	for {
		name = Name(fmt.Sprintf(anonymousImageFormat, canvas.imageCounter))
		canvas.imageCounter++
		if _, ok := canvas.page.Resources.XObject[name]; !ok {
			break
		}
	}
	return name
}

// Path is a shape that can be painted on a canvas.  The zero value is an empty
// path.
type Path struct {
	buf bytes.Buffer
}

// Move begins a new subpath by moving the current point to the given location
// (in typographical points).
func (path *Path) Move(x, y float32) {
	fmt.Fprintf(&path.buf, "%.*f %.*f m\n", marshalFloatPrec, x, marshalFloatPrec, y)
}

// Line appends a line segment from the current point to the given location (in
// typographical points).
func (path *Path) Line(x, y float32) {
	fmt.Fprintf(&path.buf, "%.*f %.*f l\n", marshalFloatPrec, x, marshalFloatPrec, y)
}

// Close appends a line segment from the current point to the starting point of
// the subpath.
func (path *Path) Close() {
	fmt.Fprintf(&path.buf, "h\n")
}
