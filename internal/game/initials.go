package game

// ---- entering initials for a new high score ------------------------------------

// play08b4 is sub-state 9: finish the screen wipe, then set up the entry.
func (g *Game) play08b4() {
	if !g.wipeStep() {
		return
	}
	g.enq(1, 0x13)
	g.enq(1, 0)
	g.enq(1, 0x14)
	g.enq(1, 0x15)
	g.enq(1, 0x0c)
	g.run4bdc()
	for i := uint16(0); i < 5; i++ {
		g.wb(0xa995+i, 0)
	}
	g.wb(0xa99a, 3)
	de := g.rw(0xa993)
	g.wb(de, g.rb(0x12c7+uint16(g.rb(0xa999))))
	g.wb(0xa990, g.rb(colorAddr(de)))
	g.nextSub()
}

// shiftIn is "RRCA / RL (HL)": push one input sample into a history byte.
func (g *Game) shiftIn(a uint16, bit uint8) {
	g.wb(a, g.rb(a)<<1|bit&1)
}

// play18c3 is sub-state 10: choose three initials with left/right and fire.
func (g *Game) play18c3() {
	if g.rb(vFrameCtr)&1 != 0 {
		g.wb(0xa99c, g.rb(0xa99c)+1)
		hl := colorAddr(g.rw(0xa993))
		if g.rb(0xa99c)&0x10 != 0 {
			g.wb(hl, 0x14)
		} else {
			g.wb(hl, 0x10)
		}
		g.initialsTail()
		return
	}
	in := g.playerInput()
	g.shiftIn(0xa995, in)
	g.shiftIn(0xa996, in>>1)
	g.shiftIn(0xa997, in>>4)
	g.shiftIn(0xa998, in>>5)
	switch {
	case g.rb(0xa998)&7 == 1 || g.rb(0xa997)&7 == 1:
		if g.acceptInitial() {
			return
		}
		g.showLetter()
	default:
		v := g.rb(0xa996)
		if v == 0xff {
			g.wb(0xa996, 0)
			v = 0
		}
		if v&7 == 1 {
			n := g.rb(0xa999) + 1
			if n >= 0x1b {
				n = 0
			}
			g.wb(0xa999, n)
			g.showLetter()
			break
		}
		v = g.rb(0xa995)
		if v == 0x7f {
			g.wb(0xa995, 0)
			v = 0
		}
		if v&7 == 1 {
			n := g.rb(0xa999) - 1
			if n >= 0x80 {
				n = 0x1a
			}
			g.wb(0xa999, n)
			g.showLetter()
			break
		}
	}
	// 0x1963: the time limit
	if g.rb(vFrameCtr)&7 == 0 {
		n := g.rb(vSubTimer) - 1
		g.wb(vSubTimer, n)
		if n == 0 {
			g.wb(g.rw(0xa993), 0xf1)
			g.finishInitials()
			return
		}
	}
	g.initialsTail()
}

// showLetter is 0x194e: draw the currently selected letter at the cursor.
func (g *Game) showLetter() {
	de := g.rw(0xa993)
	g.wb(de, g.rb(0x12c7+uint16(g.rb(0xa999))))
	g.wb(colorAddr(de), 0x10)
	g.wb(0xa99c, 0)
}

// acceptInitial is 0x1923: fix the letter and move on; true when all three
// letters have been entered.
func (g *Game) acceptInitial() bool {
	a := g.rb(0x12c7 + uint16(g.rb(0xa999)))
	hl := g.rw(0xa991)
	de := g.rw(0xa993)
	g.wb(de, a)
	g.wb(hl, a)
	g.wb(colorAddr(de), g.rb(0xa990))
	de -= 0x20
	hl++
	g.ww(0xa991, hl)
	g.ww(0xa993, de)
	n := g.rb(0xa99a) - 1
	g.wb(0xa99a, n)
	if n == 0 {
		g.finishInitials()
		return true
	}
	g.wb(0xa999, 0)
	return false
}

// finishInitials is 0x1975.
func (g *Game) finishInitials() {
	g.wb(vSubTimer, 0x3c)
	g.sndBurst7()
	g.nextSub()
}

// initialsTail is 0x199a: while both players are out, a start button may
// begin a new game even during name entry.
func (g *Game) initialsTail() {
	if g.rb(0xad10)|g.rb(0xad20) != 0 {
		return
	}
	st := g.rb(vInStart) & 0x18
	if g.rb(0xa9c0) != 0 {
		if st == 0 {
			return
		}
		g.hideSprites24()
		g.startFromAttract()
		return
	}
	cr := g.rb(vCredits)
	switch {
	case cr < 1:
		return
	case cr == 1:
		if st != 8 {
			return
		}
		g.hideSprites24()
		g.start1Player()
	default:
		if st == 0 {
			return
		}
		g.hideSprites24()
		if st == 8 {
			g.start1Player()
		} else {
			g.start2Players()
		}
	}
}
