package pdf

import (
	"compress/lzw"
	"io/ioutil"
	"testing"
)

const streamTestString = "Hello, World!\n"

func TestUnfilteredStream(t *testing.T) {
	st := newStream(streamNoFilter)
	st.WriteString(streamTestString)
	st.Close()

	if st.String() != streamTestString {
		t.Errorf("Stream is %q, wanted %q", st.String(), streamTestString)
	}
}

func TestLZWStream(t *testing.T) {
	st := newStream(streamLZWDecode)
	st.WriteString(streamTestString)
	st.Close()

	output, _ := ioutil.ReadAll(lzw.NewReader(st, lzw.MSB, 8))
	if string(output) != streamTestString {
		t.Errorf("Stream is %q, wanted %q", output, streamTestString)
	}
}
