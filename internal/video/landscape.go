package video

import (
	"image"
	"image/color"
	"timepilot/internal/game"
)

// Landscape layout: the 224x216 playfield sits in the middle of a 384x216
// (16:9) canvas and the heads-up display that the cabinet shows in strips at
// the top and bottom of the tube is moved into the two side bars.
const (
	LandscapeW = 384
	LandscapeH = 216
	sideW      = 80
	fieldTop   = 32 // first playfield row in the upright picture
)

type rect struct{ x, y, w, h int }

// blit copies src rectangle r to (dx, dy) in dst.
func blit(dst, src *image.RGBA, r rect, dx, dy int) {
	for y := 0; y < r.h; y++ {
		for x := 0; x < r.w; x++ {
			c := src.RGBAAt(r.x+x, r.y+y)
			// the HUD strips use a black background; keep it
			dst.SetRGBA(dx+x, dy+y, c)
		}
	}
}

// playerBlock draws one player's status block: the score label, the stage
// flags beside it and the reserve planes underneath.  x is the left edge of
// the bar, label the crop of the score area in the main HUD layer and ph the
// per-player layer (nil when that player is not in the game).
func playerBlock(dst, hud, ph *image.RGBA, label rect, x int) {
	blit(dst, hud, label, x+6, 8)
	if ph == nil {
		return
	}
	// stage flags: right-aligned in the picture, so find their extent
	minX := ScreenW
	for yy := 16; yy < 32; yy++ {
		for xx := 120; xx < ScreenW; xx++ {
			if c := ph.RGBAAt(xx, yy); c.R|c.G|c.B != 0 && xx < minX {
				minX = xx
			}
		}
	}
	if minX < ScreenW {
		w := ScreenW - minX
		if w > sideW-8 {
			w, minX = sideW-8, ScreenW-(sideW-8)
		}
		blit(dst, ph, rect{minX, 16, w, 16}, x+4+(sideW-8-w), 28)
	}
	for i := 0; i < 2; i++ { // reserve planes, three per line
		blit(dst, ph, rect{8 + i*48, 16, 48, 16}, x+16, 50+i*18)
	}
}

// Landscape builds the landscape picture from the upright one.
// up is the full upright picture and hud the HUD-only layer of the live
// screen (both ScreenW x ScreenH); p1 and p2 are optional per-player HUD
// layers (see game.PlayerHUD) so that both players' stage and reserve planes
// are always visible.  dst must be LandscapeW x LandscapeH.
func Landscape(dst, up, hud, p1, p2 *image.RGBA) {
	black := color.RGBA{0, 0, 0, 255}
	for y := 0; y < LandscapeH; y++ {
		for x := 0; x < LandscapeW; x++ {
			dst.SetRGBA(x, y, black)
		}
	}
	blit(dst, up, rect{0, fieldTop, ScreenW, LandscapeH}, sideW, 0)
	line := color.RGBA{24, 36, 92, 255}
	for y := 0; y < LandscapeH; y++ {
		dst.SetRGBA(sideW-1, y, line)
		dst.SetRGBA(sideW+ScreenW, y, line)
	}
	rx := sideW + ScreenW
	playerBlock(dst, hud, p1, rect{8, 0, 56, 16}, 0)
	playerBlock(dst, hud, p2, rect{160, 0, 64, 16}, rx)
	// hi-score and credits on the right, stage progress on the left
	blit(dst, hud, rect{78, 0, 68, 16}, rx+6, 110)
	blit(dst, hud, rect{8, 248, 64, 8}, 8, 184)
	blit(dst, hud, rect{72, 248, 64, 8}, 8, 196)
	blit(dst, hud, rect{150, 248, 74, 8}, rx+4, 196)
}

// DrawPlayerHUD renders player p's (0 or 1) own HUD layer into dst and
// reports whether that player is in the game.
func (r *Renderer) DrawPlayerHUD(g *game.Game, p int, dst *image.RGBA) bool {
	c, _, _, ok := g.PlayerHUD(p)
	if !ok {
		return false
	}
	r.DrawHUD(c, dst)
	return true
}
