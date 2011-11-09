package pdf

import (
	"io"
	"os"
)

//type Writer interface {
//	Line(x1, y1, x2, y2 float64)
//	EndPage() os.Error
//	Close() os.Error
//}

type Writer struct {
	writer io.WriteCloser
}

func (w *Writer) Close() os.Error {
	return w.writer.Close()
}

const header = "%PDF-1.7\r\n"

func New(w io.WriteCloser) (*Writer, os.Error) {
	ww := &Writer{
		writer: w,
	}
	_, err := w.Write([]byte(header))
	if err != nil {
		return nil, err
	}
	return ww, nil
}
