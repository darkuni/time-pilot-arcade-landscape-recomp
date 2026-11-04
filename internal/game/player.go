package game

// ---- the player's plane ------------------------------------------------------
//
// The player's record is at 0xa800 with sprite pair 0xaa10.  The plane itself
// never moves on screen; instead the world scrolls (velocity in 0xa808 for
// the horizontal axis and 0xa80a for the vertical one) and every other object
// adds that velocity to its position.

// playerInput returns the control byte of whoever is currently playing
// (0x1ed1): bit0 right, bit1 left, bit2 up, bit3 down, bit4 fire.
func (g *Game) playerInput() uint8 {
	if g.rb(vFlip) != 0 {
		return g.rb(vInP1)
	}
	return g.rb(vInP2)
}

// stageVectorBase selects the movement table for the current stage.
func (g *Game) stageVectorBase() uint16 {
	switch s := g.rb(vStage); {
	case s == 0:
		return 0x5e00
	case s < 3:
		return 0x2e3e
	}
	return 0x08fa
}

// playerUpdate is routine 0x1edf: steering and scrolling of the player.
func (g *Game) playerUpdate() {
	g.at(0xa800, 0xaa10)
	a := g.X(0)
	if a == 0 {
		return
	}
	if a != 0xff {
		g.playerExplode()
		return
	}
	if g.rb(vGameActive) == 0 {
		g.demoSteer()
		return
	}
	in := g.playerInput() & 15
	if in == 0 {
		g.playerMove()
		return
	}
	b := g.k(0x1f2e + uint16(in))
	c := g.X(2) - b
	if c == 0 {
		g.playerMove()
		return
	}
	d := uint8(4)
	if g.rb(vStage)&15 < 3 {
		d = 3
	}
	switch {
	case c+1 < 3:
		g.setX(2, b)
	case c >= 0x80:
		g.setX(2, c+d+b)
	default:
		g.setX(2, c-d+b)
	}
	g.playerMove()
}

// playerMove (0x1f42) turns the heading into the scroll velocity and picks
// the sprite for it.
func (g *Game) playerMove() {
	de, bc := g.vec(g.stageVectorBase(), g.X(2))
	g.ww(0xa808, 0-de)
	g.ww(0xa80a, 0-bc)
	g.playerSprite()
}

// playerSprite picks the plane's tile and attribute from its heading (0x20af).
func (g *Game) playerSprite() {
	idx := (g.rb(0xa802) + 4) >> 3 & 0x1f
	hl := uint16(0x20ce) + uint16(idx)
	g.wb(0xaa11, g.rb(hl))
	g.wb(0xaa40, g.rb(hl+0x20))
}

// demoSteer replays the recorded demonstration controls (0x214b).
func (g *Game) demoSteer() {
	for {
		b := g.rb(0xadf2)
		a := b & 0x3f
		if a == 0 || a == 1 {
			p := g.rw(0xadf3) + 1
			g.ww(0xadf3, p)
			g.wb(0xadf2, g.rb(p)+1)
			continue
		}
		g.wb(0xadf2, b-1)
		switch b >> 6 {
		case 0:
		case 1:
			g.wb(0xa802, g.rb(0xa802)-3)
		default:
			g.wb(0xa802, g.rb(0xa802)+3)
		}
		g.playerMove()
		return
	}
}

// playerExplode runs the death animation (0x2010).
func (g *Game) playerExplode() {
	if g.X(0) >= 0xb4 {
		g.setX(0, 0xb4)
		g.setY(1, 0xff)
		if g.rb(vStage) >= 2 {
			g.sndPushGame(g.k(0x07fe))
		}
		g.sndPushGame(g.k(0x0c5b))
		g.sndPushGame(g.k(0x0855))
		g.sndPushGame(g.k(0x1675))
		g.sndPushLoud(g.k(0x27cb))
		g.sndPushLoud(g.k(0x33a0))
	}
	n := g.decX(0)
	var de uint16
	switch n {
	case 0xb3:
		de = 0x1f76
	case 0xab:
		de = 0x1f94
	case 0xa3:
		de = 0x1fb2
	case 0x9b, 0x93:
		de = 0x1fd0
	case 0x8b:
		de = 0x1fb2
	case 0x83:
		de = 0x1fee
	default:
		return
	}
	g.drawExplosion(de)
}

// drawExplosion paints a block of explosion tiles onto the background layer
// at the plane's (fixed) screen position (0x2089).
func (g *Game) drawExplosion(de uint16) {
	hl := uint16(0xa5af)
	c := 0xc1 + g.rb(vStage)
	rows := g.k(0x337a)
	for ; rows > 0; rows-- {
		for cols := g.k(0x4902); cols > 0; cols-- {
			g.wb(hl, g.rb(de))
			g.wb(hl&^0x0400, c)
			hl++
			de++
		}
		hl += 0x1b
	}
}

// ---- the player's guns -------------------------------------------------------

// playerFire is routine 0x23e3: start a bullet when the fire button goes
// down, then move every bullet in flight.
func (g *Game) playerFire() {
	if g.rb(0xa800) == 0xff && g.rb(0xacc6) == 0 {
		in := g.playerInput()
		h := g.rb(0xa98e)<<1 | in>>4&1
		g.wb(0xa98e, h)
		if h&3 == 1 {
			g.wb(0xaa81, 3)
		}
		if g.rb(vGameActive) == 0 || g.rb(0xaa81) != 0 {
			if g.rb(0xaa82) == 0 {
				g.spawnBullet()
			}
		}
	}
	g.moveBullets()
}

// spawnBullet launches a bullet from the first free slot (0x2449).
func (g *Game) spawnBullet() {
	ix := uint16(0xaa80)
	stride := g.rw(0x0d46)
	for b := 0; ; b++ {
		if b == 6 {
			return
		}
		if g.rb(ix) == 0 {
			break
		}
		ix += stride
	}
	g.sndPushLoud(g.k(0x3270))
	g.emit(EvShot, 0, 0)
	vx := (0 - g.rw(0xa808)) << 2
	vy := (0 - g.rw(0xa80a)) << 2
	g.ww(ix+10, vx)
	g.ww(ix+12, vy)
	idx := (g.rb(0xa802) + 4) >> 3 & 0x1f
	w := g.rw(0x2771 + 2*uint16(idx))
	g.wb(ix, g.rb(ix)-1) // 0 -> 0xff
	g.wb(ix+3, 0)
	g.wb(ix+4, uint8(w))
	g.wb(ix+5, 0)
	g.wb(ix+6, uint8(w>>8))
	g.wb(0xaa81, g.rb(0xaa81)-1)
	g.wb(0xaa82, 6)
}

// moveBullets advances each bullet and draws it as background tiles (0x2496).
func (g *Game) moveBullets() {
	if v := g.rb(0xaa82); v != 0 {
		g.wb(0xaa82, v-1)
	}
	ix := uint16(0xaa80)
	for b := 0; b < 6; b++ {
		g.ix = ix
		st := g.X(0)
		if st != 0 {
			g.bulletStep(st)
		}
		ix += 0x10
	}
}

func (g *Game) bulletStep(st uint8) {
	kill := func() {
		g.setX(0, 0)
		g.setX(4, 0)
		g.setX(6, 0)
	}
	if st != 0xff {
		kill()
		return
	}
	hl := g.xw(10) + g.rw(0xa808) + (uint16(g.X(4))<<8 | uint16(g.X(3)))
	if uint8(hl>>8)+0x10 < 0x10 {
		kill()
		return
	}
	g.setX(4, uint8(hl>>8))
	g.setX(3, uint8(hl))
	hl = g.xw(12) + g.rw(0xa80a) + (uint16(g.X(6))<<8 | uint16(g.X(5)))
	if uint8(hl>>8)+8 < 0x18 {
		kill()
		return
	}
	g.setX(6, uint8(hl>>8))
	g.setX(5, uint8(hl))
	g.drawBulletTiles()
}

// fixupAdd appends a screen-cell patch to the per-frame tile fix-up list
// (colour-RAM address, tile, colour offset).
func (g *Game) fixupAdd(de uint16, tile, col uint8) {
	hl := g.rw(0xae00)
	g.wb(hl, uint8(de))
	hl = hl&0xff00 | uint16(uint8(hl)+1)
	g.wb(hl, uint8(de>>8))
	hl = hl&0xff00 | uint16(uint8(hl)+1)
	g.wb(hl, tile)
	hl = hl&0xff00 | uint16(uint8(hl)+1)
	g.wb(hl, col)
	hl = hl&0xff00 | uint16(uint8(hl)+1)
	g.ww(0xae00, hl)
}

// drawBulletTiles draws a bullet as up to four background tiles, picked by
// its pixel position within the 8x8 grid (0x5337).
func (g *Game) drawBulletTiles() {
	x := g.X(4) + 7
	d := uint16(0x28)
	d = d<<1 | uint16(x>>7&1)
	d = d<<1 | uint16(x>>6&1)
	e := (x << 2) & 0xe0
	y := g.X(6) + 7
	e += y >> 3 & 0x1f
	c := (y << 3) & 0x38
	b := uint8(0)
	if x&4 != 0 {
		b = 1
	}
	c += (x & 3) << 6
	hl := uint16(0x53d4) + (uint16(b)<<8 | uint16(c))
	de := d<<8 | uint16(e)
	put := func() {
		a := g.rb(hl)
		bb := g.rb(hl + 1)
		hl += 2
		if a != 0 {
			g.fixupAdd(de, a, bb)
		}
	}
	put()
	de++
	put()
	de += 0x1f
	put()
	de++
	put()
}
