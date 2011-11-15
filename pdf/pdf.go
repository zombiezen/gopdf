// Copyright (C) 2011, Ross Light

package pdf

import (
	"image"
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

	stream := newStream(streamLZWDecode)
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
	// TODO: LZW compress (seems to write bad codes)
	st := newStream(streamNoFilter)
	defer st.Close()
	extra := st.Extra()

	extra["Type"] = xobjectType
	extra["Subtype"] = imageSubtype
	extra["Width"] = bd.Dx()
	extra["Height"] = bd.Dy()
	extra["BitsPerComponent"] = 8
	extra["ColorSpace"] = Name("DeviceRGB")
	encodeImageStream(st, img)
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
	_ Name = ""

	Courier            = "Courier"
	CourierBold        = "Courier-Bold"
	CourierOblique     = "Courier-Oblique"
	CourierBoldOblique = "Courier-BoldOblique"

	Helvetica            = "Helvetica"
	HelveticaBold        = "Helvetica-Bold"
	HelveticaOblique     = "Helvetica-Oblique"
	HelveticaBoldOblique = "Helvetica-BoldOblique"

	Symbol = "Symbol"

	Times           = "Times-Roman"
	TimesBold       = "Times-Bold"
	TimesItalic     = "Times-Italic"
	TimesBoldItalic = "Times-BoldItalic"

	ZapfDingbats = "ZapfDingbats"
)

type standardFontDict struct {
	Type     Name
	Subtype  Name
	BaseFont Name
}
