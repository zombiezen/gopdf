// Copyright (C) 2011, Ross Light

/*
	Package pdf implements a Portable Document Format writer, as defined in ISO 32000-1.

	Most all dimensional units in this package are in typographical points,
	which are defined to be 1/72 of an inch.  If you want to use different
	units, then you can scale the canvas.

	An example of basic usage:

		package main

		import (
			"bitbucket.org/zombiezen/gopdf/pdf"
			"fmt"
			"os"
		)

		func main() {
			doc := pdf.New()
			canvas := doc.NewPage(612, 792) // standard US letter
			canvas.Translate(100, 100)

			path := new(pdf.Path)
			path.Move(0, 0)
			path.Line(100, 0)
			canvas.Stroke(path)

			text := new(pdf.Text)
			text.SetFont(pdf.Helvetica, 14)
			text.Text("Hello, World!")
			canvas.DrawText(text)

			canvas.Close()

			err := doc.Encode(os.Stdout)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
*/
package pdf
