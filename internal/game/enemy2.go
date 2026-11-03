package game

// enemyShot is routine 0x4243: a plane lets go of a parachutist / missile /
// mine, depending on the stage.
func (g *Game) enemyShot() {
	if g.rb(vFrameCtr)&7+5 != g.X(15) {
		return
	}
	if v := g.rb(0xa8f4); v != 0 {
		g.wb(0xa8f4, v-1)
		return
	}
	b := g.rb(0xa8c6)
	if b == 0 {
		return
	}
	hl := uint16(0xa8c0)
	de := uint16(0xaa28)
	for ; ; b-- {
		if g.rb(hl) == 0 {
			break
		}
		hl += 0x10
		de += 2
		if b == 1 {
			return
		}
	}
	g.ww(0xa991, hl)
	g.ww(0xa993, de)
	d := g.rb(0xa8d6)
	c := d << 1
	if !(0x78-g.Y(0x31)+d >= c || 0x84-g.Y(0)+d >= c) {
		return
	}
	c = g.X(2)
	if g.rb(vStage) == 0 {
		d = g.rb(0xa8e6)
		cc := d << 1
		if 0x84-g.Y(0)+d >= cc {
			return
		}
		if 0x78 < g.Y(0x31) {
			c = 1
		} else {
			c = 0
		}
	}
	dy, ex := g.Y(0x31), g.X(3)
	hy, lx := g.Y(0), g.X(5)
	st := g.rb(vStage)
	g.withObj(g.rw(0xa991), g.rw(0xa993), func() {
		g.setX(3, ex)
		g.setY(0x31, dy)
		g.setX(5, lx)
		g.setY(0, hy)
		g.setX(1, c)
		if st == 4 {
			g.setX(4, g.rb(0xa8e6))
		}
		switch {
		case st == 0:
			g.setY(1, 0x4f)
			a := c>>1 | c<<7 // RRCA
			a = uint8(int8(a) >> 1)
			g.setY(0x30, a&0xc0+11)
			g.setX(7, 0)
			g.setX(8, 0xff)
			g.decX(0)
			g.wb(0xa8f4, g.rb(0xa8f6))
			g.sndPushGame(g.k(0x16de))
		case st == 3:
			a := c
			if (c+0x40)&0x80 != 0 {
				g.setX(2, a-0x1a)
			} else {
				g.setX(2, a+0x1a)
			}
			de, bc := g.vec2(0x59d7, g.X(2))
			g.setXw(10, de)
			g.setXw(12, bc)
			g.setX(2, c)
			g.spriteFromHeading()
			g.setX(14, 0x20)
			g.decX(0)
			g.wb(0xa8f4, g.rb(0xa8f6))
			g.sndPushGame(g.k(0x07d8))
			g.sndPushGame(g.k(0x276b))
		case st > 3:
			h := g.aimLookup(0xac7f)
			g.setX(1, h)
			g.setX(2, h)
			g.spriteFromHeading()
			g.decX(0)
			g.wb(0xa8f4, g.rb(0xa8f6))
			g.setX(14, 0)
			g.sndPushGame(g.k(0x276b))
		default:
			h := g.aimLookup(0xac7f)
			g.setX(1, h)
			g.setX(2, (g.X(15)>>1|g.X(15)<<7)&0x80+0x40+h)
			g.spriteFromHeading()
			g.decX(0)
			g.wb(0xa8f4, g.rb(0xa8f6))
			g.setX(14, 0)
			g.sndPushGame(g.k(0x07d8))
			g.sndPushGame(g.k(0x276b))
		}
	})
}

// spriteFromHeading is 0x3faf.
func (g *Game) spriteFromHeading() {
	i := uint16((g.X(2)+8)>>4) & 15
	g.setY(1, g.rb(0x3fca+i))
	g.setY(0x30, g.rb(0x3fca+i+0x10))
}

// missiles is routine 0x40d6: the objects dropped or fired by planes on
// stages 3-5.
func (g *Game) missiles() {
	st := g.rb(vStage)
	if st < 2 {
		return
	}
	b := g.rb(0xa8c6)
	if b == 0 {
		return
	}
	g.at(0xa8c0, 0xaa28)
	for ; b > 0; b-- {
		switch s := g.X(0); {
		case s == 0:
		case s != 0xff:
			g.missileBurst()
		case st == 4:
			g.mineStep()
		case g.X(14) != 0:
			g.bulletFly()
			g.decX(14)
		default:
			g.missileFly()
		}
		g.next()
	}
}

func (g *Game) missileFly() { // 0x4117
	if g.rb(vFrameCtr)&15 == g.X(15) {
		g.setX(1, g.aimLookup(0xac7f))
	}
	g.turnSlow(1)
	g.advance(0x59d7, 2)
	g.spriteFromHeading()
	if g.offScreen() {
		g.killObj()
	}
}

// turnSlow is 0x4201 (step 1) / 0x421f (step 2): turn toward the wanted
// heading.
func (g *Game) turnSlow(step uint8) {
	c := g.X(1) - g.X(2) + 1
	if c < 2 {
		return
	}
	if c >= 0x80 {
		g.setX(2, g.X(2)-step)
	} else {
		g.setX(2, g.X(2)+step)
	}
}

func (g *Game) missileBurst() { // 0x413c
	if g.X(0) >= 0x3c {
		g.setX(0, 0x3b)
		g.sndPushGame(g.k(0x2d87))
	}
	g.applyScroll()
	if g.decX(0) == 0 {
		g.killObj()
		return
	}
	a := g.X(0)
	if a < 0x1c {
		return
	}
	d := uint16((a - 0x1c) >> 2 & 7)
	if g.rb(vStage) < 4 {
		g.setY(1, g.rb(0x416e+d))
		g.setY(0x30, 13)
	} else {
		g.setY(1, g.rb(0x4183+d))
		g.setY(0x30, 2)
	}
}

// mineStep is 0x4194 (stage 5 only).
func (g *Game) mineStep() {
	if g.X(4) != 0 {
		g.decX(4)
		g.mineSteer()
		return
	}
	g.advance(0x5e00, 2)
	g.mineSprite()
	if g.offScreen() {
		g.killObj()
	}
}

func (g *Game) mineSprite() { // 0x41f1
	g.setY(1, (g.rb(vFrameCtr)>>1&7)+0x50)
	g.setY(0x30, 10)
}

func (g *Game) mineSteer() { // 0x41b8
	if g.rb(vFrameCtr)&15 == 0 {
		hl := uint16(0xac75)
		if g.X(15)&1 == 0 {
			hl = 0xac79
		}
		b := g.aimLookup(hl)
		if g.aimD < 0x10 && g.aimE < 0x10 {
			g.setX(4, 0)
		}
		g.setX(1, b)
	}
	if g.rb(vFrameCtr)&3 != 0 {
		g.turnSlow(2)
	}
	g.advance(0x5e00, 2)
	g.mineSprite()
}

// enemyDying handles an enemy plane that has been hit (0x2b93): it drifts
// along while a ring of explosion pictures plays.
func (g *Game) enemyDying() {
	a := g.X(0)
	if a == 0xf0 {
		g.setX(0, 0x3b)
		g.explosionTick()
		return
	}
	if a == 0x3c {
		g.explosionTick()
	}
	if a >= 0x3c {
		g.decX(0)
		g.advance(0x59d7, 1)
		return
	}
	if g.decX(0) == 0 {
		g.clearObj()
		return
	}
	// 0x2c22
	if g.X(0) >= 0x20 {
		g.decX(0)
		g.advance(0x59d7, 1)
	} else {
		g.applyScroll()
	}
	a = g.X(0)
	switch {
	case a >= 0x2a:
		g.setY(0x30, g.Y(0x30)&0xc0+g.rb(vFrameCtr)&15)
	case a >= 10:
		g.setY(1, g.rb(0x2c94+uint16((a-10)>>1&15)))
		g.setY(0x30, 0x3c)
	default:
		v := g.rb(0xa821)
		if v&0x80 == 0 || v&0x7f != g.X(15) {
			g.clearObj()
			return
		}
		if g.rb(vFrameCtr)&7 != 0 {
			g.incX(0)
		}
		g.setY(1, 0xfc)
		g.setY(0x30, 0x6c)
		if g.X(0) != 1 {
			return
		}
		g.enq(4, 0x0c)
		g.wb(0xa821, 0)
	}
}

// explosionTick is 0x2bba.
func (g *Game) explosionTick() {
	g.soundBurst()
	if v := g.rb(0xad02); v != 0 {
		g.wb(0xad02, v-1)
	}
	if g.X(14)&0x80 == 0 {
		return
	}
	if g.rb(0xa812) == 0 {
		return
	}
	v := g.rb(0xa811) - 1
	g.wb(0xa811, v)
	if v != 0 {
		return
	}
	g.wb(0xa821, g.X(15)+0x80)
}
