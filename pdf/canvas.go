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

type Canvas struct {
	doc          *Document
	page         *pageDict
	ref          Reference
	contents     *stream
	imageCounter uint
}

func (canvas *Canvas) Close() os.Error {
	return canvas.contents.Close()
}

// SetSize changes the page's media box.
func (canvas *Canvas) SetSize(width, height int) {
	canvas.page.MediaBox = Rectangle{0, 0, width, height}
}

// SetCrop changes the page's crop box.
func (canvas *Canvas) SetCrop(width, height int) {
	canvas.page.CropBox = Rectangle{0, 0, width, height}
}

func (canvas *Canvas) FillStroke(p *Path) {
	io.Copy(canvas.contents, &p.buf)
	fmt.Fprintf(canvas.contents, "B\n")
}

func (canvas *Canvas) Fill(p *Path) {
	io.Copy(canvas.contents, &p.buf)
	fmt.Fprintf(canvas.contents, "f\n")
}

func (canvas *Canvas) Stroke(p *Path) {
	io.Copy(canvas.contents, &p.buf)
	fmt.Fprintf(canvas.contents, "S\n")
}

func (canvas *Canvas) SetLineWidth(w int) {
	fmt.Fprintf(canvas.contents, "%d w\n")
}

func (canvas *Canvas) SetColor(r, g, b float64) {
	fmt.Fprintf(canvas.contents, "%.2f %.2f %.2f rg\n", r, g, b)
}

func (canvas *Canvas) SetStrokeColor(r, g, b float64) {
	fmt.Fprintf(canvas.contents, "%.2f %.2f %.2f RG\n", r, g, b)
}

func (canvas *Canvas) Push() {
	fmt.Fprintln(canvas.contents, "q")
}

func (canvas *Canvas) Pop() {
	fmt.Fprintln(canvas.contents, "Q")
}

func (canvas *Canvas) Translate(x, y int) {
	fmt.Fprintf(canvas.contents, "1 0 0 1 %d %d cm\n", x, y)
}

// Rotate rotates the canvas's coordinate system by a given angle (in radians).
func (canvas *Canvas) Rotate(theta float64) {
	s, c := math.Sin(theta), math.Cos(theta)
	fmt.Fprintf(canvas.contents, "%f %f %f %f 0 0 cm\n", c, s, -s, c)
}

func (canvas *Canvas) Scale(x, y float64) {
	fmt.Fprintf(canvas.contents, "%f 0 0 %f 0 0 cm\n", x, y)
}

func (canvas *Canvas) Transform(a, b, c, d, e, f float64) {
	fmt.Fprintf(canvas.contents, "%f %f %f %f %f %f cm\n", a, b, c, d, e, f)
}

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

func (canvas *Canvas) DrawImage(img image.Image, x, y, w, h int) {
	canvas.DrawImageReference(canvas.doc.AddImage(img), x, y, w, h)
}

func (canvas *Canvas) DrawImageReference(ref Reference, x, y, w, h int) {
	name := canvas.nextImageName()
	canvas.page.Resources.XObject[name] = ref
	marshalledName, _ := name.MarshalPDF()

	canvas.Push()
	canvas.Transform(float64(w), 0, 0, float64(h), float64(x), float64(y))
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

type Path struct {
	buf bytes.Buffer
}

func (path *Path) Move(x, y int) {
	fmt.Fprintf(&path.buf, "%d %d m\n", x, y)
}

func (path *Path) Line(x, y int) {
	fmt.Fprintf(&path.buf, "%d %d l\n", x, y)
}

func (path *Path) Close() {
	fmt.Fprintf(&path.buf, "h\n")
}

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
