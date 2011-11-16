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

type standardFontDict struct {
	Type     Name
	Subtype  Name
	BaseFont Name
}
