// generate-image-testdata generates the canonical pdf files for pdf/image_test.go.
package main

import (
	"bitbucket.org/zombiezen/gopdf/pdf"
	"flag"
	_ "golang.org/x/image/bmp"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path"
)

var (
	imageDir = flag.String("input_image_dir", "../pdf/testdata", "Directory of input image files")
	pdfDir   = flag.String("output_pdf_dir", "../pdf/testdata", "Directory of output pdf files")
)

var imageFiles = []string{
	"suzanne.bmp",
	"suzanne.jpg",
	"suzanne.png",
}

func readImage(filename string) (image.Image, error) {
	f, err := os.Open(path.Join(*imageDir, filename))
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func writePDF(img image.Image) (*pdf.Document, error) {
	doc := pdf.New()
	page := doc.NewPage(12*pdf.Inch, 12*pdf.Inch)
	rect := pdf.Rectangle{
		Min: pdf.Point{1 * pdf.Inch, 1 * pdf.Inch},
		Max: pdf.Point{11 * pdf.Inch, 11 * pdf.Inch},
	}
	page.DrawImage(img, rect)
	err := page.Close()
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func encodePDF(doc *pdf.Document, filename string) error {
	f, err := os.Create(path.Join(*pdfDir, filename))
	if err != nil {
		return err
	}
	defer f.Close()
	err = doc.Encode(f)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	for _, imageFile := range imageFiles {
		img, err := readImage(imageFile)
		if err != nil {
			panic(err)
		}
		doc, err := writePDF(img)
		if err != nil {
			panic(err)
		}
		err = encodePDF(doc, imageFile+".pdf")
		if err != nil {
			panic(err)
		}
	}
}
