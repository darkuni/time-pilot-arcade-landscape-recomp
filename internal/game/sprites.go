package game

// Shadow sprite tables live in work RAM (0xaa10..0xaa6f).  Once per frame
// (at the top of the vblank handler) they are copied to the two hardware
// sprite RAMs, with a coordinate transform for the flipped (cocktail) case.

// sprSrc maps the n-th byte of a hardware sprite RAM (0..0x2f) to its shadow
// RAM address for table bank "first" (aa30 / aa60) and the others.
func sprSrc0(n int) uint16 { // feeds hardware RAM #1 (b010..b03f)
	switch {
	case n < 6:
		return 0xaa30 + uint16(n)
	case n < 0x26:
		return 0xaa10 + uint16(n-6)
	}
	return 0xaa36 + uint16(n-0x26)
}

func sprSrc1(n int) uint16 { // feeds hardware RAM #2 (b410..b43f)
	switch {
	case n < 6:
		return 0xaa60 + uint16(n)
	case n < 0x26:
		return 0xaa40 + uint16(n-6)
	}
	return 0xaa66 + uint16(n-0x26)
}

// flushSprites is routine 0x0365 (normal) / 0x0556 (flipped).
func (g *Game) flushSprites() {
	if g.rb(vFlip) != 0 {
		for n := 0; n < 0x30; n++ {
			g.wb(0xb010+uint16(n), g.rb(sprSrc0(n)))
			v := g.rb(sprSrc1(n))
			if n&1 == 1 {
				v = ^(v + 14)
			}
			g.wb(0xb410+uint16(n), v)
		}
	} else {
		for n := 0; n < 0x30; n++ {
			v := g.rb(sprSrc0(n))
			if n&1 == 0 {
				v = ^(v + 15)
			}
			g.wb(0xb010+uint16(n), v)
			w := g.rb(sprSrc1(n))
			if n&1 == 0 {
				w ^= 0xc0
			} else {
				w++
			}
			g.wb(0xb410+uint16(n), w)
		}
	}
	g.sprPark()
}

// sprPark is the tail at 0x04bc: during one phase of the title/attract state
// the last eight sprites are marked as "parked" (bit 7 of the attribute byte
// set, hardware Y shifted) so that they appear only after the beam passes.
func (g *Game) sprPark() {
	// The original hides these eight sprites until the raster beam has passed
	// them (a hardware sprite-RAM timing trick).  A rewrite draws every sprite
	// at its true position, so parking is a no-op.
	if true {
		return
	}
	if g.rb(vState) != 3 {
		return
	}
	a := g.rb(vSub)
	if a < g.k(0x0832) || a >= 8 {
		return
	}
	pairs := [8][2]uint16{
		{0xb411, 0xb010}, {0xb413, 0xb012}, {0xb415, 0xb014}, {0xb437, 0xb036},
		{0xb439, 0xb038}, {0xb43b, 0xb03a}, {0xb43d, 0xb03c}, {0xb43f, 0xb03e},
	}
	for _, p := range pairs {
		v := g.rb(p[0])
		if v >= 0x80 {
			continue
		}
		g.wb(p[0], v+0x80)
		g.wb(p[1], g.rb(p[1])+0x80)
	}
}
