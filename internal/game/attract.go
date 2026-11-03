package game

// ---- state 0: boot -----------------------------------------------------------

// clearColumnInit prepares the screen-clearing sweep (0x019a).
func (g *Game) clearColumnInit(start uint16, count uint8) {
	g.ww(0xa989, start)
	g.wb(0xa988, count)
}

// clearColumn blanks one 32-cell column and moves to the next; it reports
// true when the sweep has finished (0x01c2).
func (g *Game) clearColumn() bool {
	hl := g.rw(0xa989)
	for i := 0; i < 0x20; i++ {
		g.wb(hl, 0xf1)
		g.wb(hl&^0x0400, 0x10)
		hl += 0x20
	}
	g.ww(0xa989, g.rw(0xa989)+1)
	n := g.rb(0xa988) - 1
	g.wb(0xa988, n)
	return n == 0
}

// boot15e2: start the full-screen clear and select the next step.
func (g *Game) boot15e2() {
	g.clearColumnInit(0xa400, 0x20)
	g.wb(vSub, g.k(0x1749))
	g.wb(vState, 0) // the original folds a ROM checksum into this; it is 0 on a good ROM
}

// boot15fe: clear the screen one column per frame, then draw the title furniture
// and enter the attract sequence.
func (g *Game) boot15fe() {
	if !g.clearColumn() {
		return
	}
	g.enq(1, 5)
	g.enq(1, 6)
	g.enq(1, 7)
	g.enq(6, 1)
	g.wb(0xa701, 0x13)
	g.wb(0xa6e1, 0x13)
	// signature cells (value, then 5 in the following byte)
	rec := uint16(0x163f)
	for i := 0; i < 6; i++ {
		a := g.rw(rec)
		g.wb(a, g.rb(rec+2))
		g.wb(a+1, 5)
		rec += 3
	}
	g.drawHiScore()
	g.wb(vState, 1)
	g.wb(vSub, 2)
	if g.rb(0xa9c0) != 0 {
		g.enq(1, 0x0d)
	}
}

// ---- state 1: attract -----------------------------------------------------

func (g *Game) att1734() {
	if !g.wipeStep() {
		return
	}
	var s uint8
	for i := uint16(0); i < 0x22; i++ {
		s -= g.rb(0x1748 + i)
	}
	g.wb(0xa817, s)
	g.nextSub()
}

func (g *Game) att2d3f() {
	if g.rb(0xa9c0) != 0 {
		g.nextSub()
		return
	}
	g.drawCredits()
	g.enq(1, 8)
	g.titlePlanes()
	g.blinkText()
	g.g43e8(0x086b, 0x14)
}

func (g *Game) att083e() {
	g.blinkText()
	g.titlePlanes()
	for e := uint8(0); e < 2; e++ {
		g.enq(1, e)
	}
	for e := uint8(3); e < 8; e++ {
		g.enq(1, e)
	}
	g.enq(1, 0x14)
	g.enq(1, 0x15)
	g.nextSub()
}

// saveCell copies a screen cell's tile and colour into RAM (0x1afc).
func (g *Game) saveCell(vram, dst uint16) {
	g.wb(dst, g.rb(vram))
	g.wb(dst+1, g.rb(colorAddr(vram)))
}

func (g *Game) att1748() {
	g.titlePlanes()
	g.blinkText()
	n := g.rb(vSubTimer) - 1
	g.wb(vSubTimer, n)
	if n != 0 {
		return
	}
	g.saveCell(0xa63c, 0xacc7)
	g.enq(3, 3)
	g.enq(3, 4)
	g.nextSub()
}

func (g *Game) att176a() {
	g.enq(1, 0x13)
	g.run4bdc()
	g.saveCell(0xa5dc, 0xadfb)
	g.nextSub()
}

func (g *Game) att178c() {
	g.titlePlanes()
	g.blinkText()
	n := g.rb(vSubTimer) - 1
	g.wb(vSubTimer, n)
	if n != 0 {
		return
	}
	g.saveCell(0xa67c, 0xab43)
	g.nextSub()
}

func (g *Game) att17b9() {
	c := g.k(0x590d)
	_ = c
	a := g.k(0x4a40)
	for i := uint16(0); i < 0x33; i++ {
		a += g.rom[0x0b06+i]
	}
	if a == 0xef {
		g.nextSub()
		return
	}
	g.wb(0xc308, g.k(0x4c89))
	g.saveCell(0xa65c, 0xad39)
}

func (g *Game) att3252() {
	var e uint8
	for i := uint16(0); i < 0x300; i++ {
		e ^= g.rom[0x0008+i]
	}
	if e+0x52 != 0 {
		g.nextState()
		return
	}
	g.nextSub()
}

func (g *Game) att17e2() {
	g.wb(0xaa3f, 0xff)
	a := g.k(0x27c0)
	for i := uint16(0); i < 0x1e; i++ {
		a += g.rom[0x335e+i]
	}
	g.wb(0xaa6f, a)
	g.nextSub()
}

func (g *Game) att4b19() {
	c := uint8(0x89)
	for i := uint16(0); i < 256; i++ {
		c += g.rom[0x0bcc+i]
	}
	if c != g.k(0x1a50) {
		g.nextState()
		return
	}
	g.nextSub()
}

func (g *Game) att2730() {
	// Starts the demonstration game.
	g.hideSprites4()
	g.run210e()
	g.wb(0xad31, 0)
	g.wb(0xad20, 0)
	g.wb(vGameActive, 0)
	g.wb(vSub, 0)
	g.wb(0xad10, 1)
	g.wb(vState, 3)
}

// ---- state 2: ready -------------------------------------------------------

func (g *Game) rdy181e() {
	g.hideSprites()
	g.saveCell(0xa5fc, 0xacbe)
	g.clearColumnInit(0xa404, g.k(0x0ccd))
	g.nextSub()
}

func (g *Game) rdy2cdb() {
	if !g.clearColumn() {
		return
	}
	g.nextSub()
}

func (g *Game) rdy1830() {
	g.titlePlanes()
	g.blinkText()
	g.enq(1, 1)
	g.enq(1, 0x14)
	g.enq(1, 0x15)
	e := uint8(15)
	if g.rb(0xa9c3) != 0 {
		e += 2
	}
	g.enq(1, e)
	g.enq(1, e+1)
	g.enq(1, 0x16)
	g.enq(1, 0)
	if g.rb(vCredits) < 2 {
		g.enq(1, 0x17)
		g.nextSub()
		return
	}
	g.enq(1, 0x19)
	g.nextSub()
	g.nextSub()
}

func (g *Game) rdy07e6() {
	g.titlePlanes()
	g.blinkText()
	g.saveCell(0xa61c, 0xabfe)
	if g.rb(vInStart)&0x08 != 0 {
		g.start1Player()
		return
	}
	if g.rb(vCredits)-1 == 0 {
		return
	}
	g.enq(1, 0x19)
	g.nextSub()
}

func (g *Game) rdy188a() {
	g.titlePlanes()
	g.blinkText()
	a := g.rb(vInStart)
	if a&0x10 != 0 {
		g.start2Players()
		return
	}
	if a&0x08 != 0 {
		g.start1Player()
	}
}

// start2Players is 0x189e.
func (g *Game) start2Players() {
	g.hideSprites4()
	g.wb(vGameActive, 0xff)
	g.wb(0xad31, 0xff)
	g.wb(0xad10, g.rb(0xa9c1))
	g.wb(0xad20, g.rb(0xa9c1))
	g.run460e()
	v, _ := bcdSub8(g.rb(vCredits), 2)
	g.wb(vCredits, v)
	g.drawCredits()
	g.enterPlay()
}

// start1Player is 0x3215.
func (g *Game) start1Player() {
	g.hideSprites4()
	g.wb(0xad31, 0)
	g.wb(0xad20, 0)
	g.wb(vGameActive, 0xff)
	g.wb(0xad10, g.rb(0xa9c1))
	v, _ := bcdSub8(g.rb(vCredits), 1)
	g.wb(vCredits, v)
	g.drawCredits()
	g.run4b30()
	g.enterPlay()
}

// run4b30 is 0x4b30: save three screen cells to RAM.
func (g *Game) run4b30() {
	hl := uint16(0x0d1b)
	for i := 0; i < 3; i++ {
		src := g.rw(hl)
		dst := g.rw(hl + 2)
		hl += 4
		colour := g.rb(src)
		tile := g.rb(src + 0x400)
		g.wb(dst, tile)
		g.wb(dst+1, colour)
	}
}

// att074b is the first attract sub-state after a game: start the wipe that
// clears the playfield for the title screens.
func (g *Game) att074b() {
	same := g.rb(0xad0c) == 5
	g.wb(0xad0c, 5)
	g.wb(0xad0b, 0xf1)
	g.wipeInit()
	if same {
		g.nextSub()
	}
	g.nextSub()
}
