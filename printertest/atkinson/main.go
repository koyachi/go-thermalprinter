package main

import (
	"github.com/koyachi/go-atkinson"
	"github.com/koyachi/go-lena"
	"github.com/koyachi/go-thermalprinter"
	"github.com/nfnt/resize"
	"image"
	"log"
)

func imageToByte(img image.Image) (data []byte, w int, h int, err error) {
	// resize
	width := 384
	resizedImage := resize.Resize(uint(width), 0, img, resize.Lanczos3)
	height := resizedImage.Bounds().Size().Y

	// dither
	ditheredImage, err := atkinson.Dither(resizedImage)
	if err != nil {
		return nil, 0, 0, err
	}

	// convert to byte array
	rowBytes := (width + 7) / 8
	bitmap := make([]byte, rowBytes*height)
	for y := 0; y < height; y++ {
		n := y * rowBytes
		x := 0
		for b := 0; b < rowBytes; b++ {
			sum := 0
			bit := 128
			for bit > 0 {
				if x >= width {
					break
				}
				r, _, _, _ := ditheredImage.At(x, y).RGBA()
				// 完全には２値化できてないので
				if r < 200 {
					sum |= bit
				}
				x += 1
				bit >>= 1
			}
			bitmap[n+b] = byte(sum)
		}
	}
	return bitmap, width, height, nil
}

func main() {
	printer, err := thermalprinter.NewPrinter("/dev/ttyAMA0", 19200, 5)
	if err != nil {
		log.Fatal(err)
	}
	printer.Flush()
	//printer.SetTimes(0.1, 0.1)
	printer.Println("image test")

	img, err := lena.Image()
	if err != nil {
		log.Fatal(err)
	}
	data, w, h, err := imageToByte(img)
	log.Printf("w,h = %d,%d\n", w, h)
	if err != nil {
		log.Fatal(err)
	}
	printer.PrintBitmap(w, h, data, false)
	//printer.PrintBitmap(w, h, data, true)

	printer.Println("image test end.")
}
