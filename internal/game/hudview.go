package game

// PlayerHUD returns a scratch copy of the game whose video RAM holds only
// player p's (0 or 1) reserve planes, stage icons and stage number, drawn by
// the game's own routines.  It lets a front end show both players' status at
// once, although the original screen only ever shows the active player's.
// ok is false when that player has no game in progress.
func (g *Game) PlayerHUD(p int) (c *Game, lives, stage uint8, ok bool) {
	if g.rb(vGameActive) == 0 {
		return nil, 0, 0, false
	}
	blk := uint16(0xad10)
	pal := uint16(0xad1b)
	if p == 1 {
		blk, pal = 0xad20, 0xad2b
	}
	if (g.rb(vPlayer2) != 0) == (p == 1) {
		blk = 0xad00 // the active player's live record
		pal = 0xad0b
	}
	lives, stage = g.rb(blk), g.rb(blk+1)
	if lives == 0 {
		return nil, 0, 0, false
	}
	cp := *g
	c = &cp
	for i := range c.vram {
		c.vram[i] = 0xf1
		c.cram[i] = 0x10
	}
	c.wb(0xad01, stage)
	c.wb(0xad0c, g.rb(pal+1))
	if lives > 1 {
		c.drawLives(lives - 1)
	}
	c.drawIcons(stage)
	c.drawStageNumber()
	return c, lives, stage, true
}

// Peek returns n bytes of the address space starting at a (diagnostics).
func (g *Game) Peek(a uint16, n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = g.rb(a + uint16(i))
	}
	return b
}
