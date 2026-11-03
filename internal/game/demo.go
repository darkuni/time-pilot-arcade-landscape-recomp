package game

// run210e prepares demonstration playback: it selects the recorded input
// stream for the current stage (0x210e).
func (g *Game) run210e() {
	var hl uint16
	switch a := g.rb(0xad14); {
	case a == 0 || a == 3:
		hl = 0x218c
	case a == 1:
		hl = 0x2251
	default:
		hl = 0x22fa
	}
	g.wb(0xadf2, g.rb(hl)+1)
	g.ww(0xadf3, hl)
}

// run460e is the tamper gate at 0x460e (compares two copied screen cells);
// it has no effect on a good ROM.
func (g *Game) run460e() {}
