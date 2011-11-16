package pdf

import (
	"image"
	"image/bmp"
	"image/jpeg"
	"image/ycbcr"
	"io/ioutil"
	"os"
	"testing"
)

const suzanneBytes = 512 * 512 * 3

func loadSuzanneRGBA() (*image.RGBA, os.Error) {
	f, err := os.Open("testdata/suzanne.bmp")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, err := bmp.Decode(f)
	if err != nil {
		return nil, err
	}
	return img.(*image.RGBA), nil
}

func loadSuzanneYCbCr() (*ycbcr.YCbCr, os.Error) {
	f, err := os.Open("testdata/suzanne.jpg")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, err := jpeg.Decode(f)
	if err != nil {
		return nil, err
	}
	return img.(*ycbcr.YCbCr), nil
}

func BenchmarkEncodeRGBAGeneric(b *testing.B) {
	b.StopTimer()
	img, _ := loadSuzanneRGBA()
	b.SetBytes(suzanneBytes)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		encodeImageStream(ioutil.Discard, img)
	}
}

func BenchmarkEncodeRGBA(b *testing.B) {
	b.StopTimer()
	img, _ := loadSuzanneRGBA()
	b.SetBytes(suzanneBytes)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		encodeRGBAStream(ioutil.Discard, img)
	}
}

func BenchmarkEncodeYCbCrGeneric(b *testing.B) {
	b.StopTimer()
	img, _ := loadSuzanneYCbCr()
	b.SetBytes(suzanneBytes)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		encodeImageStream(ioutil.Discard, img)
	}
}

func BenchmarkEncodeYCbCr(b *testing.B) {
	b.StopTimer()
	img, _ := loadSuzanneYCbCr()
	b.SetBytes(suzanneBytes)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		encodeYCbCrStream(ioutil.Discard, img)
	}
}
