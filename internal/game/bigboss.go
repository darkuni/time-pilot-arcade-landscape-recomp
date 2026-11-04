package game

// ---- the end-of-stage boss (record a8a0, second half a8b0) ---------------------

// stageBoss is routine 0x43b7.
func (g *Game) stageBoss() {
	if g.rb(0xacc6) == 0xff {
		return
	}
	if g.rb(0xad0d) == 0 {
		if g.rb(vFrameCtr)&7 != 5 {
			return
		}
		g.at(0xa8a0, 0xaa24)
		if g.rb(0xad02)|g.X(0)|g.X(0x10) != 0 {
			return
		}
		g.wb(0xad0d, 0xff)
		g.setX(4, 7)
		g.bossClear()
		return
	}
	g.at(0xa8a0, 0xaa24)
	switch s := g.X(0); {
	case s == 0:
		if g.X(14) == 0 {
			g.bossSpawn()
		} else {
			g.decX(14)
		}
	case s == 0xff:
		g.bossAlive()
	default:
		if g.X(4) != 0 {
			g.decX(4)
			g.setX(0, 0xff)
			g.soundBurst()
			g.bossAlive()
			return
		}
		if s != 0xef { // the original compares state+1 with 0xf0
			g.bossExplode()
			return
		}
		g.bossKilled()
	}
}

// bossClear is 0x46db: remove the boss from the screen.
func (g *Game) bossClear() {
	g.setX(0, 0)
	g.setY(0, 0)
	g.setY(2, 0)
	g.setY(0x31, 0)
	g.setY(0x33, 0)
	g.setX(14, 0x5f)
}

// bossAlive is 0x4403: move, draw and shoot.
func (g *Game) bossAlive() {
	hl := (uint16(g.X(12))<<8 | uint16(g.X(13))) + g.rw(0xa808) + (uint16(g.Y(0x31))<<8 | uint16(g.X(3)))
	g.setY(0x31, uint8(hl>>8))
	g.setX(3, uint8(hl))
	hl = (uint16(g.X(0x1c))<<8 | uint16(g.X(0x1d))) + g.rw(0xa80a) + (uint16(g.Y(0))<<8 | uint16(g.X(5)))
	g.setY(0, uint8(hl>>8))
	g.setX(5, uint8(hl))
	g.setY(0x33, g.Y(0x31)+0x10)
	g.setY(2, g.Y(0))
	if g.bossGone() {
		g.bossClear()
		return
	}
	g.bossPicture()
	g.bossShoot()
}

// bossPicture is the sprite choice at 0x444d.
func (g *Game) bossPicture() {
	st := g.rb(vStage)
	if st == 4 {
		e := g.X(4)
		keep := e == 7
		if !keep {
			v := g.incX(6)
			if v&0x80 != 0 {
				if v&0x7f >= 3 {
					g.setX(6, 0)
				}
				g.setY(0x30, 0x51)
				g.setY(0x32, 0x51)
				g.bossEngine()
				return
			}
			if e+2 < v {
				g.setX(6, 0x80)
			}
		}
		g.setY(0x30, 0x70)
		g.setY(0x32, 0x70)
		g.bossEngine()
		return
	}
	b := st<<4 + g.rb(vFrameCtr)&2
	a := (7 - g.X(4))
	a = (a>>1 | a<<7) & 3
	idx := a<<2 + b
	hl := 0x44f1 + uint16(idx)
	bb := g.rb(hl)
	c := g.rb(hl + 1)
	d := g.rb(0x4531 + uint16(st))
	if g.X(2)+0x40 >= 0x80 {
		g.setY(1, bb)
		g.setY(3, c)
		g.setY(0x30, d+0x80)
		g.setY(0x32, d+0x80)
	} else {
		g.setY(1, c)
		g.setY(3, bb)
		g.setY(0x30, d)
		g.setY(0x32, d)
	}
}

// bossEngine is 0x44dc: the flickering exhaust tiles of the last boss.
func (g *Game) bossEngine() {
	if g.rb(vFrameCtr)&4 != 0 {
		g.setY(1, 0xd5)
		g.setY(3, 0xd4)
	} else {
		g.setY(1, 0xd7)
		g.setY(3, 0xd6)
	}
}

// bossShoot is 0x46f0.
func (g *Game) bossShoot() {
	if g.X(0) != 0xff || g.rb(0xa817) != 0 {
		return
	}
	d := g.rb(0xa827)
	e := d << 1
	fire := false
	for i := 0; i < 2; i++ {
		inside := g.Y(0)+8 < 0x28 || g.Y(0x31)+0x10 < 0x20
		if !inside && (0x84-g.Y(0)+d >= e || 0x78-g.Y(0x31)+d >= e) {
			fire = true
			break
		}
		g.next()
	}
	if !fire {
		return
	}
	var hl, iy uint16
	found := false
	hl, iy = 0xa830, 0xaa16
	for i := 0; i < 2; i++ {
		if g.rb(hl) == 0 {
			found = true
			break
		}
		hl += 0x10
		iy += 2
	}
	if !found {
		return
	}
	g.ww(0xa991, hl)
	g.ww(0xa993, iy)
	g.sndPushGame(g.k(0x07a2))
	h := g.aimLookup(0xac7f)
	n := g.rb(0xa8b4) + 1
	g.wb(0xa8b4, n)
	off := uint8(0x18)
	if n&1 == 0 {
		off = 0xe8
	}
	a := off + h
	by, bx := g.Y(0x31), g.Y(0)
	g.at(hl, iy)
	g.setX(2, a)
	g.setY(0x31, by)
	g.setY(0, bx)
	var de, bc uint16
	if g.rb(vStage) < 2 {
		de, bc = g.vec2(0x59d7, a)
	} else {
		de, bc = g.vec2(0x5c00, a)
	}
	g.setXw(10, de)
	g.setXw(12, bc)
	g.setY(1, 0x4d)
	g.setY(0x30, 0x62)
	g.decX(0)
	g.wb(0xa817, g.rb(0xa814))
}

// bossExplode is 0x45b3.
func (g *Game) bossExplode() {
	g.applyScroll()
	b := g.Y(0x31)
	hide := b+0x13 < 3
	if !hide {
		g.setY(0x33, b+0x10)
		b = g.Y(0)
		if b+8 < 0x28 {
			hide = true
		} else {
			g.setY(2, b)
		}
	}
	if hide {
		g.setY(1, 0xff)
		g.setY(3, 0xff)
	}
	a := g.X(0)
	if a == 0xb4 {
		g.decX(0)
		g.setY(1, 0xfe)
		g.setY(3, 0xfd)
		g.setY(0x30, 0x6c)
		g.setY(0x32, 0x6c)
		if g.rb(0xa800) == 0xff {
			g.sndPushGame(g.k(0x49ee))
		}
		g.enq(4, 0x0d)
		g.emit(EvBoss, 0, g.scoreValue(0x0d))
		return
	}
	if a > 0xb4 {
		v := g.rb(0x461b + uint16(((a-0xb4)>>3-1)&7))
		g.setY(3, v)
		g.setY(1, v+1)
	}
	if g.decX(0) == 0 {
		g.wb(0xacc6, 0xff)
		g.setX(0, 0)
		return
	}
	if g.X(0) != 0x5a {
		return
	}
	g.setY(1, 0xff)
	g.setY(3, 0xff)
}

// bossKilled is 0x455a: the boss is dead, so every enemy blows up with it.
func (g *Game) bossKilled() {
	g.wb(0xa8dc, 0)
	g.sndBurst7()
	g.explosionSounds()
	c := uint8(0x14)
	for i := uint16(0); i < 15; i++ {
		p := 0xa810 + i*0x10
		switch g.rb(p) {
		case 0xff:
			g.wb(p, c)
			g.enq(4, 2)
		case 0xfe:
			g.wb(p, 0)
		}
		c += 10
	}
	g.wb(0xacc6, 0xfe)
	g.setX(0, 0xe4)
	g.setY(0x30, 0x3d)
	g.setY(0x32, 0x3d)
}

// sndBurst7 is 0x5634.
func (g *Game) sndBurst7() {
	for _, a := range []uint16{0x167c, 0x0a9c, 0x1484, 0x0c78, 0x07d3, 0x33b4} {
		g.sndPush(g.k(a))
	}
	g.sndPush(g.rb(vStage) + 0x8c)
}

// explosionSounds is 0x56d2 (+ 0x56e4).
func (g *Game) explosionSounds() {
	g.sndPushGame(g.k(0x0c5b))
	g.sndPushGame(g.k(0x0855))
	g.sndPushGame(g.k(0x1675))
	g.sndPushLoud(g.k(0x27cb))
	g.sndPushLoud(g.k(0x33a0))
}

// bossSpawn is 0x4663.
func (g *Game) bossSpawn() {
	if g.rb(0xacc6) != 0 {
		return
	}
	b := g.rb(0xa802)
	a := uint8(0x10)
	if g.rb(vFrameCtr)&8 == 0 {
		a = 0xf0
	}
	a += b
	hl := 0x3c84 + uint16(a>>2&0x3e)
	g.setY(0x31, g.rb(hl))
	g.setY(0, g.rb(hl+1))
	g.setX(2, (b+0xc0)&0x80)
	var base uint16
	switch g.rb(vStage) & 7 {
	case 0:
		base = 0x59d7
	case 1, 2:
		base = 0x5e00
	case 3:
		base = 0x2e3e
	default:
		base = 0x08fa
	}
	de, bc := g.vec(base, g.X(2))
	g.setX(12, uint8(de>>8))
	g.setX(13, uint8(de))
	g.setX(0x1c, uint8(bc>>8))
	g.setX(0x1d, uint8(bc))
	if g.X(4) < 6 {
		g.setX(4, 5)
	}
	g.setX(0, 0xff)
	g.sndPushGame(g.rb(vStage) + 12)
}
