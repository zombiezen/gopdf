// Copyright (C) 2011, Ross Light

package pdf

import (
	"fmt"
	"io"
	"os"
)

// Encoder writes the PDF file format structure.
type Encoder struct {
	objects []interface{}
	Root    Reference
}

type trailer struct {
	Size int
	Root Reference
}

// Add appends an object to the file.  The object is marshalled only when an
// encoding is requested.
func (enc *Encoder) Add(v interface{}) Reference {
	enc.objects = append(enc.objects, v)
	return Reference{uint(len(enc.objects)), 0}
}

const (
	header  = "%PDF-1.7" + newline + "%\x93\x8c\x8b\x9e" + newline
	newline = "\r\n"
)

// Cross reference strings
const (
	crossReferenceSectionHeader    = "xref" + newline
	crossReferenceSubsectionFormat = "%d %d" + newline
	crossReferenceFormat           = "%010d %05d n" + newline
	crossReferenceFreeFormat       = "%010d %05d f" + newline
)

const trailerHeader = "trailer" + newline

const startxrefFormat = "startxref" + newline + "%d" + newline

const eofString = "%%EOF" + newline

// Encode writes an entire PDF document by marshalling the added objects.
func (enc *Encoder) Encode(wr io.Writer) os.Error {
	w := &offsetWriter{Writer: wr}
	if err := enc.writeHeader(w); err != nil {
		return err
	}
	objectOffsets, err := enc.writeBody(w)
	if err != nil {
		return err
	}
	tableOffset := w.offset
	if err := enc.writeXrefTable(w, objectOffsets); err != nil {
		return err
	}
	if err := enc.writeTrailer(w); err != nil {
		return err
	}
	if err := enc.writeStartxref(w, tableOffset); err != nil {
		return err
	}
	if err := enc.writeEOF(w); err != nil {
		return err
	}
	return nil
}

func (enc *Encoder) writeHeader(w *offsetWriter) os.Error {
	_, err := io.WriteString(w, header)
	return err
}

func (enc *Encoder) writeBody(w *offsetWriter) ([]int64, os.Error) {
	objectOffsets := make([]int64, len(enc.objects))
	for i, obj := range enc.objects {
		objectOffsets[i] = w.offset
		data, err := Marshal(indirectObject{Reference{uint(i + 1), 0}, obj})
		if err != nil {
			return nil, err
		}
		if _, err = w.Write(data); err != nil {
			return nil, err
		}
		if _, err = io.WriteString(w, newline); err != nil {
			return nil, err
		}
	}
	return objectOffsets, nil
}

func (enc *Encoder) writeXrefTable(w *offsetWriter, objectOffsets []int64) os.Error {
	if _, err := io.WriteString(w, crossReferenceSectionHeader); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, crossReferenceSubsectionFormat, 0, len(enc.objects)+1); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, crossReferenceFreeFormat, 0, 65535); err != nil {
		return err
	}
	for _, offset := range objectOffsets {
		if _, err := fmt.Fprintf(w, crossReferenceFormat, offset, 0); err != nil {
			return err
		}
	}
	return nil
}

func (enc *Encoder) writeTrailer(w *offsetWriter) os.Error {
	if _, err := io.WriteString(w, trailerHeader); err != nil {
		return err
	}
	trailerDict := trailer{
		Size: len(enc.objects) + 1,
		Root: enc.Root,
	}
	trailerData, err := Marshal(trailerDict)
	if err != nil {
		return err
	}
	if _, err := w.Write(trailerData); err != nil {
		return err
	}
	if _, err := io.WriteString(w, newline); err != nil {
		return err
	}
	return nil
}

func (enc *Encoder) writeStartxref(w *offsetWriter, tableOffset int64) os.Error {
	_, err := fmt.Fprintf(w, startxrefFormat, tableOffset)
	return err
}

func (enc *Encoder) writeEOF(w *offsetWriter) os.Error {
	_, err := io.WriteString(w, eofString)
	return err
}

// offsetWriter tracks how many bytes have been written to it.
type offsetWriter struct {
	io.Writer
	offset int64
}

func (w *offsetWriter) Write(p []byte) (n int, err os.Error) {
	n, err = w.Writer.Write(p)
	w.offset += int64(n)
	return
}
