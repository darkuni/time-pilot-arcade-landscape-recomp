package game

// ---- small routines shared by many states ---------------------------------

// nextState advances to the next top-level state (0x0f11).
func (g *Game) nextState() {
	g.wb(vState, g.rb(vState)+1)
	g.wb(vSub, 0)
}

// nextSub advances to the next sub-state (0x0f1a).
func (g *Game) nextSub() { g.wb(vSub, g.rb(vSub)+1) }

// hideSprites clears the attribute byte of 24 sprites (0x15b6).
func (g *Game) hideSprites() {
	for i := uint16(0); i < 0x18; i++ {
		g.wb(0xaa41+2*i, 0)
	}
}

// hideSprites4 clears the first four (0x0b2b).
func (g *Game) hideSprites4() {
	for i := uint16(0); i < 4; i++ {
		g.wb(0xaa41+2*i, 0)
	}
}

// titlePlanes places the four aircraft sprites of the title screen (0x0b06).
func (g *Game) titlePlanes() {
	d := uint8(0xa0)
	c := uint8(4)
	for i := uint16(0); i < 4; i++ {
		g.wb(0xaa10+2*i+0x31, d)
		g.wb(0xaa10+2*i, 0xd8)
		g.wb(0xaa10+2*i+1, c)
		g.wb(0xaa10+2*i+0x30, 0x6c)
		c++
		d -= 0x10
	}
}

// blinkText toggles the "PUSH START" style text every frame (0x0b39).
func (g *Game) blinkText() {
	if g.rb(vFrameCtr)&1 != 0 {
		g.enq(1, 0x00)
	} else {
		g.enq(1, 0x1f)
	}
}

// ---- screen wipe (0x01e1 / 0x0201 / 0x026f) --------------------------------

// wipeInit starts the spiral wipe (0x01e1).
func (g *Game) wipeInit() {
	g.wb(0xa9e2, 0)
	g.ww(0xa9e3, g.rw(0x0d45))
	g.ww(0xa9e5, g.rw(0x280c))
}

// wipePlot plots the current wipe cell with the wipe tile/colour and returns
// the video-RAM address it touched (0x026f).
func (g *Game) wipePlot() uint16 {
	l := g.rb(0xa9e4) * 8
	hl := uint16(l) * 4
	lo := uint8(hl) + g.rb(0xa9e6)
	hi := 0xa4 + uint8(hl>>8)
	addr := uint16(hi)<<8 | uint16(lo)
	g.wb(addr, g.rb(0xad0b))
	g.wb(addr&^0x0400, g.rb(0xad0c))
	return addr
}

func fixedDelta(target, cur uint16) uint16 {
	d := target - cur
	c := (d >> 12) & 1
	return uint16(uint8(0)-uint8(c))<<8 | uint16(uint8(d>>4))
}

// wipeStep advances the wipe by one waypoint; it reports true once the path
// table is exhausted (0x0201: "Z" result).
func (g *Game) wipeStep() bool {
	g.wipePlot()
	g.ww(0xa9e7, fixedDelta(g.rw(0x32f5), g.rw(0xa9e3)))
	g.ww(0xa9e9, fixedDelta(g.rw(0x0b45), g.rw(0xa9e5)))
	for {
		g.ww(0xa9e3, g.rw(0xa9e3)+g.rw(0xa9e7))
		g.ww(0xa9e5, g.rw(0xa9e5)+g.rw(0xa9e9))
		if g.wipePlot() == g.rw(0x14b2) {
			break
		}
	}
	n := g.rb(0xa9e2) + 1
	g.wb(0xa9e2, n)
	w := g.tbl16(0x0290, n)
	g.ww(0xa9e3, w<<8&0xff00)
	g.wb(0xa9e5, 0)
	g.wb(0xa9e6, uint8(w>>8))
	return uint8(w) == 0
}
