package game

// ---- the stage-2 boss (records a8c0 / a8d0 / a8e0) -----------------------------
//
// On the second stage a big craft crosses the screen; the ROM keeps it in
// the record at a8c0 (a second half at a8d0) and fires from it.

// boss1 is routine 0x3b5f.
func (g *Game) boss1() {
	if g.rb(vStage) != 1 {
		return
	}
	g.at(0xa8c0, 0xaa28)
	switch s := g.X(0); {
	case s == 0:
		g.bossWait()
		return
	case s == 0xff:
	default:
		c := s - 1
		if v := g.rb(0xa8dc); v != 0 {
			g.wb(0xa8dc, v-1)
			g.setX(0, 0xff)
			g.soundBurst()
			break
		}
		if c >= 0x61 {
			g.setX(0, 0x61)
			g.soundBurst()
			g.setY(0x30, 0x3d)
			g.setY(0x32, 0x3d)
		}
		g.bossDie()
		return
	}
	// 0x3b77: flying
	g.applyVelScroll()
	g.setY(0x33, g.Y(0x31)+0x10)
	g.setY(2, g.Y(0))
	if g.bossGone() {
		g.bossReset()
		return
	}
	g.bossSprite()
	g.bossFire()
}

// bossDie is 0x3bbd: the explosion sequence.
func (g *Game) bossDie() {
	if g.decX(0) == 0 {
		g.bossReset()
		return
	}
	g.applyScroll()
	g.setY(0x33, g.Y(0x31)+0x10)
	g.setY(2, g.Y(0))
	a := g.X(0) - 0x40
	if g.X(0) == 0x40 {
		g.enq(4, 0x0b)
		g.setY(3, 0xfa)
		g.setY(1, 0xfb)
		g.setY(0x30, 0x6c)
		g.setY(0x32, 0x6c)
		g.decX(0)
		return
	}
	if g.X(0) < 0x40 {
		return
	}
	if a&7 != 0 {
		return
	}
	v := g.rb(0x3c09 + uint16(a>>3-1))
	g.setY(3, v)
	g.setY(1, v+1)
}

// bossWait is 0x3c25: count down until the boss is sent in.
func (g *Game) bossWait() {
	if g.rb(vFrameCtr)&1 != 0 {
		return
	}
	if g.decX(14) != 0 {
		return
	}
	if g.rb(0xad0d) != 0 {
		return
	}
	b := g.rb(0xa802)
	a := b
	if (b+8)&0x7f < 0x10 {
		a = 0x10
		if g.rb(vFrameCtr)&8 == 0 {
			a = 0xf0
		}
		a += b
	}
	hl := 0x3c84 + uint16(a>>2&0x3e)
	g.setY(0x31, g.rb(hl))
	g.setY(0, g.rb(hl+1))
	g.setX(2, (b+0xc0)&0x80)
	de, bc := g.vec(0x59d7, g.X(2))
	g.setXw(10, de)
	g.setXw(12, bc)
	g.wb(0xa8dc, 3)
	g.setX(0, 0xff)
}

// bossGone is 0x3cc4: true once the boss has left the screen.
func (g *Game) bossGone() bool {
	if (g.X(2)+0x40)&0x80 != 0 {
		if g.Y(0x31)+0x10 < 3 {
			return true
		}
	} else {
		if g.Y(0x31)+0x13 < 3 {
			return true
		}
	}
	return g.Y(0)+2 < 4
}

// bossSprite is 0x3ce9.
func (g *Game) bossSprite() {
	b := g.rb(vFrameCtr) & 2
	c := (3-g.rb(0xa8dc))<<2 + 0xa0 + b
	if g.X(2)+0x40 >= 0x80 {
		g.setY(1, c)
		g.setY(3, c+1)
		g.setY(0x30, 0xed)
		g.setY(0x32, 0xed)
	} else {
		g.setY(3, c)
		g.setY(1, c+1)
		g.setY(0x30, 0x6d)
		g.setY(0x32, 0x6d)
	}
}

// bossFire is 0x3d25.
func (g *Game) bossFire() {
	if g.X(0) != 0xff {
		return
	}
	if g.rb(0xa8f4) != 0 {
		return
	}
	n := g.rb(0xa8c6)
	if n == 0 {
		return
	}
	if n == 1 || g.rb(0xa8e0) != 0 {
		if g.rb(0xa840) != 0 {
			return
		}
	}
	d := g.rb(0xa8d6)
	e := d << 1
	shoot := false
	for i := 0; i < 2; i++ {
		if 0x84-g.Y(0)+d >= e || 0x78-g.Y(0x31)+d >= e {
			shoot = true
			break
		}
		g.next()
	}
	if !shoot {
		return
	}
	g.sndPushGame(g.k(0x07a2))
	h := g.aimLookup(0xac7f)
	k := g.rb(0xa8d4) + 1
	g.wb(0xa8d4, k)
	off := uint8(0x18)
	if k&1 == 0 {
		off = 0xe8
	}
	heading := off + h
	by, bx := g.Y(0x31), g.Y(0)
	if n == 1 || g.rb(0xa8e0) != 0 {
		g.at(0xa840, 0xaa18)
	} else {
		g.at(0xa8e0, 0xaa2c)
	}
	g.setY(0x31, by)
	g.setY(0, bx)
	de, bc := g.vec2(0x59d7, heading)
	g.setXw(10, de)
	g.setXw(12, bc)
	g.setY(1, 0x4d)
	g.setY(0x30, 0x62)
	g.decX(0)
	g.wb(0xa8f4, g.rb(0xa8f6))
}

// boss1b is routine 0x3dda: the escorting object of the second stage.
func (g *Game) boss1b() {
	if g.rb(vStage) != 1 {
		return
	}
	g.at(0xa8e0, 0xaa2c)
	a := g.X(0)
	if a == 0 {
		return
	}
	if a == 0xff {
		g.applyVelScroll()
		if !g.offScreen() {
			return
		}
	}
	g.killObj()
	g.setX(14, g.rb(0xa8f6))
}
