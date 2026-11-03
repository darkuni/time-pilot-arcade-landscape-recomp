package game

// ---- state 3: the game proper -----------------------------------------------
//
// Sub-states (a9ac) of the play state, in the order they normally occur:
//
//	0  0x27b1  new game / demo set-up
//	1  0x335e  choose the stage palette
//	2  0x5bd7  wipe the screen
//	3  0x4c75  draw lives / stage counters
//	4  0x0774  "get ready" caption and level initialisation
//	5  0x16af  fly-in (the plane drifts in while the caption is shown)
//	6  0x5694  short run-in before control is handed over
//	7  0x1199  normal play (one call per video frame)
//	8  0x330b  player died: high-score check
//	9  0x08b4  high-score entry screen set-up
//	10 0x18c3  high-score initials entry
//	11 0x12e2  next player / game over check
//	12 0x12fb  return to attract mode
//	13 0x4a0f  "stage clear" screen set-up
//	14 0x1323  "stage clear" animation
//	15 0x15b5  (unused)

// loadLevelParams copies one 4-byte row of the difficulty table at 0x186a
// into a9d3..a9d6 (routine 0x0f7b).
func (g *Game) loadLevelParams(idx uint8) {
	src := uint16(0x186a) + uint16(idx*4)
	for i := uint16(0); i < 4; i++ {
		g.wb(0xa9d3+i, g.rb(src+i))
	}
}

// play27b1: sub-state 0, new game / demo.
func (g *Game) play27b1() {
	g.sndPushGame(g.k(0x1767))
	g.wb(0xac64, 0x78)
	g.wb(0xac65, 0x84)
	g.ww(0xad16, 0)
	g.ww(0xad26, 0)
	g.wb(0xad12, g.rb(0xa9cd))
	g.wb(0xad22, g.rb(0xa9cd))
	for _, a := range []uint16{0xad14, 0xad24, 0xad32, 0xad13, 0xad23, 0xad1d, 0xad2d, 0xad0c} {
		g.wb(a, 0)
	}
	for _, a := range []uint16{0xad11, 0xad21, 0xad1e, 0xad2e} {
		g.wb(a, 1)
	}
	if g.rb(vGameActive) == 0 {
		g.demoSetup()
		return
	}
	// real game: clear both scores
	g.wb(vScoreP1, 0)
	g.ww(vScoreP1+1, 0)
	g.wb(vScoreP2, 0)
	g.ww(vScoreP2+1, 0)
	g.enq(4, 0)
	g.loadLevelParams(g.rb(0xa9c4))
	g.wb(0xad1a, g.rb(0xa9d3))
	g.wb(0xad2a, g.rb(0xa9d3))
	g.wb(vSubTimer, 0x96)
	g.nextSub()
}

// demoSetup is the attract-mode branch of 0x27b1 (0x2835).
func (g *Game) demoSetup() {
	a := g.rb(0xa9d0) + 1
	if a >= 4 {
		a = 1
	}
	g.wb(0xa9d0, a)
	g.wb(0xad14, a)
	g.wb(0xad11, a+1)
	g.wb(vFrameCtr, 0)
	g.wb(vNmiTick, 0)
	g.wb(0xa9cf, 0)
	// 0x4b67 copies 17 bytes of (now meaningless) data to 0xab30
	for i := uint16(0); i < 0x11; i++ {
		g.wb(0xab30+i, g.rb(0x4b84+i))
	}
	for i := uint16(0); i < 0x60; i++ {
		g.wb(0xaa80+i, 0)
	}
	for i := uint16(0); i < 0x180; i++ {
		g.wb(0xa800+i, 0)
	}
	g.loadLevelParams(2)
	g.wb(0xad1a, g.rb(0xa9d3))
	g.wb(0xad2a, g.rb(0xa9d3))
	for i := uint16(0); i < 16; i++ {
		g.wb(0xac74+i, 0x80)
	}
	g.wb(vSubTimer, 0x5a)
	g.nextSub()
}

// play335e: sub-state 1.  Picks this stage's palette (ad0b/ad0c) from the
// table at 0x0f8d and stores it for the current player.
func (g *Game) play335e() {
	var de uint16
	var a uint8
	if g.rb(vPlayer2) == 0 {
		de, a = 0xad1b, g.rb(0xad14)
	} else {
		de, a = 0xad2b, g.rb(0xad24)
	}
	hl := uint16(0x0f8d) + uint16(a*2)
	v := g.rb(hl)
	g.wb(de, v)
	g.wb(0xad0b, v)
	hl++
	de++
	v = g.rb(hl)
	g.wb(de, v)
	same := v == g.rb(0xad0c)
	g.wb(0xad0c, v)
	if same {
		g.nextSub()
	}
	g.wipeInit()
	g.nextSub()
}

// play5bd7: sub-state 2, run the screen wipe to its end.
func (g *Game) play5bd7() {
	g.run07d2()
	if !g.wipeStep() {
		return
	}
	g.nextSub()
}

// play330b: sub-state 8, the player is out: see whether the score earns a
// place in the high-score table.
func (g *Game) play330b() {
	n := g.rb(vSubTimer) - 1
	g.wb(vSubTimer, n)
	if n != 0 {
		return
	}
	if g.run4cc3() { // no new record
		g.enq(3, 9)
		g.enq(3, 11)
		g.wb(vSub, g.k(0x0843))
		g.nextPlayerCheck()
		return
	}
	g.sndPushGame(g.k(0x18fa))
	g.wb(0xad0c, 0)
	g.wb(0xad0b, 0xf1)
	g.wipeInit()
	g.nextSub()
}

// play12e2: sub-state 11, wait, then decide who plays next.
func (g *Game) play12e2() {
	n := g.rb(vSubTimer) - 1
	g.wb(vSubTimer, n)
	if n != 0 {
		return
	}
	g.nextPlayerCheck()
}

// nextPlayerCheck is 0x12e7: continue with the other player if he still has
// lives, otherwise advance to the game-over sub-state.
func (g *Game) nextPlayerCheck() {
	hl := uint16(0xad20)
	if g.rb(vPlayer2) != 0 {
		hl = 0xad10
	}
	if g.rb(hl) != 0 {
		g.switchPlayer()
		return
	}
	g.nextSub()
}

// switchPlayer is 0x1226.
func (g *Game) switchPlayer() {
	g.wb(vPlayer2, (g.rb(vPlayer2)+1)&1)
	g.wb(vSubTimer, 0x5a)
	g.wb(vSub, g.k(0x4b52))
}

// play12fb: sub-state 12, back to attract mode.
func (g *Game) play12fb() {
	g.wb(vGameActive, 0)
	g.wb(vSub, 0)
	g.wb(vPlayer2, 0)
	g.wb(vState, g.k(0x16d3))
	a := g.k(0x4901)
	hl := g.rw(0x4902) + uint16(a)
	r := uint8(hl) ^ uint8(hl>>8)
	g.wb(vSub, r-0x9b)
}

// play0774: sub-state 4, caption and level initialisation.
func (g *Game) play0774() {
	if g.rb(vGameActive) != 0 {
		de := g.rw(0x125b)
		e := uint8(de)
		if g.rb(vPlayer2) != 0 {
			e++
		}
		g.enq(uint8(de>>8), e)
		if g.rb(0xad0e) != 0 {
			g.enq(7, e)
		} else {
			// (the original tests ad0e a second time; same result)
			g.enq(2, 2)
		}
	} else {
		g.enq(2, 2)
	}
	g.drawProgress()
	g.levelInit()
	g.nextSub()
}

// drawProgress is 0x0809: the vertical bar at the left of the screen that
// shows how far through the stage the player is.
func (g *Game) drawProgress() {
	s := g.rb(vStage)
	hl := uint16(0x087c) + uint16(s)*10
	b := g.rb(hl)
	c := g.rb(hl + 1)
	hl += 2
	cnt := g.rb(0xad02)
	partial := g.rb(hl + uint16(cnt&7))
	pos := uint16(0xa79f)
	a := cnt >> 2 & 0x1f
	if a != 0 {
		for {
			g.wb(pos, b)
			pos -= 0x20
			a--
			if a == 0 {
				break
			}
			g.wb(pos, c)
			pos -= 0x20
			a--
			if a == 0 {
				break
			}
		}
	}
	g.wb(pos, partial)
	pos -= 0x20
	g.wb(pos, 0xf1)
}

// play16af: sub-state 5, the fly-in.
func (g *Game) play16af() {
	g.playerUpdate()
	g.updateClouds()
	if g.rb(vFrameCtr)&1 != 0 {
		n := g.rb(vSubTimer) - 1
		g.wb(vSubTimer, n)
		if n == 0 {
			g.enq(3, 9)
			g.enq(3, 14)
			g.enq(3, 0x1a)
			g.wb(0xad0e, 0)
			g.wb(vSubTimer, 0x2a)
			g.nextSub()
			return
		}
	}
	if g.rb(0xad0e) == 0 {
		return
	}
	var d uint8
	switch g.rb(vFrameCtr) & 15 {
	case 0:
		d = 2
	case 5:
		d = 10
	case 10:
		d = 11
	default:
		return
	}
	g.enq(d, g.rb(vStage)+0x1a)
}

// play5694: sub-state 6, a short run-in before play starts.
func (g *Game) play5694() {
	g.playerUpdate()
	g.updateClouds()
	g.playerFire()
	n := g.rb(vSubTimer) - 1
	g.wb(vSubTimer, n)
	if n != 0 {
		return
	}
	g.nextSub()
}

// play4a0f: sub-state 13, set up the "stage clear" bonus screen.
func (g *Game) play4a0f() {
	g.wb(0xa9f0, g.k(0x3213))
	g.wb(0xa9f1, 0)
	g.wb(0xa9f2, 0xff)
	g.wb(0xa9f3, 4)
	g.wb(0xa9f4, 0xff)
	g.wb(0xa9f6, 8)
	g.ww(0xa9f7, 0x56f1)
	hl := uint16(0xa400)
	for i := 0; i < 13; i++ {
		g.wb(hl, 0x14)
		hl++
	}
	g.wb(hl, 0)
	hl++
	g.wb(hl, 0)
	hl++
	for i := 0; i < 13; i++ {
		g.wb(hl, 0x14)
		hl++
	}
	for i := 0; i < 4; i++ {
		g.wb(hl, 14)
		hl++
	}
	c := g.rb(0xad0c)
	g.fillColumn(0xa7b1&^0x0400, 0xa0+c)
	g.fillColumn(0xa5d1&^0x0400, 0x20+c)
	g.wb(0xa610&^0x0400, 0xa0+c)
	g.wb(0xa610&^0x0400-0x20, 0x20+c) // HL+DE with DE=0xffe0 from fillColumn
	g.wb(0xa612&^0x0400, 0xe0+c)
	g.wb(0xa612&^0x0400-0x20, 0x60+c)
	g.wb(0xa611&^0x0400, 0xa0+c)
	g.wb(0xa611&^0x0400-0x20, 0x20+c)
	g.copyBonusLine()
	g.nextSub()
}

// fillColumn writes a value to 13 cells going up the screen (0x1319).
func (g *Game) fillColumn(hl uint16, a uint8) {
	for i := 0; i < 13; i++ {
		g.wb(hl, a)
		hl += 0xffe0
	}
}

// copyBonusLine is 0x339c: store the stage's bonus value pair.
func (g *Game) copyBonusLine() {
	var de uint16
	var a uint8
	if g.rb(vPlayer2) == 0 {
		de, a = 0xad1b, g.rb(0xad14)
	} else {
		de, a = 0xad2b, g.rb(0xad24)
	}
	hl := uint16(0x0f8d) + uint16(a*2)
	g.wb(de, g.rb(hl))
	g.wb(de+1, g.rb(hl+1))
}

// play15b5 is an empty sub-state.
func (g *Game) play15b5() {}
