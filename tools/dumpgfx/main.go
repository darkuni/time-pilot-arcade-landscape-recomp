// dumpgfx writes PNG sheets of decoded graphics for visual verification.
package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"timepilot/internal/assets"
)

func save(name string, img image.Image) {
	f, _ := os.Create(name)
	defer f.Close()
	png.Encode(f, img)
}

func main() {
	r, err := assets.Load(os.Args[1])
	if err != nil {
		panic(err)
	}
	out := os.Args[2]
	chars := r.DecodeChars()
	g4 := []uint8{0, 85, 170, 255}
	img := image.NewGray(image.Rect(0, 0, 32*8, 16*8))
	for c := 0; c < 512; c++ {
		for y := 0; y < 8; y++ {
			for x := 0; x < 8; x++ {
				img.SetGray((c%32)*8+x, (c/32)*8+y, color.Gray{g4[chars[c][y][x]]})
			}
		}
	}
	save(out+"/chars.png", scaled(img, 3))
	sp := r.DecodeSprites()
	simg := image.NewGray(image.Rect(0, 0, 16*16, 16*16))
	for s := 0; s < 256; s++ {
		for y := 0; y < 16; y++ {
			for x := 0; x < 16; x++ {
				simg.SetGray((s%16)*16+x, (s/16)*16+y, color.Gray{g4[sp[s][y][x]]})
			}
		}
	}
	save(out+"/sprites.png", scaled(simg, 3))
}

func scaled(src image.Image, k int) image.Image {
	b := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx()*k, b.Dy()*k))
	for y := 0; y < b.Dy()*k; y++ {
		for x := 0; x < b.Dx()*k; x++ {
			dst.Set(x, y, src.At(b.Min.X+x/k, b.Min.Y+y/k))
		}
	}
	return dst
}
