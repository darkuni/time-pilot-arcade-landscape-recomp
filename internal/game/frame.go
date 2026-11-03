package game

// ---- the per-frame gameplay routine (state 3, sub-state 7) ---------------------

// playFrame is routine 0x1199.
func (g *Game) playFrame() {
	g.chooseAim()
	g.playerUpdate()
	g.playerFire()
	g.spawner()
	g.bonusPlane()
	g.stageBoss()
	g.enemyPlanes()
	g.updateClouds()
	g.missiles()
	g.boss1()
	g.boss1b()
	g.enemyBullets()
	g.parachutes()
	g.collisions()
	g.bossAlarm()
	g.extraLife()
	g.hitTicker()
	g.difficultyTick()
	g.drawProgress()
	switch g.rb(0xa800) {
	case 0xff:
		g.stageEndCheck()
	case 0:
		g.playerDead()
	}
}

// stageEndCheck is 0x1271: the player is alive; has the stage been won?
func (g *Game) stageEndCheck() {
	if g.rb(0xad02) != 0 {
		return
	}
	if g.rb(0xacc6) == 0 {
		return
	}
	for i := uint16(0); i < 15; i++ {
		if g.rb(0xa810+i*0x10) != 0 {
			return
		}
	}
	g.sndBurst7()
	if g.rb(vGameActive) == 0 {
		g.wb(0xacc6, g.k(0x07d1))
		g.hideSprites24()
		g.play12fb()
		return
	}
	for i := uint16(0); i < 0x17; i++ {
		g.wb(0xaa43+i*2, 0)
	}
	g.nextStage()
	g.saveLife()
	g.wb(vSub, g.k(0x4a35))
}

// saveLife copies the 16-byte player block at ad00 to the current player's
// save area (ad10 / ad20).
func (g *Game) saveLife() {
	dst := uint16(0xad10)
	if g.rb(vPlayer2) != 0 {
		dst = 0xad20
	}
	for i := uint16(0); i < 16; i++ {
		g.wb(dst+i, g.rb(0xad00+i))
	}
}

// hideSprites24 is 0x15b6.
func (g *Game) hideSprites24() {
	for i := uint16(0); i < 0x18; i++ {
		g.wb(0xaa41+i*2, 0)
	}
}

// nextStage is 0x2db8: advance to the next stage.
func (g *Game) nextStage() {
	g.wb(0xad01, g.rb(0xad01)+1)
	s := g.rb(vStage) + 1
	if s >= 5 {
		s = 0
	}
	g.wb(vStage, s)
	var a uint8
	switch n := g.rb(0xad01); {
	case n < 6:
		a = g.rb(0xa9d3)
	case n < 11:
		a = g.rb(0xa9d4)
	default:
		a = g.rb(0xa9d5)
	}
	g.wb(0xad0a, a)
	g.wb(0xad02, g.rb(0xa9cd))
	g.wb(0xad0d, 0)
	g.wb(0xacc6, 0)
	g.wb(0xad0e, 0xff)
}

// playerDead is the tail of 0x1199 at 0x11ed: the plane has finished
// exploding.
func (g *Game) playerDead() {
	g.hideSprites24()
	if g.rb(0xacc6) != 0 {
		g.nextStage()
	}
	g.sndBurst7()
	lives := g.rb(0xad00) - 1
	g.wb(0xad00, lives)
	g.saveLife()
	if lives == 0 {
		if g.rb(vGameActive) == 0 {
			g.play12fb()
			return
		}
		e := uint8(9)
		if g.rb(vPlayer2) != 0 {
			e = 10
		}
		g.enq(2, e)
		g.enq(0x0a, 0x0b)
		g.wb(vSubTimer, 0xb4)
		g.nextSub()
		return
	}
	other := uint16(0xad20)
	if g.rb(vPlayer2) != 0 {
		other = 0xad10
	}
	if g.rb(other) == 0 {
		g.wb(vSubTimer, 0x5a)
		g.wb(vSub, g.k(0x4b52))
		return
	}
	g.switchPlayer()
}

// difficultyTick is routine 0x4d3a: the BCD clock that drives the waves and
// periodically raises the enemy parameters.
func (g *Game) difficultyTick() {
	inc := func(a uint16) bool { // 0x4d67: true when it did not wrap
		v := bcdInc(g.rb(a))
		g.wb(a, v)
		if v < 0x60 {
			return true
		}
		g.wb(a, 0)
		return false
	}
	if inc(0xad05) {
		return
	}
	if !inc(0xad06) {
		inc(0xad07)
	}
	v := g.rb(0xa9d7)
	if v == 0 {
		return
	}
	v--
	g.wb(0xa9d7, v)
	if v != 0 {
		return
	}
	g.wb(0xa9d7, g.rb(0xa9d6))
	a := g.rb(0xacc0) + 1
	if a >= 0x10 {
		a = 15
	}
	g.wb(0xacc0, a)
	g.stageParams()
}

// extraLife is routine 0x4dde.
func (g *Game) extraLife() {
	if g.rb(vGameActive) == 0 {
		return
	}
	tab := uint16(0x4e1b)
	if g.rb(0xa9c3)&1 != 0 {
		tab = 0x4e30
	}
	n := g.rb(tab)
	score := g.rb(0xad35)
	if g.rb(vPlayer2) != 0 {
		score = g.rb(0xad38)
	}
	found := false
	for i := uint16(0); i < uint16(n); i++ {
		if g.rb(tab+1+i) == score {
			found = true
			break
		}
	}
	if !found {
		g.wb(0xad03, g.rb(0xad03)&^1)
		return
	}
	if g.rb(0xad03)&1 != 0 {
		return
	}
	g.wb(0xad03, g.rb(0xad03)|1)
	l := g.rb(0xad00)
	g.wb(0xad00, l+1)
	g.enq(5, l)
	g.sndPushGame(g.k(0x2d4e))
}

// bonusPlane is routine 0x47b3: the occasional flying bonus (parachutist).
func (g *Game) bonusPlane() {
	if g.rb(vStage) == 4 {
		return
	}
	g.at(0xa8f0, 0xaa2e)
	switch s := g.X(0); {
	case s == 0:
		if g.rb(0xad0d) != 0 || g.rb(vFrameCtr)&1 == 0 {
			return
		}
		if g.decX(14) != 0 {
			return
		}
		hl := 0x488d + uint16((g.rb(0xa802)+8)>>3&0x1e)
		g.setY(0x31, g.rb(hl))
		g.setY(0, g.rb(hl+1))
		g.setX(10, 0)
		g.setX(11, 0)
		g.setX(12, 0x40)
		g.setX(13, 0)
		g.setX(0, 0xff)
	case s == 0xff:
		g.applyVelScroll()
		if g.offScreen() {
			g.resetTimerObj()
			return
		}
		g.setY(1, g.rb(0x47ea+uint16(g.rb(vFrameCtr)>>4&7)))
		g.setY(0x30, 0x75)
	default:
		g.applyScroll()
		a := g.X(0)
		switch {
		case a == 0x10:
			g.decX(0)
			n := g.X(7)
			g.setX(7, n+1)
			if n >= 4 {
				g.enq(4, 0x0f)
			} else {
				g.enq(4, g.rb(0x484f+uint16(n)))
			}
		case a >= 0x3c:
			g.setX(0, 0x3b)
			g.sndPushGame(g.k(0x079b))
			if g.X(7) >= 4 {
				g.setY(1, 0x8f)
			} else {
				g.setY(1, g.rb(0x482d+uint16(g.X(7))))
			}
			g.setY(0x30, 0x6c)
		default:
			if g.decX(0) == 0 {
				g.resetTimerObj()
			}
		}
	}
}

// drawLives is queue command 5 (0x4d72): redraw the column of spare-plane
// icons at the left of the screen.
func (g *Game) drawLives(n uint8) {
	if g.rb(vGameActive) == 0 {
		return
	}
	de := uint16(0xa783)
	if n >= 7 {
		n = 6
	}
	for ; n > 0; n-- {
		g.lifeIcon(&de, 9, 0x18)
	}
	for de >= 0xa623 {
		g.wb(de, 0xf1)
		g.wb(de-1, 0xf1)
		g.wb(colorAddr(de-1), 0x10)
		g.wb(colorAddr(de), 0x10)
		de -= 0x20
	}
}

// lifeIcon is 0x4daf: one two-by-two plane icon.
func (g *Game) lifeIcon(de *uint16, b, c uint8) {
	d := *de
	g.wb(d, b+3)
	g.wb(d-1, b+2)
	g.wb(d-0x21, b)
	g.wb(d-0x20, b+1)
	cell := d - 0x20
	g.wb(colorAddr(cell), c)
	g.wb(colorAddr(cell-1), c)
	g.wb(colorAddr(d-1), c)
	g.wb(colorAddr(d), c)
	*de = d - 0x40
}
