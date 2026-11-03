// Package video turns the game's video, colour and sprite RAM into pixels.
package video

import (
	"image"
	"image/color"

	"timepilot/internal/assets"
	"timepilot/internal/game"
)

// Raster size of the visible area before the cabinet's 90° rotation.
const (
	RasterW = 256
	RasterH = 224
	// Screen is the size of the upright (rotated) picture.
	ScreenW = RasterH
	ScreenH = RasterW
)

// Renderer draws frames from decoded ROM graphics.
type Renderer struct {
	pal     [32]assets.RGB
	chars   [512]assets.CharTile
	sprites [256]assets.SpriteTile
	sprLUT  [256]byte
	charLUT [256]byte
	raster  [RasterW * RasterH]uint8 // palette index per pixel
	front   [RasterW * RasterH]bool
}

// New decodes the graphics from the ROM set.
func New(r *assets.ROMs) *Renderer {
	return &Renderer{
		pal:     r.Palette(),
		chars:   r.DecodeChars(),
		sprites: r.DecodeSprites(),
		sprLUT:  r.SprLUT,
		charLUT: r.CharLUT,
	}
}

func (r *Renderer) put(x, y int, idx uint8) {
	if x < 0 || x >= RasterW || y < 0 || y >= RasterH {
		return
	}
	r.raster[y*RasterW+x] = idx
}

func (r *Renderer) drawTiles(g *game.Game, frontOnly bool) {
	vram, cram := g.Video(), g.Color()
	for t := 0; t < 0x400; t++ {
		attr := cram[t]
		isFront := attr&0x10 != 0
		if frontOnly && !isFront {
			continue
		}
		code := int(vram[t]) + 8*int(attr&0x20)
		col := int(attr & 0x1f)
		tile := &r.chars[code]
		bx := (t & 31) * 8
		by := (t>>5)*8 - 16
		for ty := 0; ty < 8; ty++ {
			sy := ty
			if attr&0x80 != 0 {
				sy = 7 - ty
			}
			for tx := 0; tx < 8; tx++ {
				sx := tx
				if attr&0x40 != 0 {
					sx = 7 - tx
				}
				p := tile[sy][sx]
				if frontOnly && p == 0 {
					continue
				}
				r.put(bx+tx, by+ty, 0x10+r.charLUT[col*4+int(p)]&0x0f)
			}
		}
	}
}

func (r *Renderer) drawSprites(g *game.Game) {
	s1, s2 := g.Sprites1(), g.Sprites2()
	for offs := 0x3e; offs >= 0x10; offs -= 2 {
		sx := int(s1[offs])
		code := int(s1[offs+1])
		attr := s2[offs]
		sy := 241 - int(s2[offs+1])
		col := int(attr & 0x3f)
		flipx := attr&0x40 == 0
		flipy := attr&0x80 != 0
		tile := &r.sprites[code]
		for ty := 0; ty < 16; ty++ {
			yy := ty
			if flipy {
				yy = 15 - ty
			}
			for tx := 0; tx < 16; tx++ {
				xx := tx
				if flipx {
					xx = 15 - tx
				}
				p := tile[yy][xx]
				if p == 0 {
					continue
				}
				r.put(sx+tx, sy+ty-16, r.sprLUT[col*4+int(p)]&0x0f)
			}
		}
	}
}

// DrawHUD renders only the heads-up-display layer (the front-priority tiles:
// scores, lives, stage icons, credits) on black, with no sprites and none of
// the cloud / bullet background tiles that drift through the HUD strips.
func (r *Renderer) DrawHUD(g *game.Game, dst *image.RGBA) {
	for i := range r.raster {
		r.raster[i] = 0
	}
	vram, cram := g.Video(), g.Color()
	for t := 0; t < 0x400; t++ {
		attr := cram[t]
		if attr&0x10 == 0 {
			continue
		}
		code := int(vram[t]) + 8*int(attr&0x20)
		col := int(attr & 0x1f)
		tile := &r.chars[code]
		bx, by := (t&31)*8, (t>>5)*8-16
		for ty := 0; ty < 8; ty++ {
			sy := ty
			if attr&0x80 != 0 {
				sy = 7 - ty
			}
			for tx := 0; tx < 8; tx++ {
				sx := tx
				if attr&0x40 != 0 {
					sx = 7 - tx
				}
				r.put(bx+tx, by+ty, 0x10+r.charLUT[col*4+int(tile[sy][sx])]&0x0f)
			}
		}
	}
	r.blit(g.FlipScreen, dst)
}

// Draw renders the current game state into dst, which must be ScreenW x
// ScreenH.  The picture is rotated as on the upright cabinet.
func (r *Renderer) Draw(g *game.Game, dst *image.RGBA) {
	for i := range r.raster {
		r.raster[i] = 0
	}
	r.drawTiles(g, false)
	r.drawSprites(g)
	r.drawTiles(g, true)
	r.blit(g.FlipScreen, dst)
}

func (r *Renderer) blit(flip bool, dst *image.RGBA) {
	for y := 0; y < RasterH; y++ {
		for x := 0; x < RasterW; x++ {
			c := r.pal[r.raster[y*RasterW+x]]
			rx, ry := x, y
			if flip {
				rx, ry = RasterW-1-x, RasterH-1-y
			}
			// rotate 90° anticlockwise: raster (x,y) -> screen (y, W-1-x)
			dx, dy := ry, RasterW-1-rx
			dst.SetRGBA(dx, dy, color.RGBA{c.R, c.G, c.B, 255})
		}
	}
}
