// Copyright (C) 2011, Ross Light

package pdf

import (
	"image"
	"image/ycbcr"
	"io"
	"os"
)

type Document struct {
	Encoder
	catalog *catalog
	pages   []indirectObject
	fonts   map[Name]Reference
}

// New creates a document value.
func New() *Document {
	doc := new(Document)
	doc.catalog = &catalog{
		Type: catalogType,
	}
	doc.Root = doc.Add(doc.catalog)
	doc.fonts = make(map[Name]Reference, 14)
	return doc
}

func (doc *Document) NewPage(width, height int) *Canvas {
	page := &pageDict{
		Type:     pageType,
		MediaBox: Rectangle{0, 0, width, height},
		CropBox:  Rectangle{0, 0, width, height},
		Resources: resources{
			ProcSet: []Name{pdfProcSet, textProcSet, imageCProcSet},
			Font:    make(map[Name]interface{}),
			XObject: make(map[Name]interface{}),
		},
	}
	pageRef := doc.Add(page)
	doc.pages = append(doc.pages, indirectObject{pageRef, page})

	stream := newStream(streamFlateDecode)
	page.Contents = doc.Add(stream)

	return &Canvas{
		doc:      doc,
		page:     page,
		ref:      pageRef,
		contents: stream,
	}
}

// StandardFont returns a reference to a standard font dictionary.  If there is
// no font dictionary for the font in the document yet, it is added
// automatically.
func (doc *Document) StandardFont(name Name) Reference {
	if ref, ok := doc.fonts[name]; ok {
		return ref
	}

	// TODO: check name is standard?
	ref := doc.Add(standardFontDict{
		Type:     fontType,
		Subtype:  fontType1Subtype,
		BaseFont: name,
	})
	doc.fonts[name] = ref
	return ref
}

func (doc *Document) AddImage(img image.Image) Reference {
	bd := img.Bounds()
	st := newImageStream(streamFlateDecode, bd.Dx(), bd.Dy())
	defer st.Close()

	switch i := img.(type) {
	case *image.RGBA:
		encodeRGBAStream(st, i)
	case *ycbcr.YCbCr:
		encodeYCbCrStream(st, i)
	default:
		encodeImageStream(st, i)
	}
	return doc.Add(st)
}

// encodeImageStream writes RGB data from an image in PDF format.
func encodeImageStream(w io.Writer, img image.Image) os.Error {
	// TODO: alpha

	bd := img.Bounds()
	var buf [3]byte
	for y := bd.Min.Y; y < bd.Max.Y; y++ {
		for x := bd.Min.X; x < bd.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			buf[0] = byte(r & 0xff)
			buf[1] = byte(g & 0xff)
			buf[2] = byte(b & 0xff)
			if _, err := w.Write(buf[:]); err != nil {
				return err
			}
		}
	}
	return nil
}

func encodeRGBAStream(w io.Writer, img *image.RGBA) os.Error {
	for i := 0; i < len(img.Pix); i += 4 {
		if _, err := w.Write(img.Pix[:3]); err != nil {
			return err
		}
	}
	return nil
}

func encodeYCbCrStream(w io.Writer, img *ycbcr.YCbCr) os.Error {
	var buf [3]byte
	var yy, cb, cr byte
	var i, j int
	dx, dy := img.Rect.Dx(), img.Rect.Dy()
	for y := 0; y < dy; y++ {
		for x := 0; x < dx; x++ {
			i, j = x, y
			switch img.SubsampleRatio {
			case ycbcr.SubsampleRatio420:
				j /= 2
				fallthrough
			case ycbcr.SubsampleRatio422:
				i /= 2
			}
			yy = img.Y[y*img.YStride+x]
			cb = img.Cb[j*img.CStride+i]
			cr = img.Cr[j*img.CStride+i]

			buf[0], buf[1], buf[2] = ycbcr.YCbCrToRGB(yy, cb, cr)
			if _, err := w.Write(buf[:]); err != nil {
				return err
			}
		}
	}
	return nil
}

func (doc *Document) Encode(w io.Writer) os.Error {
	pageRoot := &pageRootNode{
		Type:  pageNodeType,
		Count: len(doc.pages),
	}
	doc.catalog.Pages = doc.Add(pageRoot)
	for _, p := range doc.pages {
		page := p.Object.(*pageDict)
		page.Parent = doc.catalog.Pages
		pageRoot.Kids = append(pageRoot.Kids, p.Reference)
	}

	return doc.Encoder.Encode(w)
}

const (
	catalogType  Name = "Catalog"
	pageNodeType Name = "Pages"
	pageType     Name = "Page"
	fontType     Name = "Font"
	xobjectType  Name = "XObject"
)

const (
	imageSubtype Name = "Image"

	fontType1Subtype Name = "Type1"
)

type catalog struct {
	Type  Name
	Pages Reference
}

type pageRootNode struct {
	Type  Name
	Kids  []Reference
	Count int
}

type pageNode struct {
	Type   Name
	Parent Reference
	Kids   []Reference
	Count  int
}

type pageDict struct {
	Type      Name
	Parent    Reference
	Resources resources
	MediaBox  Rectangle
	CropBox   Rectangle
	Contents  Reference
}

type Rectangle [4]int // in points

type resources struct {
	ProcSet []Name
	Font    map[Name]interface{}
	XObject map[Name]interface{}
}

const (
	pdfProcSet    Name = "PDF"
	textProcSet   Name = "Text"
	imageBProcSet Name = "ImageB"
	imageCProcSet Name = "ImageC"
	imageIProcSet Name = "ImageI"
)

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

type standardFontDict struct {
	Type     Name
	Subtype  Name
	BaseFont Name
}
