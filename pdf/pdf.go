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
}

// New creates a document value.
func New() *Document {
	doc := new(Document)
	doc.catalog = &catalog{
		Type: catalogType,
	}
	doc.Root = doc.Add(doc.catalog)
	return doc
}

func (doc *Document) NewPage(width, height int) *Canvas {
	page := &pageDict{
		Type:      pageType,
		Resources: map[Name]interface{}{},
		MediaBox:  Rectangle{0, 0, width, height},
		CropBox:   Rectangle{0, 0, width, height},
	}
	pageRef := doc.Add(page)
	doc.pages = append(doc.pages, indirectObject{pageRef, page})

	stream := newStream(streamNoFilter)
	page.Contents = doc.Add(stream)

	return &Canvas{
		doc:      doc,
		page:     page,
		ref:      pageRef,
		contents: stream,
	}
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
	Resources map[Name]interface{}
	MediaBox  Rectangle
	CropBox   Rectangle
	Contents  Reference
}

type Rectangle [4]int // in points
