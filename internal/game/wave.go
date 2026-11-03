package game

// ---- random numbers, aiming and enemy wave spawning ---------------------------

// rand is routine 0x4b4b: a 17-byte shift register mixed with the frame
// counter.
func (g *Game) rand() uint8 {
	for i := uint16(0); i < 16; i++ {
		g.wb(0xab40-i, g.rb(0xab3f-i))
	}
	a := g.rb(0xab37) ^ g.rb(0xab40)
	g.wb(0xab30, a)
	return a + g.rb(vFrameCtr)
}

// seqStep is 0x323a: step the object's flight-plan counter and fetch the
// next movement code from the plan table.
func (g *Game) seqStep() {
	t := g.X(9)
	if t == 0 {
		return
	}
	t--
	g.setX(9, t)
	plan := g.rw(0x3438 + 2*uint16(g.X(10)))
	g.setX(8, g.rb(plan+uint16(t)))
}

// chooseAim is routine 0x31b4: once per frame one enemy plane (picked by
// the low bits of ad05) is pointed at the player or at a fixed spot.
func (g *Game) chooseAim() {
	c := g.rb(0xad05)
	hi := c & 0xf0
	if hi != 0 && hi != 0x30 {
		if c&15 == 7 {
			g.aimRing()
		}
		return
	}
	n := c & 15
	if n >= 7 {
		return
	}
	g.at(0xa850+uint16(n)*0x10, 0xaa1a+uint16(n)*2)
	if g.X(0) != 0xff {
		return
	}
	g.seqStep()
	m := g.X(8)
	switch m {
	case 0x10:
		return
	case 0x11:
		g.setX(1, g.aimLookup(0xac65)+0x80)
		g.setX(8, 0x10)
		g.setX(9, 0)
	default:
		g.setX(1, g.aimLookup(0xac65+uint16(m)*2))
	}
}

// aimRing is 0x326c: fill the table of eight target points arranged round
// the player (used by the squadron that circles you).
func (g *Game) aimRing() {
	const base = 0xac64
	put := func(o uint16, v uint8) { g.wb(base+o, v) }
	de, bc := g.vec2(0x5e00, g.rb(0xa802)+0x40)
	hl := de << 3
	h := uint8(hl >> 8)
	put(0x10, h+0x78)
	put(0x14, 0x78-h)
	hl <<= 1
	h = uint8(hl >> 8)
	put(0x12, h+0x78)
	put(0x16, 0x78-h)
	hl = bc << 3
	h = uint8(hl >> 8)
	put(0x11, h+0x84)
	put(0x15, 0x84-h)
	hl <<= 1
	h = uint8(hl >> 8)
	put(0x13, h+0x84)
	put(0x17, 0x84-h)
	de, bc = g.vec2(0x5e00, g.rb(0xa802))
	hl = de << 3
	h = uint8(hl >> 8)
	put(0x18, h+0x78)
	hl <<= 1
	put(0x1a, uint8(hl>>8)+0x78)
	hl = bc << 3
	h = uint8(hl >> 8)
	put(0x19, h+0x84)
	hl <<= 1
	put(0x1b, uint8(hl>>8)+0x84)
}

// pickKind is 0x382d: choose the enemy type for a new plane.
func (g *Game) pickKind() uint8 {
	r := g.rand()
	if r >= g.rb(0xacc4) {
		return r&3 + 5
	}
	v := g.rb(0xa9cf) + 1
	if v >= 5 {
		v = 0
	}
	g.wb(0xa9cf, v)
	return v
}

// spawner is routine 0x36af: bring new enemy planes into the sky.
func (g *Game) spawner() {
	if g.rb(0xacc6) != 0 {
		return
	}
	if g.rb(vStage) == 4 {
		g.spawnStage4()
		return
	}
	a := g.rb(0xad06) & 15
	switch {
	case a == 7:
		g.spawnReform()
	case a < 7:
		g.spawnSingles(false)
	case a < 9:
		g.spawnSingles(true)
	default:
		g.spawnSquadron()
	}
}

// spawnSquadron (0x36d1) launches a group of planes that arrive together.
func (g *Game) spawnSquadron() {
	if g.rb(0xad05) != 0 {
		return
	}
	r := g.rand()
	g.wb(0xacc2, 0xff)
	g.wb(0xacc3, g.rb(vStage)<<1|r&1)
	c := g.rb(0x38d9 + uint16((g.rb(0xa802)+8)>>4&15))
	de := 0x397b + uint16(g.rb(0xacc3)<<4)
	b := g.rb(0xacc1)
	if g.rb(0xad02) == 0 {
		b = 5
	}
	g.wb(0xa811, 0)
	g.at(0xa850, 0xaa1a)
	for ; b > 0; b-- {
		if g.X(0) == 0 {
			idx := (g.rb(de) + c) << 1
			hl := 0x38e9 + uint16(idx)
			g.setY(0x31, g.rb(hl))
			g.setY(0, g.rb(hl+1))
			h := g.rb(0xa802) + 0x80
			g.setX(1, h)
			g.setX(2, h)
			g.setX(10, g.pickKind()+9)
			de++
			g.setX(14, g.rb(de))
			de++
			g.setX(3, 0)
			g.setX(5, 0)
			g.setX(9, 0x20)
			g.seqStep()
			g.setX(0, 0xfe)
			if g.X(14) == 0 {
				g.incX(0)
			}
			g.wb(0xa811, g.rb(0xa811)+1)
		}
		g.next()
	}
	g.wb(0xacc2, 0)
	g.wb(0xa812, 0xe4)
	n := g.rb(0xa811)
	if n >= 5 {
		g.sndPushGame(g.k(0x273a))
		return
	}
	want := g.rb(0xacc1)
	g.wb(0xa811, want)
	if n >= want {
		g.sndPushGame(g.k(0x273a))
	}
}

// spawnSingles (0x379f / 0x37bd): trickle in lone planes when few remain.
func (g *Game) spawnSingles(countCheck bool) {
	a := g.rb(0xad05)
	if countCheck {
		if a != 0 && a != 0x30 {
			return
		}
		cnt := 0
		for i := uint16(0); i < 7; i++ {
			if g.rb(0xa850+i*0x10) != 0 {
				cnt++
			}
		}
		if cnt >= 2 {
			return
		}
	} else if a != 0 && a != 0x30 {
		return
	}
	var b uint8
	var ix, iy uint16
	if g.rb(0xad02) == 0 {
		b, ix, iy = 5, 0xa890, 0xaa22
	} else {
		b, ix, iy = g.rb(0xacc1), 0xa8b0, 0xaa26
	}
	g.at(ix, iy)
	for {
		if g.X(0) == 0 {
			g.setX(0, 0xff)
			c := g.rb(0xa802) >> 2 & 0x3f
			r := (g.rand()&15 - 8 + c) & 0x3f
			a := g.rb(0x39fb+uint16(r)) << 2
			hl := 0x3a3b + uint16(a)
			g.setY(0x31, g.rb(hl))
			g.setY(0, g.rb(hl+1))
			h := g.rb(0xa802) + 0x80
			g.setX(1, h)
			g.setX(2, h)
			g.setX(10, g.pickKind())
			g.wb(0xacc5, 0)
			g.setX(3, 0)
			g.setX(5, 0)
			g.setX(9, 0x20)
			g.seqStep()
			g.setX(14, 0)
			return
		}
		g.ix -= 0x10
		g.iy -= 2
		b--
		if b == 0 {
			return
		}
	}
}

// spawnReform (0x3855) sends the planes into the circling formation.
func (g *Game) spawnReform() {
	if g.rb(0xad05) != 0 {
		return
	}
	for i := uint16(0); i < 5; i++ {
		g.wb(0xa850+i*0x10+8, 0x11)
		g.wb(0xa850+i*0x10+9, 0)
	}
}

// spawnStage4 (0x386e): the last stage launches every free plane at once.
func (g *Game) spawnStage4() {
	g.at(0xa850, 0xaa1a)
	b := g.rb(0xacc1)
	if g.rb(0xad0d) != 0 {
		b = 5
	}
	for b0 := b; b0 > 0; b0-- {
		if g.X(0) == 0 {
			r := g.rand() & 0xfc
			hl := 0x3a3b + uint16(r)
			g.setY(0x31, g.rb(hl))
			g.setY(0, g.rb(hl+1))
			h := g.rb(hl + 2)
			g.setX(1, h)
			g.setX(2, h)
			g.setX(10, g.rb(0x38d2+uint16(g.rb(0xacc1)-b0)))
			g.setX(9, 0x20)
			g.seqStep()
			g.setX(4, 1)
			g.setX(14, 0)
			g.decX(0)
		}
		g.next()
	}
	g.wb(0xa812, 0xe4)
}
