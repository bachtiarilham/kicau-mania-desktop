//go:build ignore

// Generate a code-drawn camera icon for the Windows resource bundle.
package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"os"
)

func main() {
	const size = 128
	im := image.NewNRGBA(image.Rect(0, 0, size, size))
	dark := color.NRGBA{19, 28, 29, 255}
	lime := color.NRGBA{212, 237, 149, 255}
	rect := func(x0, y0, x1, y1 int, c color.NRGBA) {
		for y := y0; y < y1; y++ {
			for x := x0; x < x1; x++ {
				im.SetNRGBA(x, y, c)
			}
		}
	}
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			cx, cy := x, y
			if cx < 20 {
				cx = 20
			}
			if cx > 107 {
				cx = 107
			}
			if cy < 20 {
				cy = 20
			}
			if cy > 107 {
				cy = 107
			}
			if (x-cx)*(x-cx)+(y-cy)*(y-cy) <= 400 {
				im.SetNRGBA(x, y, dark)
			}
		}
	}
	rect(25, 38, 88, 78, lime)
	for y := 45; y < 72; y++ {
		for x := 88; x < 106; x++ {
			if x-88 < y-43 && x-88 < 74-y {
				im.SetNRGBA(x, y, lime)
			}
		}
	}
	rect(49, 78, 60, 94, lime)
	rect(37, 92, 75, 99, lime)
	rect(32, 45, 79, 70, dark)
	var data bytes.Buffer
	_ = png.Encode(&data, im)
	out, _ := os.Create("build/windows/icon.ico")
	defer out.Close()
	_ = binary.Write(out, binary.LittleEndian, []uint16{0, 1, 1})
	out.Write([]byte{128, 128, 0, 0})
	_ = binary.Write(out, binary.LittleEndian, []uint16{1, 32})
	_ = binary.Write(out, binary.LittleEndian, uint32(data.Len()))
	_ = binary.Write(out, binary.LittleEndian, uint32(22))
	out.Write(data.Bytes())
}
