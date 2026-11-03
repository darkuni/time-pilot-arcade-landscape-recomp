package machine

import (
	"image"
	"image/color"
	"timepilot/internal/assets"
)

// Render draws the current screen (256x224 raw, unrotated) from video RAM.
// This is a reference renderer for the oracle only.
func (m *Machine) Render(chars *[512]assets.CharTile, spr *[256]assets.SpriteTile, pal *[32]assets.RGB) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 256, 224))
	mb := m.mb
	for i := 3; i < len(img.Pix); i += 4 {
		img.Pix[i] = 255
	}
	back := func(x, y int, c assets.RGB) { img.Set(x, y, color.RGBA{c.R, c.G, c.B, 255}) }
	// background fill: pen 0 of char lookup
	drawTiles := func(cat int) {
		for ty := 0; ty < 32; ty++ {
			for tx := 0; tx < 32; tx++ {
				idx := ty*32 + tx
				attr := mb.colorRAM[idx]
				if int(attr>>4)&1 != cat {
					continue
				}
				code := int(mb.videoRAM[idx]) + 8*int(attr&0x20)
				col := int(attr & 0x1f)
				fx, fy := attr&0x40 != 0, attr&0x80 != 0
				for y := 0; y < 8; y++ {
					for x := 0; x < 8; x++ {
						sx, sy := x, y
						if fx {
							sx = 7 - x
						}
						if fy {
							sy = 7 - y
						}
						p := chars[code][sy][sx]
						pen := int(m.ROM.CharLUT[col<<2|int(p)]&0xf) + 0x10
						px, py := tx*8+x, ty*8+y-16
						if py < 0 || py >= 224 {
							continue
						}
						if cat == 1 && p == 0 {
							continue
						}
						back(px, py, pal[pen])
					}
				}
			}
		}
	}
	drawTiles(0)
	for offs := 0x3e; offs >= 0x10; offs -= 2 {
		sx := int(mb.spr[offs])
		sy := 241 - int(mb.spr2[offs+1])
		code := int(mb.spr[offs+1])
		col := int(mb.spr2[offs] & 0x3f)
		fx := ^mb.spr2[offs]&0x40 != 0
		fy := mb.spr2[offs]&0x80 != 0
		for y := 0; y < 16; y++ {
			for x := 0; x < 16; x++ {
				ssx, ssy := x, y
				if fx {
					ssx = 15 - x
				}
				if fy {
					ssy = 15 - y
				}
				p := spr[code][ssy][ssx]
				if p == 0 {
					continue
				}
				pen := int(m.ROM.SprLUT[col<<2|int(p)] & 0xf)
				px, py := sx+x, sy+y-16
				if px < 0 || px >= 256 || py < 0 || py >= 224 {
					continue
				}
				back(px, py, pal[pen])
			}
		}
	}
	drawTiles(1)
	return img
}
