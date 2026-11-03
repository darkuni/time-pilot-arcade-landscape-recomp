package game

// Frame runs one complete video frame of game logic: the vertical-blank
// interrupt handler (everything the original does in its NMI at 0x00d8) and
// then the foreground work that follows it (draining the draw-command queue).
func (g *Game) Frame() {
	g.flushSprites()
	g.tileFixups()

	g.NMIEnabled = false
	// Normal orientation, except when player 2 plays on a cocktail table.
	g.wb(vFlip, 1)
	if g.rb(vPlayer2) != 0 && g.rb(0xa9c2) == 0 {
		g.wb(vFlip, 0)
	}
	g.FlipScreen = g.rb(vFlip)&1 != 0

	// Latch the (inverted) input ports into work RAM.
	g.wb(vInCoin, ^g.Dsw1)
	g.wb(vInStart, ^g.In0)
	g.wb(vInP1, ^g.In1)
	g.wb(vInP2, ^g.In2)
	g.wb(vInDsw0, ^g.Dsw0)

	g.wb(vFrameCtr, g.rb(vFrameCtr)+1)
	g.wb(vNmiTick, bcdInc(g.rb(vNmiTick)))
	for _, a := range []uint16{0xa817, 0xa812, 0xa8f4} {
		if v := g.rb(a); v != 0 {
			g.wb(a, v-1)
		}
	}

	g.coinHandling()

	switch g.rb(vState) & 3 {
	case 0:
		g.stateBoot()
	case 1:
		g.stateAttract()
	case 2:
		g.stateReady()
	case 3:
		g.statePlay()
	}

	g.soundPump()
	g.NMIEnabled = g.k(0x1600)&1 != 0
	g.drainQueue()
}

func bcdInc(v uint8) uint8 {
	r, _ := bcdAdd8(v, 1, false)
	return r
}

// ---- tile fix-up lists (0x5286) --------------------------------------------
//
// Two small lists at 0xae00 and 0xae80 record screen cells that were
// temporarily altered; the "undo" list from the previous frame is applied
// first and the current list is then promoted to be the next undo list.

func (g *Game) tileFixups() {
	g.undoFixups()
	g.applyFixups()
	l := g.rb(0xae00)
	if l == 4 {
		g.resetFixups()
		return
	}
	for i := uint16(0); i < uint16(l); i++ {
		g.wb(0xae80+i, g.rb(0xae00+i))
	}
	g.wb(0xae80, l+0x80)
	g.ww(0xae00, 0xae04)
}

func (g *Game) resetFixups() {
	g.ww(0xae80, 0xae84)
	g.ww(0xae00, 0xae04)
}

// applyFixups is routine 0x52d2.
func (g *Game) applyFixups() {
	c := g.rb(0xad0c) & 15
	l := uint8(g.rw(0xae00))
	if l == 4 {
		return
	}
	n := (l - 4) >> 2 & 0x1f
	hl := uint16(0xae04)
	for ; n > 0; n-- {
		de := g.rw(hl)
		hl += 2
		if g.rb(de)&0x10 != 0 {
			hl += 2
			continue
		}
		g.wb(de|0x0400, g.rb(hl))
		hl++
		g.wb(de&^0x0400, g.rb(hl)+c)
		hl++
	}
}

// undoFixups is routine 0x530e.
func (g *Game) undoFixups() {
	l := uint8(g.rw(0xae80)) & 0x7f
	if l == 4 {
		return
	}
	n := (l - 4) >> 2 & 0x1f
	hl := uint16(0xae84)
	for ; n > 0; n-- {
		de := g.rw(hl)
		hl += 2
		if g.rb(de)&0x10 != 0 {
			hl += 2
			continue
		}
		hl += 2
		g.wb(de|0x0400, 0x20)
	}
}

// ---- sound command queue ---------------------------------------------------

// sndPush appends a command to the sound queue (routine 0x5628).
func (g *Game) sndPush(v uint8) {
	n := g.rb(vSndQ) + 1
	g.wb(vSndQ, n)
	g.wb(vSndQ+uint16(n), v)
}

// sndPushGame pushes only while a game is running (0x560c).
func (g *Game) sndPushGame(v uint8) {
	if g.rb(vGameActive) != 0 {
		g.sndPush(v)
	}
}

// sndPushLoud pushes while a game runs or attract-mode sound is enabled
// (0x5617).
func (g *Game) sndPushLoud(v uint8) {
	if g.rb(vGameActive) != 0 || g.rb(0xa9c6) != 0 {
		g.sndPush(v)
	}
}

// soundPump sends at most one queued command to the sound board per frame
// (routine 0x55d4).
func (g *Game) soundPump() {
	n := g.rb(vSndQ)
	if n == 0 {
		return
	}
	g.wb(vSndQ, n-1)
	g.sendSound(g.rb(vSndQ + 1))
	if n-1 == 0 {
		return
	}
	for i := uint16(0); i < uint16(n-1); i++ {
		g.wb(vSndQ+1+i, g.rb(vSndQ+2+i))
	}
}

// sendSound writes a command byte to the sound latch and pulses the sound
// CPU interrupt (0x55f8).
func (g *Game) sendSound(v uint8) {
	g.wb(0xc000, v)
	g.wb(0xc304, 1)
	g.wb(0xc304, 0)
}
