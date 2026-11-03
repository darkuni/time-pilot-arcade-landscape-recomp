package game

// ---- enemy planes (records a850..a8b0) -----------------------------------------

// enemyPlanes is routine 0x28a1: run every enemy plane for this frame.
func (g *Game) enemyPlanes() {
	for i := uint16(0); i < 5; i++ {
		g.enemyPlane(0xa850+i*0x10, 0xaa1a+i*2)
	}
	if g.rb(0xad0d) != 0 {
		return
	}
	g.enemyPlane(0xa8a0, 0xaa24)
	g.enemyPlane(0xa8b0, 0xaa26)
}

// enemyPlane dispatches on the stage (0x290e) and on the object's state.
func (g *Game) enemyPlane(ix, iy uint16) {
	g.at(ix, iy)
	st := g.rb(vStage) & 7
	if st > 4 {
		return
	}
	switch g.X(0) {
	case 0:
		return
	case 0xff:
	case 0xfe:
		g.waitLaunch()
		return
	default:
		g.enemyDying()
		return
	}
	switch st {
	case 0:
		g.steer()
		g.advance(0x59d7, 1)
		if g.offScreen() {
			g.clearObj()
			return
		}
		g.enemyFire()
		g.spriteStage0()
		g.enemyShot()
	case 1:
		g.steer()
		g.advance(0x5e00, 1)
		if g.offScreen() {
			g.clearObj()
			return
		}
		g.enemyFire()
		g.spriteStage1()
	case 2:
		if g.rb(vFrameCtr)&3 < 3 {
			g.steer()
		}
		g.advance(0x59d7, 1)
		if g.offScreen() {
			g.clearObj()
			return
		}
		g.enemyFire()
		g.spriteStage2()
		g.enemyShot()
	case 3:
		g.steer()
		g.advance(0x08fa, 1)
		if g.offScreen() {
			g.clearObj()
			return
		}
		g.enemyFire()
		g.spriteStage3()
		g.enemyShot()
	case 4:
		g.stage4Move()
		if g.offScreen() {
			g.clearObj()
			return
		}
		g.spriteStage4()
		g.enemyFire()
		g.enemyShot()
	}
}

// waitLaunch (0x2b52): a plane waiting to enter counts down its delay.
func (g *Game) waitLaunch() {
	if g.decX(14) == 0 {
		g.incX(0)
		g.setX(14, 0x80)
	}
}

// steer turns the plane toward its wanted heading by the stage's turn rate
// (0x2bef).
func (g *Game) steer() {
	c := g.X(1) - g.X(2)
	if c+2 < 4 {
		return
	}
	b := g.X(2)
	rate := g.rb(0x2c1d + uint16(g.rb(vStage)))
	if c < 0x80 {
		g.setX(2, rate+b)
	} else {
		g.setX(2, b-rate)
	}
}

// stage4Move is 0x29f7.
func (g *Game) stage4Move() {
	if 0x78-g.Y(0x31)+0x48 < 0x90 || 0x84-g.Y(0x31)+0x48 < 0x90 {
		g.wb(vStage, 0)
		g.steer()
		g.wb(vStage, 4)
	} else {
		g.steer()
	}
	if g.rb(vFrameCtr)>>1&1 == 0 {
		g.advance(0x59d7, 2)
	} else {
		g.advance(0x2e3e, 1)
	}
}

// enemyShot is routine 0x4243 (kept in enemy2.go).

// spriteStage0..4 choose the picture for the plane from its heading.
func (g *Game) spriteStage0() { // 0x2a3c
	b, c := g.headingFrame()
	g.setY(0x30, c)
	g.setY(1, b)
}

func (g *Game) spriteStage1() { // 0x2a47
	b, c := g.headingFrame()
	g.setY(0x30, c+0x35)
	g.setY(1, b+0x10)
}

// headingFrame is 0x2a57.
func (g *Game) headingFrame() (b, c uint8) {
	i := uint16((g.X(2)+8)>>4) & 15
	b = g.rb(0x2a77 + i)
	c = g.rb(0x2a77 + i + 0x10)
	if g.rb(vFrameCtr)&2 != 0 {
		b += 8
	}
	return
}

func (g *Game) spriteStage2() { // 0x2a97
	i := uint16((g.X(2)+4)&0xf8>>2) & 0x3f
	hl := 0x2abc + i
	b := g.rb(hl)
	a := g.rb(vFrameCtr) & 2
	if a != 0 {
		a = 8
	}
	g.setY(1, a+b)
	g.setY(0x30, g.rb(hl+1))
}

func (g *Game) spriteStage3() { // 0x2afc
	i := uint16((g.X(2)+8)>>4) & 15
	g.setY(1, g.rb(0x2b18+i))
	g.setY(0x30, g.rb(0x2b18+i+0x10))
}

func (g *Game) spriteStage4() { // 0x2b38
	b := (g.rb(vFrameCtr)>>2)&3 + 0xd8
	g.setY(1, (g.X(4)-1)<<2+b)
	g.setY(0x30, 0x61)
}

// ---- enemy bullets (records a810..a840) ---------------------------------------

func (g *Game) enemyBullets() { // 0x3e36
	for i := uint16(0); i < 4; i++ {
		g.at(0xa810+i*0x10, 0xaa12+i*2)
		g.enemyBullet()
	}
}

func (g *Game) enemyBullet() { // 0x3e63
	switch s := g.X(0); {
	case s == 0:
		return
	case s == 0xff:
		g.bulletFly()
	default:
		if g.rb(vStage) != 4 {
			g.killObj()
			return
		}
		a := g.X(0)
		if a == 1 {
			g.killObj()
			return
		}
		g.decX(0)
		if a >= 0x3c {
			g.setX(0, 0x3b)
			g.soundBurst()
		}
		g.applyScroll()
		a = g.X(0)
		if a < 0x1c {
			return
		}
		g.setY(1, g.rb(0x3ec3+uint16((a-0x1c)>>2&7)))
		g.setY(0x30, 3)
	}
}

// bulletFly is 0x3e6c.
func (g *Game) bulletFly() {
	if g.rb(vStage) == 4 {
		g.setY(1, (g.rb(vFrameCtr)>>1&7)+0x40)
		g.setY(0x30, 0x44)
	}
	g.applyVelScroll()
	if g.offScreen() {
		g.killObj()
	}
}

// soundBurst is 0x5683.
func (g *Game) soundBurst() {
	g.sndPushLoud(g.k(0x07a6))
	g.sndPushLoud(g.k(0x4cda))
}

// enemyFire is 0x3ed6: let an enemy plane shoot at the player.
func (g *Game) enemyFire() {
	if g.rb(vFrameCtr)&7+5 != g.X(15) {
		return
	}
	if g.rb(0xa817) != 0 {
		return
	}
	b := g.rb(0xa844)
	if b == 0 {
		return
	}
	hl := uint16(0xa810)
	de := uint16(0xaa12)
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
	rng := g.rb(0xa827)
	c := rng << 1
	if !(0x78-g.Y(0x31)+rng >= c || 0x84-g.Y(0)+rng >= c) {
		return
	}
	cone := g.rb(0xa837)
	if g.rb(0xa802)-g.X(2)+cone >= cone<<1 {
		return
	}
	h := g.aimLookup(0xac7f)
	if h-g.X(2)+0x10 >= 0x20 {
		return
	}
	if g.rb(vStage) < 3 {
		g.sndPushGame(g.k(0x07a2))
	} else {
		g.sndPushGame(g.k(0x4c9f))
	}
	d, e := g.Y(0x31), g.Y(0)
	g.withObj(g.rw(0xa991), g.rw(0xa993), func() {
		g.setY(0x31, d)
		g.setY(0, e)
		var de, bc uint16
		if g.rb(vStage) != 0 {
			de, bc = g.vec2(0x5e00, h)
		} else {
			de, bc = g.vec2(0x5c00, h)
		}
		g.setXw(10, de)
		g.setXw(12, bc)
		g.setY(1, 0x4d)
		g.setY(0x30, 0x62)
		g.wb(0xa817, g.rb(0xa814))
		g.decX(0)
	})
}

// ---- stage-0 helper objects (records a8c0..a8e0) ---------------------------------

func (g *Game) parachutes() { // 0x3fea
	if g.rb(vStage) != 0 {
		return
	}
	g.at(0xa8c0, 0xaa28)
	for i := 0; i < 3; i++ {
		switch g.X(0) {
		case 0:
		case 0xff:
			g.parachuteFall()
		default:
			g.parachuteBurst()
		}
		g.next()
	}
}

func (g *Game) parachuteFall() { // 0x4017
	hl := uint16(0xfe80)
	if g.X(1) == 0 {
		hl = 0x0180
	}
	hl += uint16(g.Y(0x31))<<8 | uint16(g.X(3))
	hl += g.rw(0xa808)
	g.setY(0x31, uint8(hl>>8))
	g.setX(3, uint8(hl))
	hl = g.xw(7) + 9
	g.setXw(7, hl)
	hl += uint16(g.Y(0))<<8 | uint16(g.X(5))
	hl += g.rw(0xa80a)
	g.setY(0, uint8(hl>>8))
	g.setX(5, uint8(hl))
	if g.Y(0x31)+0x10 < 0x20 {
		g.killObj()
		return
	}
	if g.Y(0) >= 0xf8 {
		g.killObj()
	}
}

func (g *Game) parachuteBurst() { // 0x406c
	if g.X(0) >= 0x3c {
		g.setX(0, 0x3b)
		g.sndPushGame(g.k(0x2d87))
	}
	if g.decX(0) == 0 {
		g.killObj()
		return
	}
	g.applyScroll()
	a := g.X(0)
	if a < 0x1c {
		return
	}
	g.setY(1, g.rb(0x4094+uint16((a-0x1c)>>2&15)))
	g.setY(0x30, 14)
}

// bossAlarm is 0x40b8.
func (g *Game) bossAlarm() {
	if g.rb(vStage) < 2 {
		return
	}
	if g.rb(vFrameCtr)&0x1f != 0 {
		return
	}
	if g.rb(0xa8c0) == 0xff || g.rb(0xa8d0) == 0xff || g.rb(0xa8e0) == 0xff {
		return
	}
	g.sndPushGame(g.k(0x07fe))
}
