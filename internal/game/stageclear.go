package game

// ---- the "stage clear" animation (sub-state 14, routine 0x1323) -----------------
//
// The playfield is wiped by a band of tiles that sweeps out of the middle of
// the screen, then the plane flies off.  The routine keeps its phase in a9f0
// and uses a9f1..a9f7 as scratch.

func (g *Game) play1323() {
	if g.rb(vFrameCtr)&2 != 0 {
		return
	}
	switch g.rb(0xa9f0) {
	case 0:
		g.clearTurn()
	case 1:
		g.clearTurn()
		g.clearOut()
	case 2:
		g.clearFlash()
		g.clearIn()
	case 3:
		g.clearIn()
	case 4:
		g.clearFill()
	default:
		g.wb(vSubTimer, 0x5a)
		g.hideSprites24()
		g.run4c75()
		g.wb(vSub, g.k(0x2750))
	}
}

// clearTurn is 0x1367: the plane banks.
func (g *Game) clearTurn() {
	n := g.rb(0xa9f1)
	if n == 8 {
		g.wb(0xa9f0, 1)
		g.sndPushGame(g.k(0x07a9))
	}
	b := uint8(0x3e)
	if g.rb(0xa9f1)&1 != 0 {
		b = 0
	}
	g.wb(0xaa40, g.rb(0xaa40)&0xc0+b)
	g.wb(0xa9f1, g.rb(0xa9f1)+1)
}

// clearFlash is 0x1393.
func (g *Game) clearFlash() {
	n := g.rb(0xa9f3)
	var a uint8
	switch {
	case n == 0:
		g.wb(0xa9f0, 3)
		a = 0x3f
	case n&4 == 0:
		a = 0x3f
	default:
		a = 0x37
	}
	g.wb(0xaa40, g.rb(0xaa40)&0xc0+a)
	g.wb(0xa9f3, g.rb(0xa9f3)-1)
}

// clearFill is 0x13cc: flood the colour RAM with the stage colour.
func (g *Game) clearFill() {
	g.wb(0xa9f0, 5)
	col := g.rb(0xad1c)
	if g.rb(vPlayer2) != 0 {
		col = g.rb(0xad2c)
	}
	if g.rb(vFlip) != 0 {
		hl := uint16(0xa044)
		for r := 0; r < 0x1c; r++ {
			for i := uint16(0); i < 0x1b; i++ {
				g.wb(hl+i, col)
			}
			hl += 0x20
		}
	} else {
		hl := uint16(0xa3be)
		for r := 0; r < 0x1c; r++ {
			for i := uint16(0); i < 0x1b; i++ {
				g.wb(hl-i, col)
			}
			hl -= 0x20
		}
	}
	g.wb(0xa9f6, g.rb(0xa9f6)-1)
}

// bandStep is 0x4a9d: shift one 13-cell strip of the band.
func (g *Game) bandStep(c uint8, de uint16) {
	hl := g.rw(0xa9f7)
	for b := 0; b < 13; b++ {
		if g.rb(hl) != 0 {
			v := g.rb(de) + 1
			if c&1 != 0 {
				v -= 2
			}
			g.wb(de, v)
		}
		if c&2 != 0 {
			de -= 0x20
		} else {
			de += 0x20
		}
		hl = g.rw(0xa9f7) + 1
		if c&1 != 0 {
			hl -= 2
		}
		g.ww(0xa9f7, hl)
	}
}

// copyBand is 0x1563 (save) / 0x158c (restore): move the top strip of the
// tile map to a scratch area and back.
func (g *Game) copyBand(save bool) {
	mv := func(a, b uint16) {
		if save {
			g.wb(b, g.rb(a))
		} else {
			g.wb(a, g.rb(b))
		}
	}
	de := uint16(0xa400)
	hl := uint16(0xa451)
	for i := 0; i < 0x1c; i++ {
		mv(de, hl)
		de++
		hl += 0x20
	}
	mv(de, 0xa5f0)
	mv(de+1, 0xa610)
	mv(de+2, 0xa5f2)
	mv(de+3, 0xa612)
}

func (g *Game) bump(a uint16, d int8) {
	g.wb(a, g.rb(a)+uint8(d))
	g.wb(a+0x20, g.rb(a+0x20)+uint8(d))
}

func (g *Game) clearBlank() {
	g.fillColumn(0xa7b1, 0xf1)
	g.fillColumn(0xa5d1, 0xf1)
	for _, a := range []uint16{0xa610, 0xa611, 0xa612} {
		g.wb(a, 0xf1)
		g.wb(a-0x20, 0xf1)
	}
}

// clearOut is 0x142a: the band opens out.
func (g *Game) clearOut() {
	odd := g.rb(0xa9f2)&1 != 0
	if odd {
		hl := g.rw(0xa9f7)
		if g.rb(hl) == 0xff {
			g.wb(0xa9f2, 0)
			g.wb(0xa9f0, 2)
			g.ww(0xa9f7, g.rw(0xa9f7)-1)
			return
		}
		g.copyBand(true)
		hl = g.rw(0xa9f7)
		f := g.rb(hl) & 1
		g.ww(0xa9f7, hl+1)
		if f != 0 {
			g.bump(0xa5f0, 1)
			g.bump(0xa5f2, 1)
		}
		hl = g.rw(0xa9f7)
		f = g.rb(hl) & 1
		g.ww(0xa9f7, hl+1)
		if f != 0 {
			g.bump(0xa5f1, 1)
		}
		g.bandStep(2, 0xa5d1)
		g.ww(0xa9f7, g.rw(0xa9f7)-13)
		g.bandStep(0, 0xa631)
		g.copyBand(false)
	} else {
		g.clearBlank()
	}
	g.wb(0xa9f2, g.rb(0xa9f2)-1)
}

// clearIn is 0x14c5: the band closes again.
func (g *Game) clearIn() {
	if g.rb(0xa9f4)&1 != 0 {
		hl := g.rw(0xa9f7)
		if g.rb(hl)&0xfe != 0 {
			g.wb(0xa9f4, 0)
			g.wb(0xa9f0, 4)
			g.sndPushLoud(g.k(0x27cb))
			g.sndPushLoud(g.k(0x33a0))
			g.ww(0xa9f7, g.rw(0xa9f7)+1)
			return
		}
		g.copyBand(true)
		g.bandStep(1, 0xa451)
		g.ww(0xa9f7, g.rw(0xa9f7)+13)
		g.bandStep(3, 0xa7b1)
		hl = g.rw(0xa9f7)
		f := g.rb(hl) & 1
		g.ww(0xa9f7, hl-1)
		if f != 0 {
			g.bump(0xa5f1, -1)
		}
		hl = g.rw(0xa9f7)
		f = g.rb(hl) & 1
		g.ww(0xa9f7, hl-1)
		if f != 0 {
			g.bump(0xa5f0, -1)
			g.bump(0xa5f2, -1)
		}
		g.copyBand(false)
	} else {
		g.clearBlank()
	}
	g.wb(0xa9f4, g.rb(0xa9f4)-1)
}
