package game

// g43e8 is the tail shared by the title screens (0x43e8): a ROM checksum gate.
func (g *Game) g43e8(hl uint16, n int) {
	var a uint8
	for i := 0; i < n; i++ {
		a += g.rom[hl+uint16(i)]
	}
	if a != 0x67 {
		panic("tamper path 0f8d")
	}
	g.nextSub()
}

// run4bdc draws the five-line high-score table (0x4bdc).
func (g *Game) run4bdc() {
	g.hiScoreLine(0xab08, 0xa711, 0x14)
	g.hiScoreLine(0xab10, 0xa713, 0x16)
	g.hiScoreLine(0xab18, 0xa715, 0x12)
	g.hiScoreLine(0xab20, 0xa717, 0x15)
	g.hiScoreLine(0xab28, 0xa719, 0x13)
}

// hiScoreLine draws one 8-byte table record: [rank, score(3, LSB first),
// initials(3), pad] as rank glyphs, six score digits, then the initials.
func (g *Game) hiScoreLine(rec, de uint16, c uint8) {
	r := g.rb(rec)
	glyph := 0x4cb4 + 3*uint16(r)
	for i := uint16(0); i < 3; i++ {
		g.setTile(de, g.rb(glyph+i), c)
		if i < 2 {
			de -= 0x20
		}
	}
	de += 0xff80
	b := uint8(0)
	hl := rec + 3
	de = g.drawByteSuppressed(de, hl, c, &b)
	hl--
	de = g.drawByteSuppressed(de, hl, c, &b)
	hl--
	de = g.drawByte(de, hl, c)
	de += 0xffa0
	for i := uint16(0); i < 3; i++ {
		g.setTile(de, g.rb(rec+4+i), c)
		if i < 2 {
			de -= 0x20
		}
	}
}

// run4c75 is play sub-state 3: load the active player's record and show the
// reserve count.
func (g *Game) run4c75() {
	g.run07d2()
	src := uint16(0xad10)
	if g.rb(vPlayer2) != 0 {
		src = 0xad20
	}
	for i := uint16(0); i < 0x10; i++ {
		g.wb(0xad00+i, g.rb(src+i))
	}
	if g.rb(vGameActive) == 0 {
		g.nextSub()
		return
	}
	g.enq(6, g.rb(0xad01))
	g.enq(5, g.rb(0xad00)-1)
	var x uint8
	for i := uint16(0); i < 256; i++ {
		x ^= g.rom[0x5b50+i]
	}
	g.wb(0xc308, x+0xff)
	g.nextSub()
}

// run07d2 blanks a 14-cell column of the playfield border (0x07d2).
func (g *Game) run07d2() {
	hl := uint16(0xa79f)
	for i := 0; i < 14; i++ {
		g.wb(hl, 0xf1)
		g.wb(hl&^0x0400, 0x16)
		hl -= 0x20
	}
}

// scoreRank is routine 0x4d2b: compares the 3-byte score at de (MSB first,
// walking down) with the table entry at hl; true when score < entry.
func (g *Game) scoreLess(de, hl uint16) bool {
	for c := 0; c < 3; c++ {
		a, h := g.rb(de), g.rb(hl)
		if a < h {
			return true
		}
		if a != h {
			return false
		}
		de--
		hl--
	}
	return false
}

// run4cc3 inserts the finished player's score into the high-score table
// (0x4cc3).  It returns true when the score did not qualify (carry set).
func (g *Game) run4cc3() bool {
	hl := uint16(0xab0b)
	de := uint16(0xad35)
	if g.rb(vPlayer2) != 0 {
		de = 0xad38
	}
	b := 5
	for ; b > 0; b-- {
		if !g.scoreLess(de, hl) {
			break
		}
		hl += 8
	}
	if b == 0 {
		return true
	}
	b--
	var slot uint16
	if b == 0 {
		slot = 0xab2f
	} else {
		n := uint16(b) * 8
		from, to := uint16(0xab27), uint16(0xab2f)
		for i := uint16(0); i < n; i++ {
			g.wb(to-i, g.rb(from-i))
		}
		slot = to - n
	}
	hl = slot
	for i := 0; i < 3; i++ {
		hl--
		g.wb(hl, 0xf1)
	}
	g.ww(0xa991, hl)
	hl--
	// copy the 3 score bytes (LDDR) from the player's score to the slot
	src := de
	dst := hl
	for i := uint16(0); i < 3; i++ {
		g.wb(dst-i, g.rb(src-i))
	}
	rank := g.rb(dst - 3)
	g.ww(0xa993, 0xa531+uint16(rank*2))
	hl = 0xab08
	for a := uint8(0); a < 5; a++ {
		g.wb(hl, a)
		hl += 8
	}
	return false
}
