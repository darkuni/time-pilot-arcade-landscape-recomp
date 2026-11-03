// boot runs the reference machine and dumps frames as PNG for analysis.
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"strconv"
	"strings"
	"timepilot/internal/assets"
	"timepilot/tools/oracle/machine"
)

func rot(src *image.RGBA) *image.RGBA {
	b := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dy(), b.Dx()))
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			// 90 degrees clockwise
			dst.Set(b.Dy()-1-y, x, src.At(x, y))
		}
	}
	return dst
}

func main() {
	r, err := assets.Load(os.Args[1])
	if err != nil {
		panic(err)
	}
	m := machine.New(r)
	chars := r.DecodeChars()
	sp := r.DecodeSprites()
	p := r.Palette()
	pal := p
	dumpAt := map[int]bool{}
	type ev struct {
		name     string
		from, to int
	}
	var evs []ev
	last := 0
	for _, a := range os.Args[3:] {
		if strings.Contains(a, ":") {
			parts := strings.Split(a, ":")
			f, _ := strconv.Atoi(parts[1])
			d := 4
			if len(parts) > 2 {
				d, _ = strconv.Atoi(parts[2])
			}
			evs = append(evs, ev{parts[0], f, f + d})
			continue
		}
		n, _ := strconv.Atoi(a)
		dumpAt[n] = true
		if n > last {
			last = n
		}
	}
	_ = color.RGBA{}
	for f := 0; f <= last; f++ {
		m.In0, m.In1, m.In2 = 0xff, 0xff, 0xff
		for _, e := range evs {
			if f >= e.from && f < e.to {
				switch e.name {
				case "coin":
					m.In0 &^= 1
				case "start1":
					m.In0 &^= 8
				case "start2":
					m.In0 &^= 16
				case "left":
					m.In1 &^= 1
				case "right":
					m.In1 &^= 2
				case "up":
					m.In1 &^= 4
				case "down":
					m.In1 &^= 8
				case "fire":
					m.In1 &^= 16
				}
			}
		}
		m.RunFrame()
		if dumpAt[f] {
			img := m.Render(&chars, &sp, &pal)
			fo, _ := os.Create(fmt.Sprintf("%s/f%04d.png", os.Args[2], f))
			png.Encode(fo, rot(img))
			fo.Close()
			fo, _ = os.Create(fmt.Sprintf("%s/f%04d_raw.png", os.Args[2], f))
			png.Encode(fo, img)
			fo.Close()
		}
	}
	fmt.Printf("PC main=%04x sound=%04x frame=%d\n", m.Main.PC, m.Sound.PC, m.Frame)
}
