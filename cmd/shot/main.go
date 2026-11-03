// shot runs the game headless and writes PNG snapshots (development tool).
package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"strconv"
	"strings"

	"timepilot/internal/assets"
	"timepilot/internal/game"
	"timepilot/internal/video"
)

func main() {
	r, err := assets.Load(os.Args[1])
	if err != nil {
		panic(err)
	}
	frames := map[int]bool{}
	for _, s := range strings.Split(os.Args[2], ",") {
		n, _ := strconv.Atoi(s)
		frames[n] = true
	}
	g := game.New(r.Main[:])
	g.Boot()
	rd := video.New(r)
	up := image.NewRGBA(image.Rect(0, 0, video.ScreenW, video.ScreenH))
	img := up
	max := 0
	for n := range frames {
		if n > max {
			max = n
		}
	}
	for f := 0; f <= max; f++ {
		g.In0 = 0xff
		if f >= 600 && f < 604 { // coin
			g.In0 &^= 1
		}
		if f >= 700 && f < 704 { // 1P start
			g.In0 &^= 8
			if os.Getenv("P2") != "" {
				g.In0 &^= 0x10
				g.In0 |= 8
			}
		}
		if os.Getenv("P2") != "" && f >= 590 && f < 594 { // second coin
			g.In0 &^= 1
		}
		g.In1 = 0xff
		if f > 1100 {
			g.In1 &^= 0x10
		}
		g.Frame()
		if frames[f] {
			rd.Draw(g, up)
			if os.Getenv("LANDSCAPE") != "" {
				l := image.NewRGBA(image.Rect(0, 0, video.LandscapeW, video.LandscapeH))
				hud := image.NewRGBA(image.Rect(0, 0, video.ScreenW, video.ScreenH))
				rd.DrawHUD(g, hud)
				var ph [2]*image.RGBA
				for i := range ph {
					h := image.NewRGBA(image.Rect(0, 0, video.ScreenW, video.ScreenH))
					if rd.DrawPlayerHUD(g, i, h) {
						ph[i] = h
					}
				}
				video.Landscape(l, up, hud, ph[0], ph[1])
				img = l
			} else {
				img = up
			}
			if os.Getenv("HUDDBG") != "" {
				for p := 0; p < 2; p++ {
					if c, _, _, ok := g.PlayerHUD(p); ok {
						h := image.NewRGBA(image.Rect(0, 0, video.ScreenW, video.ScreenH))
						rd.DrawHUD(c, h)
						fh, _ := os.Create(fmt.Sprintf("/tmp/hud_%d_%d.png", f, p))
						png.Encode(fh, h)
						fh.Close()
					}
				}
			}
			fn := fmt.Sprintf("/tmp/shot_%d.png", f)
			fh, _ := os.Create(fn)
			png.Encode(fh, img)
			fh.Close()
			fmt.Println("wrote", fn)
		}
	}
}
