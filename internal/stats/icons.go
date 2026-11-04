package stats

import (
	"image"

	"timepilot/internal/game"
	"timepilot/internal/video"
)

// EnemyIcons pictures each era's enemy plane from the ROM's own sprite tables.
func EnemyIcons(g *game.Game, r *video.Renderer) [Stages]*image.RGBA {
	code, attr := g.EnemyArt()
	var out [Stages]*image.RGBA
	for i := range out {
		out[i] = r.SpriteImage(code[i], attr[i])
	}
	return out
}
