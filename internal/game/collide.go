package game

// ---- collisions ----------------------------------------------------------------
//
// All tests are "box" tests on unsigned 8-bit differences, exactly like the
// original: (a - b + l) < h.

// inBox reports whether a-b+l < h using 8-bit wrap-around.
func inBox(a, b, l, h uint8) bool { return a-b+l < h }

// bonusSound is routine 0x51de: score sound/points for a kill.  Successive
// kills inside half a second climb the points ladder.
func (g *Game) bonusSound() {
	if g.rb(0xa99d) != 0 {
		n := g.rb(0xa99e) + 1
		g.wb(0xa99e, n)
		g.enq(4, n&7+1)
		g.emit(EvKill, int(n)+1, 0)
	} else {
		g.enq(4, 1)
		g.emit(EvKill, 1, 0)
	}
	g.wb(0xa99d, 0x1e)
}

// hitTicker is routine 0x5205: the half-second window for the kill ladder.
func (g *Game) hitTicker() {
	if v := g.rb(0xa99d); v != 0 {
		g.wb(0xa99d, v-1)
		return
	}
	g.wb(0xa99e, 0)
}

// cursor is a pair (record pointer, sprite-pair pointer) walked in step.
type cursor struct{ de, iy uint16 }

func (c *cursor) step() { c.de = c.de&0xff00 | (c.de+0x10)&0xff; c.iy += 2 }

// playerVs implements 0x5121 (kind 0), 0x5152 (kind 1), 0x5185 (kind 2) and
// 0x51b3 (kind 3): test the player's plane against b records starting at c.
func (g *Game) playerVs(c *cursor, b int, l, h uint8, kind int) {
	if g.rb(0xa800) != 0xff {
		return
	}
	px, py := g.rb(0xaa10), g.rb(0xaa41)
	for ; b > 0; b-- {
		if g.rb(c.de) == 0xff {
			ex, ey := g.rb(c.iy), g.rb(c.iy+0x31)
			hit := false
			switch kind {
			case 0, 2, 3:
				hit = inBox(px, ex, l, h) && inBox(py, ey, l, h)
			case 1:
				hit = inBox(px, ex, l, h) && py-ey+8 < 0x11
			}
			if hit {
				if kind != 3 {
					g.wb(0xa800, 0xf0)
				}
				g.wb(c.de, 0xf0)
				if kind == 0 || kind == 1 {
					g.bonusSound()
				}
			}
		}
		c.step()
	}
}

// playerVsBoss is 0x50ee / 0x50b1: the plane against the stage boss a8a0.
func (g *Game) playerVsBoss() {
	if g.rb(0xa800) != 0xff || g.rb(0xa8a0) != 0xff {
		return
	}
	px, py := g.rb(0xaa10), g.rb(0xaa41)
	bx, by := g.rb(0xaa24), g.rb(0xaa55)
	st := g.rb(vStage)
	var hit bool
	if st == 0 || st == 4 {
		hit = bx-px+8 < 0x11 && by-py+0x19 < 0x23
	} else {
		hit = bx-px+6 < 13 && by-py+0x19 < 0x23
	}
	if hit {
		g.wb(0xa800, 0xf0)
		g.wb(0xa8a0, 0xf0)
		g.wb(0xa8a4, 0)
		g.bonusSound()
	}
}

// playerVsBoss1 is 0x507e: the plane against the stage-2 boss a8c0.
func (g *Game) playerVsBoss1() {
	if g.rb(0xa800) != 0xff || g.rb(0xa8c0) != 0xff {
		return
	}
	px, py := g.rb(0xaa10), g.rb(0xaa41)
	if g.rb(0xaa28)-px+6 < 13 && g.rb(0xaa59)-py+0x18 < 0x21 {
		g.wb(0xa800, 0xf0)
		g.wb(0xa8c0, 0xf0)
		g.wb(0xa8dc, 0)
		g.bonusSound()
	}
}

// bulletsVsRecords is 0x5211: every bullet in flight against b records.
func (g *Game) bulletsVsRecords(c cursor, b int, l, h uint8) {
	g.ww(0xa993, c.de)
	g.ww(0xa991, c.iy)
	for i := uint16(0); i < 6; i++ {
		bp := 0xaa80 + i*0x10
		if g.rb(bp) != 0xff {
			continue
		}
		cc := c
		for n := b; n > 0; n-- {
			if g.rb(cc.de) == 0xff {
				ex, ey := g.rb(cc.iy), g.rb(cc.iy+0x31)
				if ex+8 >= 0x19 && ey+0x10 >= 0x11 &&
					inBox(g.rb(bp+6), ex, l, h) && inBox(g.rb(bp+4), ey, l, h) {
					g.wb(bp, 0xf0)
					g.wb(cc.de, 0xf0)
					g.emit(EvHit, 0, 0)
					g.bonusSound()
				}
			}
			cc.step()
		}
	}
}

// bulletsVsBoss is 0x4fee / 0x4f91: bullets against one big record (a8a0 or
// a8c0, whose position lives at sprite aa24 / aa28).
func (g *Game) bulletsVsBoss(rec, spr uint16, l, h uint8) {
	if g.rb(rec) != 0xff {
		return
	}
	for i := uint16(0); i < 6; i++ {
		bp := 0xaa80 + i*0x10
		if g.rb(bp) != 0xff {
			continue
		}
		if inBox(g.rb(spr), g.rb(bp+6), l, h) && inBox(g.rb(spr+0x31), g.rb(bp+4), 0x17, 0x1f) {
			g.wb(rec, 0xf0)
			g.wb(bp, 0xf0)
			g.emit(EvHit, 0, 0)
			g.bonusSound()
		}
	}
}

// bulletsVsBossStage chooses the box for the a8a0 test (0x4fe0).
func (g *Game) bulletsVsStageBoss() {
	l, h := uint8(6), uint8(13)
	if st := g.rb(vStage); st == 0 || st == 4 {
		l, h = 8, 0x11
	}
	g.bulletsVsBoss(0xa8a0, 0xaa24, l, h)
}

// collisions is routine 0x4e4f.
func (g *Game) collisions() {
	st := g.rb(vStage)
	odd := g.rb(vFrameCtr)&1 != 0
	switch {
	case st == 4:
		if odd {
			g.collideOddStage4()
			return
		}
		g.collideEven()
	case st == 1:
		if odd {
			g.collideOdd()
			return
		}
		g.collideStage1()
	default:
		if odd {
			g.collideOdd()
			return
		}
		g.collideEven()
	}
}

// collideEven is 0x4e63: the plane against everything else.
func (g *Game) collideEven() {
	// bullets against the three objects at a8c0 (0x4f5d)
	g.bulletsVsRecords(cursor{0xa8c0, 0xaa28}, 3, 7, 15)
	c := cursor{0xa810, 0xaa12}
	g.playerVs(&c, 4, 5, 11, 2)
	if g.rb(0xad0d) == 0 {
		g.playerVs(&c, 7, 7, 15, 1)
		g.playerVs(&c, 3, 6, 13, 0)
		g.playerVs(&c, 1, 8, 0x11, 3)
		return
	}
	g.playerVs(&c, 5, 7, 15, 1)
	g.playerVsBoss()
	c = cursor{0xa8c0, 0xaa28}
	g.playerVs(&c, 3, 6, 13, 0)
	g.playerVs(&c, 1, 8, 0x11, 3)
}

// collideStage1 is 0x4ebc (even frames of the second stage).
func (g *Game) collideStage1() {
	g.bulletsVsBoss(0xa8c0, 0xaa28, 6, 13) // 0x4f7e, only while a8c0 is alive
	c := cursor{0xa810, 0xaa12}
	g.playerVs(&c, 4, 5, 11, 2)
	if g.rb(0xad0d) == 0 {
		g.playerVs(&c, 7, 7, 15, 1)
		g.playerVsBoss1()
	} else {
		g.playerVs(&c, 5, 7, 15, 1)
		g.playerVsBoss()
		g.playerVsBoss1()
	}
	c = cursor{0xa8e0, 0xaa2c}
	g.playerVs(&c, 1, 5, 11, 2)
	g.playerVs(&c, 1, 8, 0x11, 3)
}

// collideOdd is 0x4f35: bullets against enemy planes (and the boss).
func (g *Game) collideOdd() {
	if g.rb(0xad0d) != 0 {
		g.bulletsVsRecords(cursor{0xa850, 0xaa1a}, 5, 7, 15)
		g.bulletsVsStageBoss()
		return
	}
	g.bulletsVsRecords(cursor{0xa850, 0xaa1a}, 7, 7, 15)
}

// collideOddStage4 is 0x5032.
func (g *Game) collideOddStage4() {
	if g.rb(0xad0d) == 0 {
		g.bulletsVsRecords(cursor{0xa810, 0xaa12}, 11, 7, 15)
		return
	}
	g.bulletsVsRecords(cursor{0xa810, 0xaa12}, 9, 7, 15)
	g.bulletsVsStageBoss()
}
