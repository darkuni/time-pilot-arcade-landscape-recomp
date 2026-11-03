package game

// ---- level set-up and the background clouds ----------------------------------

// levelInit is routine 0x19f0: reset the scroll, the player and every object
// record for the start of a life, then load this stage's parameters.
func (g *Game) levelInit() {
	g.ww(0xa808, 0)
	g.ww(0xa80a, 0)
	g.ww(0xad06, 0)
	g.wb(0xad0d, 0)
	g.wb(0xa8f7, 0)
	g.wb(0xad05, 0)
	g.wb(0xa9d7, g.rb(0xa9d6))
	g.wb(0xacc0, g.rb(0xad0a))
	g.wb(0xaa81, 0)
	g.wb(0xacc6, 0)
	g.wb(0xa802, 0x80)
	g.wb(0xa801, 0)
	g.wb(0xa800, 0xff)
	g.wb(0xaa41, 0x78)
	g.wb(0xaa10, 0x84)
	g.playerSprite()
	g.clearBullets()
	g.at(0xa8c0, 0xaa28)
	g.bossReset()
	g.at(0xa8e0, 0xaa2c)
	g.killObj()
	g.setX(14, g.rb(0xa8f6))
	g.at(0xa8f0, 0xaa2e)
	g.resetTimerObj()
	for i := 0; i < 7; i++ {
		g.clearObj()
		g.next()
	}
	g.initSlots()
	g.at(0xa800, 0xaa28)
	for _, o := range []uint16{0, 2, 4, 6} {
		g.setY(o, 0)
		g.setY(0x31+o, 0)
	}
	g.stageSetup()
	g.stageParams()
}

// clearBullets zeroes the state and position byte of the six bullet slots
// (0x2755).
func (g *Game) clearBullets() {
	for i := uint16(0); i < 6; i++ {
		g.wb(0xaa80+i*0x10, 0)
		g.wb(0xaa84+i*0x10, 0)
	}
}

// bossReset is 0x3c0d.
func (g *Game) bossReset() {
	g.setX(0, 0)
	g.setX(0x10, 0)
	g.setY(0, 0)
	g.setY(0x31, 0)
	g.wb(0xaa5b, 0)
	g.wb(0xaa2a, 0)
	g.setX(14, 0x80)
}

// resetTimerObj is 0x48ad.
func (g *Game) resetTimerObj() {
	g.setX(0, 0)
	g.setY(0, 0)
	g.setY(0x31, 0)
	g.setX(14, 0xf0)
}

// initSlots is 0x1ae4: free the 23 general records and number them 1..23.
func (g *Game) initSlots() {
	for i := uint16(0); i < 0x17; i++ {
		p := 0xa810 + i*0x10
		g.wb(p, 0)
		g.wb(p+15, uint8(i+1))
	}
}

// stageParams is 0x1a9a: per-stage, per-level enemy parameters.
func (g *Game) stageParams() {
	a := (g.rb(vStage)<<4)&0xf0 + g.rb(0xacc0)
	de := g.rw(0x1b04 + 2*uint16(a))
	r := func(i uint16) uint8 { return g.rb(de + i) }
	g.wb(0xa844, r(0))
	g.wb(0xa837, r(1))
	g.wb(0xa827, r(2))
	g.wb(0xa817, r(3))
	g.wb(0xa814, r(3))
	g.wb(0xacc1, r(4))
	g.wb(0xacc4, r(5))
	g.wb(0xa8c6, r(6))
	g.wb(0xa8d6, r(7))
	g.wb(0xa8e6, r(8))
	g.wb(0xa8f4, r(9))
	g.wb(0xa8f6, r(9))
}

// stageSetup is 0x30a5: initial cloud sprites for the stage.
func (g *Game) stageSetup() {
	st := g.rb(vStage)
	hl := 0x3176 + uint16(st)*8
	for i := uint16(0); i < 8; i++ {
		g.wb(0xaa31+i*2, g.rb(hl+i))
	}
	attr := uint8(0xcc)
	if st == 4 {
		attr = 0x28
	}
	for i := uint16(0); i < 8; i++ {
		g.wb(0xaa60+i*2, attr)
	}
	if st >= 4 {
		g.at(0xa900, 0xaa30)
		for i := uint16(0); i < 8; i++ {
			g.wb(0xaa30+i*2+0x31, g.rb(0x315e+i*2))
			g.wb(0xaa30+i*2, g.rb(0x315e+i*2+1))
		}
	} else {
		for i := uint16(0); i < 4; i++ {
			y := g.rb(0x316e + i*2)
			x := g.rb(0x316e + i*2 + 1)
			p := 0xaa30 + i*4
			g.wb(p+0x31, y)
			g.wb(p+0x33, y+0x10)
			g.wb(p, x)
			g.wb(p+2, x)
		}
	}
	g.updateClouds()
}

// updateClouds is 0x2cbc: scroll the background clouds at different speeds
// to give an impression of depth.
func (g *Game) updateClouds() {
	g.at(0xa900, 0xaa30)
	switch st := g.rb(vStage); {
	case st == 0:
		g.cloudA()
		g.cloudB()
		g.cloudB()
		g.cloudD()
	case st == 4:
		g.cloudC()
		g.cloudC()
		g.cloudE()
		g.cloudE()
		g.cloudD()
		g.cloudD()
	default:
		g.cloudF()
		g.cloudB()
		g.cloudB()
		g.cloudD()
	}
}

// scrollBy moves the current object by the scroll velocity times a factor
// given as a function of the velocity.
func (g *Game) scrollBy(f func(v uint16) uint16) {
	hl := f(g.rw(0xa808)) + (uint16(g.Y(0x31))<<8 | uint16(g.X(3)))
	g.setY(0x31, uint8(hl>>8))
	g.setX(3, uint8(hl))
	hl = f(g.rw(0xa80a)) + (uint16(g.Y(0))<<8 | uint16(g.X(5)))
	g.setY(0, uint8(hl>>8))
	g.setX(5, uint8(hl))
}

func sar(v uint16, n uint) uint16 { return uint16(int16(v) >> n) }

func (g *Game) scroll125() { g.scrollBy(func(v uint16) uint16 { return v + sar(v, 2) }) } // 0x2d6e
func (g *Game) scroll075() { g.scrollBy(func(v uint16) uint16 { return v - sar(v, 2) }) } // 0x2d93
func (g *Game) scroll050() { g.scrollBy(func(v uint16) uint16 { return v - sar(v, 1) }) } // 0x2df4

// pairSecond is 0x3058: place the second half of a two-sprite cloud.
func (g *Game) pairSecond() {
	b, c := g.Y(0x31), g.Y(0)
	g.setY(0x33, b+0x10)
	g.setY(2, c)
	g.next()
}

// pairSecondFar is 0x308a.
func (g *Game) pairSecondFar() {
	hl := 0xf010 + (uint16(g.Y(0x31))<<8 | uint16(g.Y(0)))
	g.setY(0x33, uint8(hl>>8))
	g.setY(2, uint8(hl))
	g.next()
}

func (g *Game) cloudA() { g.scroll125(); g.pairSecond(); g.pairSecond(); g.next() }    // 0x2d15
func (g *Game) cloudF() { g.scroll125(); g.pairSecond(); g.pairSecondFar(); g.next() } // 0x2d21
func (g *Game) cloudC() { g.scroll125(); g.pairSecond(); g.next() }                    // 0x2d2d
func (g *Game) cloudB() { g.scroll075(); g.pairSecond(); g.next() }                    // 0x2d36
func (g *Game) cloudE() { g.scroll075(); g.next() }                                    // 0x2d62
func (g *Game) cloudD() { g.scroll050(); g.next() }                                    // 0x2d68
