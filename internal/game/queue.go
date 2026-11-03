package game

import "fmt"

// ---- draw command queue --------------------------------------------------
//
// The original game never touches video RAM from its main logic for "slow"
// operations; instead it posts (command,argument) byte pairs into a ring
// buffer (RST 0x38) which a foreground loop drains (0x0b93).  The rewrite
// keeps that two-stage design because the order of VRAM writes matters.

// enq posts a draw command (RST 0x38).  The pair is dropped if the slot at
// the write cursor is still occupied.
func (g *Game) enq(d, e uint8) {
	l := g.rb(vQWr)
	a := uint16(0xac00) | uint16(l)
	if g.rb(a)&0x80 == 0 {
		return
	}
	g.wb(a, d)
	g.wb(a+1, e)
	g.wb(vQWr, (l+2)&0x3f)
}

// drainQueue executes every pending queued command (the foreground loop).
func (g *Game) drainQueue() {
	for {
		l := g.rb(vQRd)
		a := uint16(0xac00) | uint16(l)
		if g.rb(a)&0x80 != 0 {
			return
		}
		c := g.rb(a)
		g.wb(a, 0xff)
		b := g.rb(a + 1)
		g.wb(a+1, 0xff)
		g.wb(vQRd, (l+2)&0x3f)
		g.queueCmd(c&15, b)
	}
}

func (g *Game) queueCmd(cmd, arg uint8) {
	switch cmd {
	case 0:
		g.drawTextNoColor(arg)
	case 1:
		pos, c, p := g.textRecord(arg)
		g.drawTextRun(pos, p, c)
	case 2:
		pos, _, p := g.textRecord(arg)
		g.drawTextRun(pos, p, g.rb(0xad0c)&15)
	case 3:
		g.eraseText(arg)
	case 4:
		g.addScore(arg)
	case 5:
		g.drawLives(arg)
	case 6:
		g.drawIcons(arg)
	case 7:
		g.drawStageNumber()
	case 10:
		pos, _, p := g.textRecord(arg)
		g.drawTextRun(pos, p, (g.rb(0xad0c)+5)&15)
	case 11:
		pos, _, p := g.textRecord(arg)
		g.drawTextRun(pos, p, (g.rb(0xad0c)+10)&15)
	}
}

// todo marks a routine that has not been ported yet.
func (g *Game) todo(addr uint16) {
	panic(fmt.Sprintf("unported routine %04x", addr))
}

// textRecord decodes a record of the text table at 0x0c50:
// word position, colour byte, tile codes ..., 0xb9 terminator.
func (g *Game) textRecord(idx uint8) (pos uint16, color uint8, chars uint16) {
	rec := g.tbl16(0x0c50, idx)
	return g.rw(rec), g.rb(rec + 2), rec + 3
}

// colorAddr maps a video-RAM address to the matching colour-RAM address
// (RES 2,D).
func colorAddr(a uint16) uint16 { return a &^ 0x0400 }

// setTile writes a tile code and colour at video address a (0xa400..).
func (g *Game) setTile(a uint16, tile, color uint8) {
	g.wb(a, tile)
	g.wb(a&^0x0400, color)
}

// drawTextRun draws tile codes starting at p until the 0xb9 terminator, one
// cell per step, advancing down the tile RAM by 0x20.
func (g *Game) drawTextRun(pos, p uint16, color uint8) uint16 {
	for {
		a := g.rb(p)
		if a == 0xb9 {
			return pos
		}
		g.setTile(pos, a, color)
		p++
		pos -= 0x20
	}
}

// drawTextNoColor is queue command 0: tile codes only.
func (g *Game) drawTextNoColor(idx uint8) {
	pos, _, p := g.textRecord(idx)
	for {
		a := g.rb(p)
		if a == 0xb9 {
			return
		}
		g.wb(pos, a)
		p++
		pos -= 0x20
	}
}

// eraseText is queue command 3: blank the cells of a text record.
func (g *Game) eraseText(idx uint8) {
	pos, _, p := g.textRecord(idx)
	for g.rb(p) != 0xb9 {
		g.wb(pos, 0xf1)
		p++
		pos -= 0x20
	}
}

// ---- score handling ---------------------------------------------------------

// addScore is queue command 4: add table entry idx to the current player's
// BCD score, update the high score, and redraw.
func (g *Game) addScore(idx uint8) {
	if g.rb(vGameActive) == 0 {
		return
	}
	if idx == 0 {
		g.refreshScoreLabels()
		return
	}
	src := 0x0d27 + 3*uint16(idx)
	dst := uint16(vScoreP1)
	if g.rb(vPlayer2) != 0 {
		dst = vScoreP2
	}
	var carry bool
	var v uint8
	for i := uint16(0); i < 3; i++ {
		v, carry = bcdAdd8(g.rb(dst+i), g.rb(src+i), carry)
		g.wb(dst+i, v)
	}
	// compare the 3-byte score (MSB at dst+2) with the high score (MSB at a98d)
	de := dst + 2
	hl := uint16(vHiScore + 2)
	c := 3
	newHi := false
	for c > 0 {
		a, h := g.rb(de), g.rb(hl)
		if a < h {
			break
		}
		if a != h {
			newHi = true
			break
		}
		de--
		hl--
		c--
	}
	if newHi {
		// LDDR: copy c bytes downward from the player's score to the hi score
		for i := 0; i < c; i++ {
			g.wb(hl-uint16(i), g.rb(de-uint16(i)))
		}
		g.drawHiScore()
	}
	if g.rb(vPlayer2) == 0 {
		g.drawP1Score()
	} else {
		g.drawP2Score()
	}
}

// refreshScoreLabels handles the "add 0 points" request: it (re)draws the
// score panel for the current number of players (0x0ce9).
func (g *Game) refreshScoreLabels() {
	if g.rb(0xad31) == 0 {
		g.drawTextIdx(g.k(0x0b31))
		g.drawP1Score()
		g.eraseText(g.k(0x15c6))
		de := uint16(0xa501)
		for i := 0; i < 6; i++ {
			g.wb(de, 0xf1)
			de -= 0x20
		}
		return
	}
	g.drawTextIdx(6)
	g.drawP1Score()
	g.drawTextIdx(7)
	g.drawP2Score()
}

// drawTextIdx is routine 0x0bf2: draw text record idx with its own colour.
func (g *Game) drawTextIdx(idx uint8) {
	pos, c, p := g.textRecord(idx)
	g.drawTextRun(pos, p, c)
}

// k returns a constant byte stored in the program ROM.  The original code
// loads many of its tuning constants from ROM addresses; keeping the lookup
// preserves behaviour exactly (a later pass can replace these with literals).
func (g *Game) k(a uint16) uint8 { return g.rom[a] }

// ---- number drawing -------------------------------------------------------

func (g *Game) drawP1Score() { g.drawNumber(0xa781, 0xad35, 0x10) }
func (g *Game) drawP2Score() { g.drawNumber(0xa501, 0xad38, 0x10) }
func (g *Game) drawHiScore() { g.drawNumber(0xa641, 0xa98d, 0x10) }
func (g *Game) drawCredits() { g.drawByte(0xa47f, 0xa986, 0x10) }

// drawNumber draws a 6-digit BCD number (three bytes, most significant at hl,
// stored downwards) with leading-zero suppression, at screen address de.
func (g *Game) drawNumber(de, hl uint16, c uint8) {
	b := uint8(0)
	de = g.drawByteSuppressed(de, hl, c, &b)
	hl--
	de = g.drawByteSuppressed(de, hl, c, &b)
	hl--
	g.drawByte(de, hl, c)
}

// drawByteSuppressed is routine 0x0da0.
func (g *Game) drawByteSuppressed(de, hl uint16, c uint8, b *uint8) uint16 {
	v := g.rb(hl)
	de = g.drawDigitSuppressed(de, c, v>>4, b)
	de -= 0x20
	de = g.drawDigitSuppressed(de, c, v&15, b)
	return de - 0x20
}

// drawDigitSuppressed is routine 0x0daf: a zero digit before the first
// non-zero digit is drawn as the blank glyph.
func (g *Game) drawDigitSuppressed(de uint16, c, d uint8, b *uint8) uint16 {
	a := d & 15
	if a != 0 {
		*b++
	} else {
		a = g.k(0x3246)
		if *b != 0 {
			a = 0
		}
	}
	g.setTile(de, g.rom[0x0dcc+uint16(a)], c)
	return de
}

// drawByte is routine 0x0d81: two digits, no suppression.
func (g *Game) drawByte(de, hl uint16, c uint8) uint16 {
	v := g.rb(hl)
	g.setTile(de, g.rom[0x0dcc+uint16(v>>4)], c)
	de -= 0x20
	g.setTile(de, g.rom[0x0dcc+uint16(v&15)], c)
	return de - 0x20
}
