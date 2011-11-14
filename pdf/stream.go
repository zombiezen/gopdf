package pdf

import (
	"bytes"
	"compress/lzw"
	"io"
	"os"
)

const (
	streamNoFilter  Name = ""
	streamLZWDecode      = "LZWDecode"
)

// Stream is a blob of data.
type stream struct {
	bytes.Buffer
	writer io.Writer
	filter Name
}

func newStream(filter Name) *stream {
	st := new(stream)
	st.filter = filter
	switch filter {
	case streamLZWDecode:
		st.writer = lzw.NewWriter(&st.Buffer, lzw.MSB, 8)
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

const (
	streamBegin = "stream\r\n"
	streamEnd   = "\r\nendstream"
)

func (st *stream) MarshalPDF() ([]byte, os.Error) {
	d := map[Name]interface{}{
		"Length": st.Len(),
	}
	if st.filter != streamNoFilter {
		d["Filter"] = st.filter
	}
	mdict, err := Marshal(d)
	if err != nil {
		return nil, err
	}

	var b bytes.Buffer
	b.Write(mdict)
	b.WriteString(streamBegin)
	b.Write(st.Bytes())
	b.WriteString(streamEnd)
	return b.Bytes(), nil
}
