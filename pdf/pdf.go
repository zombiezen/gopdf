// Copyright (C) 2011, Ross Light

package pdf

import (
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
			ProcSet: []Name{pdfProcSet, textProcSet},
			Font:    make(map[Name]interface{}),
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
		Subtype:  "Type1",
		BaseFont: name,
	})
	doc.fonts[name] = ref
	return ref
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
	pageNodeType      = "Pages"
	pageType          = "Page"
	fontType          = "Font"
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
}

const (
	pdfProcSet    Name = "PDF"
	textProcSet        = "Text"
	imageBProcSet      = "ImageB"
	imageCProcSet      = "ImageC"
	imageIProcSet      = "ImageI"
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
