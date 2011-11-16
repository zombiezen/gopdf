// Copyright (C) 2011, Ross Light

package pdf

import (
	"image"
	"image/ycbcr"
	"io"
	"os"
)

const (
	deviceRGBColorSpace Name = "DeviceRGB"
)

type imageStream struct {
	*stream
	Width            int
	Height           int
	BitsPerComponent int
	ColorSpace       Name
}

type imageStreamInfo struct {
	Type             Name
	Subtype          Name
	Length           int
	Filter           Name `pdf:",omitempty"`
	Width            int
	Height           int
	BitsPerComponent int
	ColorSpace       Name
}

func newImageStream(filter Name, w, h int) *imageStream {
	return &imageStream{
		stream:           newStream(filter),
		Width:            w,
		Height:           h,
		BitsPerComponent: 8,
		ColorSpace:       deviceRGBColorSpace,
	}
}

func (st *imageStream) MarshalPDF() ([]byte, os.Error) {
	return marshalStream(imageStreamInfo{
		Type:             xobjectType,
		Subtype:          imageSubtype,
		Length:           st.Len(),
		Filter:           st.filter,
		Width:            st.Width,
		Height:           st.Height,
		BitsPerComponent: st.BitsPerComponent,
		ColorSpace:       st.ColorSpace,
	}, st.Bytes())
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

func encodeRGBAStream(w io.Writer, img *image.RGBA) os.Error {
	for i := 0; i < len(img.Pix); i += 4 {
		if _, err := w.Write(img.Pix[:3]); err != nil {
			return err
		}
	}
	return nil
}

func encodeYCbCrStream(w io.Writer, img *ycbcr.YCbCr) os.Error {
	var buf [3]byte
	var yy, cb, cr byte
	var i, j int
	dx, dy := img.Rect.Dx(), img.Rect.Dy()
	for y := 0; y < dy; y++ {
		for x := 0; x < dx; x++ {
			i, j = x, y
			switch img.SubsampleRatio {
			case ycbcr.SubsampleRatio420:
				j /= 2
				fallthrough
			case ycbcr.SubsampleRatio422:
				i /= 2
			}
			yy = img.Y[y*img.YStride+x]
			cb = img.Cb[j*img.CStride+i]
			cr = img.Cr[j*img.CStride+i]

			buf[0], buf[1], buf[2] = ycbcr.YCbCrToRGB(yy, cb, cr)
			if _, err := w.Write(buf[:]); err != nil {
				return err
			}
		}
	}
	return nil
}
