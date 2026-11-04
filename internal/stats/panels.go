package stats

import (
	"fmt"
	"image"
	"image/color"
)

var (
	colLabel = color.RGBA{73, 182, 182, 255}
	colValue = color.RGBA{255, 255, 255, 255}
	colHot   = color.RGBA{255, 255, 73, 255}
	colDim   = color.RGBA{110, 110, 140, 255}
)

func num(n int) string {
	s := fmt.Sprint(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}

func blitIcon(dst, ic *image.RGBA, x, y int) {
	b := ic.Bounds()
	for yy := 0; yy < b.Dy(); yy++ {
		for xx := 0; xx < b.Dx(); xx++ {
			c := ic.RGBAAt(xx, yy)
			if c.A != 0 && image.Pt(x+xx, y+yy).In(dst.Rect) {
				dst.SetRGBA(x+xx, y+yy, c)
			}
		}
	}
}

// Side bars of the 384x216 landscape picture: the playfield occupies
// x 80..303. The landscape layout already puts the score, stage flags and
// reserve planes at the top of each bar and the stage progress at the foot of
// the left one; the read-outs go in the space between.
const (
	leftX  = 6
	rightX = 308
)

// DrawPanels draws the Time Pilot+ read-outs. They follow the player whose
// turn it is; between games they show player 1's totals.
func DrawPanels(dst *image.RGBA, t *Tracker) {
	cur := t.Cur
	if !t.Playing {
		cur = 0
	}
	game, life := &t.GameP[cur], &t.LifeP[cur]

	// ---- left: this player's game --------------------------------------
	y := 88
	if t.Playing && t.Players > 1 {
		DrawText(dst, leftX, y, fmt.Sprintf("PLAYER %d STATS", cur+1), colHot, 1)
		y += 8
	}
	DrawText(dst, leftX, y, "KILLS", colLabel, 1)
	DrawText(dst, leftX, y+7, num(game.Kills), colValue, 2)
	y += 20
	small := func(label, value string, c color.RGBA) {
		DrawText(dst, leftX, y, label, colDim, 1)
		DrawText(dst, leftX+TextWidth(label+" ", 1), y, value, c, 1)
		y += 7
	}
	small("BOSSES", num(game.Bosses), colValue)
	small("RESCUED", num(game.Rescued), colValue)
	small("ACCURACY", fmt.Sprintf("%d%%", game.Accuracy()), colValue)
	y += 2
	DrawText(dst, leftX, y, "ALL TIME", colHot, 1)
	y += 8
	small("GAMES", num(life.Games), colValue)
	small("KILLS", num(life.Kills), colValue)
	small("BOSSES", num(life.Bosses), colValue)
	small("RESCUED", num(life.Rescued), colValue)

	// ---- right: best chain, then the kills per era ----------------------
	ry := 90
	DrawText(dst, rightX, ry, "BEST CHAIN", colLabel, 1)
	DrawText(dst, rightX, ry+7, "x"+num(life.BestChain), colHot, 2)
	ry = 130
	for s := 0; s < Stages; s++ {
		if ic := t.Icons[s]; ic != nil {
			w, h := ic.Bounds().Dx(), ic.Bounds().Dy()
			blitIcon(dst, ic, rightX+(14-w)/2, ry+(12-h)/2)
		} else {
			DrawText(dst, rightX+5, ry+3, "?", colDim, 1)
		}
		DrawText(dst, rightX+18, ry, num(game.Stage[s]), colValue, 1)
		DrawText(dst, rightX+18, ry+6, num(life.Stage[s]), colDim, 1)
		ry += 13
	}
}
