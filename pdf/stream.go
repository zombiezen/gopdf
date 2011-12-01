package pdf

import (
	"bytes"
	"compress/lzw"
	"compress/zlib"
	"io"
)

const (
	streamNoFilter    Name = ""
	streamLZWDecode   Name = "LZWDecode"
	streamFlateDecode Name = "FlateDecode"
)

// stream is a blob of data stored in a PDF file.
type stream struct {
	bytes.Buffer
	writer io.Writer
	filter Name
}

type streamInfo struct {
	Length int
	Filter Name `pdf:",omitempty"`
}

func newStream(filter Name) *stream {
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

func (st *stream) ReadFrom(r io.Reader) (n int64, err error) {
	return io.Copy(st.writer, r)
}

func (st *stream) Write(p []byte) (n int, err error) {
	return st.writer.Write(p)
}

func (st *stream) WriteByte(c byte) error {
	_, err := st.writer.Write([]byte{c})
	return err
}

func (st *stream) WriteString(s string) (n int, err error) {
	return io.WriteString(st.writer, s)
}

func (st *stream) Close() error {
	if wc, ok := st.writer.(io.WriteCloser); ok {
		return wc.Close()
	}
	return nil
}

func (st *stream) MarshalPDF() ([]byte, error) {
	return marshalStream(streamInfo{
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
func marshalStream(obj interface{}, data []byte) ([]byte, error) {
	mobj, err := Marshal(obj)
	if err != nil {
		return nil, err
	}

	b := make([]byte, 0, len(mobj)+len(streamBegin)+len(data)+len(streamEnd))
	b = append(b, mobj...)
	b = append(b, []byte(streamBegin)...)
	b = append(b, data...)
	b = append(b, []byte(streamEnd)...)
	return b, nil
}
