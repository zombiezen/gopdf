// Copyright (C) 2011, Ross Light

package pdf

import (
	"bytes"
	"fmt"
	"io"
	"math"
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

func (canvas *Canvas) Scale(x, y float64) {
	fmt.Fprintf(canvas.contents, "%f 0 0 %f 0 0 cm\n", x, y)
}

// Rotate rotates the canvas's coordinate system by a given angle (in radians).
func (canvas *Canvas) Rotate(theta float64) {
	s, c := math.Sin(theta), math.Cos(theta)
	fmt.Fprintf(canvas.contents, "%f %f %f %f 0 0 cm\n", c, s, -s, c)
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
