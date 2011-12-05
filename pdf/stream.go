package pdf

import (
	"bytes"
	"compress/lzw"
	"compress/zlib"
	"io"
	"os"
)

const (
	streamNoFilter    name = ""
	streamLZWDecode   name = "LZWDecode"
	streamFlateDecode name = "FlateDecode"
)

// stream is a blob of data stored in a PDF file.
type stream struct {
	bytes.Buffer
	writer io.Writer
	filter name
}

type streamInfo struct {
	Length int
	Filter name `pdf:",omitempty"`
}

func newStream(filter name) *stream {
	st := new(stream)
	st.filter = filter
	switch filter {
	case streamLZWDecode:
		st.writer = lzw.NewWriter(&st.Buffer, lzw.MSB, 8)
	case streamFlateDecode:
		st.writer, _ = zlib.NewWriter(&st.Buffer)
	default:
		// TODO: warn about bad filter names?
		st.writer = &st.Buffer
	}
	return st
}

func (st *stream) ReadFrom(r io.Reader) (n int64, err os.Error) {
	return io.Copy(st.writer, r)
}

func (st *stream) Write(p []byte) (n int, err os.Error) {
	return st.writer.Write(p)
}

func (st *stream) WriteByte(c byte) os.Error {
	_, err := st.writer.Write([]byte{c})
	return err
}

func (st *stream) WriteString(s string) (n int, err os.Error) {
	return io.WriteString(st.writer, s)
}

func (st *stream) Close() os.Error {
	if wc, ok := st.writer.(io.WriteCloser); ok {
		return wc.Close()
	}
	return nil
}

func (st *stream) marshalPDF(dst []byte) ([]byte, os.Error) {
	return marshalStream(dst, streamInfo{
		Length: st.Len(),
		Filter: st.filter,
	}, st.Bytes())
}

const (
	streamBegin = " stream\r\n"
	streamEnd   = "\r\nendstream"
)

// marshalStream encodes a generic stream.  The resulting data encodes the
// given object and a sequence of bytes.  This function does not enforce any
// rules about the object being encoded.
func marshalStream(dst []byte, obj interface{}, data []byte) ([]byte, os.Error) {
	var err os.Error
	if dst, err = marshal(dst, obj); err != nil {
		return nil, err
	}
	dst = append(dst, streamBegin...)
	dst = append(dst, data...)
	dst = append(dst, streamEnd...)
	return dst, nil
}
