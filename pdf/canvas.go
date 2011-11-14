// Copyright (C) 2011, Ross Light

package pdf

import (
	"bytes"
	"fmt"
	"io"
)

type Canvas struct {
	doc      *Document
	page     *pageDict
	ref      Reference
	contents *stream
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

func (canvas *Canvas) SetColor(r, g, b float64) {
	fmt.Fprintf(canvas.contents, "%.2f %.2f %.2f rg\n", r, g, b)
}

func (canvas *Canvas) SetStrokeColor(r, g, b float64) {
	fmt.Fprintf(canvas.contents, "%.2f %.2f %.2f RG\n", r, g, b)
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
