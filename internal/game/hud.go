package game

// ---- reserve / stage icons (queue command 6, routine 0x0dd7) -----------------
//
// A count (0..99) is drawn as a column of icons: thirties, tens, fives and
// ones, each with its own glyph, the remainder blanked out.

func (g *Game) iconSingle(de uint16, b, c uint8) uint16 { // 0x0e8d
	g.wb(de, b)
	g.wb(de-1, 0xf1)
	g.wb(colorAddr(de-1), c)
	g.wb(colorAddr(de), c)
	return de + 0x20
}

func (g *Game) iconFive(de uint16, b, c uint8) uint16 { // 0x0e9c
	g.wb(de, b+1)
	g.wb(de-1, b)
	g.wb(colorAddr(de-1), c)
	g.wb(colorAddr(de), c)
	return de + 0x20
}

func (g *Game) iconBig(de0 uint16, b, c uint8) uint16 { // 0x0e70: 2x2 icon
	g.wb(de0, b+1)
	de := de0 - 1
	g.wb(de, b)
	de += 0x20
	g.wb(de, b+2)
	de++
	g.wb(de, b+3)
	hl := de + 0xfc00 // colour RAM cell of the last tile
	de += 0x20
	g.wb(hl, c)
	hl--
	g.wb(hl, c)
	hl -= 0x20
	g.wb(hl, c)
	hl++
	g.wb(hl, c)
	return de
}

func (g *Game) drawIcons(a uint8) {
	de := uint16(0xa463)
	if a >= 0x64 {
		a = 0x63
	}
	n30 := a / 30
	a %= 30
	n10 := a / 10
	a %= 10
	n5 := a / 5
	n1 := a % 5
	for i := uint8(0); i < n1; i++ {
		de = g.iconSingle(de, 1, 0x13)
	}
	for i := uint8(0); i < n5; i++ {
		de = g.iconFive(de, 0x32, 0x11)
	}
	for i := uint8(0); i < n10; i++ {
		de = g.iconBig(de, 0xce, 0x16)
	}
	for i := uint8(0); i < n30; i++ {
		de = g.iconBig(de, 0x23, 0x11)
	}
	for {
		if uint32(0x59dd)+uint32(de) > 0xffff {
			return
		}
		de = g.iconSingle(de, 0xf1, 0x10)
	}
}

// drawStageNumber is queue command 7 (routine 0x0eac): the stage label and
// its two-digit number.
func (g *Game) drawStageNumber() {
	stage := g.rb(0xad01)
	if stage >= 0x64 {
		return
	}
	pos, _, p := g.textRecord(14)
	de := g.drawTextRun(pos, p, g.rb(0xad0c)&15)
	de += 0x40
	b := uint8(1)
	c := g.rb(0xad0c)
	tens, ones := stage/10, stage%10
	de = g.stageDigit(de, tens, &b, c) - 0x20
	g.stageDigit(de, ones, &b, c)
}

// stageDigit is routine 0x0eeb: a zero digit while b != 0 is suppressed.
func (g *Game) stageDigit(de uint16, a uint8, b *uint8, c uint8) uint16 {
	a &= 15
	if a == 0 {
		if *b != 0 {
			*b--
			return de + 0x20
		}
	} else {
		*b = 0
	}
	g.setTile(de, g.rom[0x0f06+uint16(a)], c)
	return de
}
