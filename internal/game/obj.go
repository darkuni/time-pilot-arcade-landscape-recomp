package game

// Object records.
//
// The original keeps every moving thing in a 16-byte record in work RAM
// (player at 0xa800, enemy planes from 0xa850, etc.) and addresses it with the
// Z80's IX register; the matching hardware-sprite shadow entry is addressed
// with IY (a pair of bytes in each of the shadow tables at 0xaa10/0xaa40).
// The rewrite keeps the same layout so that state can be compared with the
// reference byte for byte.  g.ix / g.iy play the part of the two registers.
//
// Record fields (offsets from IX):
//
//	+0   state: 0 free, 0xff alive, other values count an explosion down
//	+1   wanted heading
//	+2   current heading (0..255 = full circle)
//	+3,5 sub-pixel position fractions
//	+4,6 spare/pixel counters (bullets keep their tile position here)
//	+10..13 velocity words (bullets, bosses)
//	+14  timer
//	+15  slot number
//
// Shadow sprite fields (offsets from IY):
//
//	+0    X            +1   tile code
//	+2    X (2nd half) +3   tile code (2nd half)
//	+0x30 attribute    +0x32 attribute (2nd half)
//	+0x31 Y            +0x33 Y (2nd half)

func (g *Game) X(o uint16) uint8         { return g.rb(g.ix + o) }
func (g *Game) setX(o uint16, v uint8)   { g.wb(g.ix+o, v) }
func (g *Game) Y(o uint16) uint8         { return g.rb(g.iy + o) }
func (g *Game) setY(o uint16, v uint8)   { g.wb(g.iy+o, v) }
func (g *Game) incX(o uint16) uint8      { v := g.X(o) + 1; g.setX(o, v); return v }
func (g *Game) decX(o uint16) uint8      { v := g.X(o) - 1; g.setX(o, v); return v }
func (g *Game) xw(o uint16) uint16       { return g.rw(g.ix + o) }
func (g *Game) setXw(o uint16, v uint16) { g.ww(g.ix+o, v) }

// next advances IX/IY to the following record (0x309b).
func (g *Game) next() {
	g.ix += 0x10
	g.iy += 2
}

// at selects the object and sprite pair to work on.
func (g *Game) at(ix, iy uint16) { g.ix, g.iy = ix, iy }

// withObj runs f with IX/IY temporarily set to the given record.
func (g *Game) withObj(ix, iy uint16, f func()) {
	sx, sy := g.ix, g.iy
	g.ix, g.iy = ix, iy
	f()
	g.ix, g.iy = sx, sy
}

// ---- trigonometry (0x596e and friends) --------------------------------------
//
// A heading is a byte (256 steps per turn).  The ROM tables hold 256 signed
// 16-bit values; the entry at h is the horizontal component and the entry at
// h-64 the vertical one.

// vec returns the pair (T[h], T[h-64]) from the table at base.
func (g *Game) vec(base uint16, h uint8) (de, bc uint16) {
	de = g.rw(base + 2*uint16(h))
	bc = g.rw(base + 2*uint16(h-0x40))
	return
}

// vec2 is vec with both values doubled (0x599d / 0x59a0).
func (g *Game) vec2(base uint16, h uint8) (de, bc uint16) {
	de, bc = g.vec(base, h)
	return de << 1, bc << 1
}

// advance moves the current object along its heading using the table at base,
// compensating for the player's scrolling (0x58bc).  mult is 1 for the plain
// version and 2 for the double-speed version (0x58fe).
func (g *Game) advance(base uint16, mult uint16) {
	de, bc := g.vec(base, g.X(2))
	hl := g.rw(0xa808) + de*mult + (uint16(g.Y(0x31))<<8 | uint16(g.X(3)))
	g.setX(3, uint8(hl))
	g.setY(0x31, uint8(hl>>8))
	hl = g.rw(0xa80a) + bc*mult + (uint16(g.Y(0))<<8 | uint16(g.X(5)))
	g.setX(5, uint8(hl))
	g.setY(0, uint8(hl>>8))
}

// applyScroll adds the player's scroll velocity to the object's position
// (0x2b60).
func (g *Game) applyScroll() {
	hl := (uint16(g.Y(0x31))<<8 | uint16(g.X(3))) + g.rw(0xa808)
	g.setY(0x31, uint8(hl>>8))
	g.setX(3, uint8(hl))
	hl = (uint16(g.Y(0))<<8 | uint16(g.X(5))) + g.rw(0xa80a)
	g.setY(0, uint8(hl>>8))
	g.setX(5, uint8(hl))
}

// applyVelScroll is 0x3e05: position += own velocity (IX+10..13) + scroll.
func (g *Game) applyVelScroll() {
	hl := g.xw(10) + g.rw(0xa808) + (uint16(g.Y(0x31))<<8 | uint16(g.X(3)))
	g.setY(0x31, uint8(hl>>8))
	g.setX(3, uint8(hl))
	hl = g.xw(12) + g.rw(0xa80a) + (uint16(g.Y(0))<<8 | uint16(g.X(5)))
	g.setY(0, uint8(hl>>8))
	g.setX(5, uint8(hl))
}

// offScreen is 0x2b83: reports (carry) true when the object has strayed
// outside the visible playfield ... the test is "Y+9 < 3" or "X-3 < 3"
// inverted exactly as the original.
func (g *Game) offScreen() bool {
	if g.Y(0x31)+9 < 3 {
		return true
	}
	return g.Y(0)-3 < 3
}

// clearObj frees the object and its sprite (0x2bde).
func (g *Game) clearObj() {
	g.setX(0, 0)
	g.setX(3, 0)
	g.setX(5, 0)
	g.setY(0, 0)
	g.setY(0x31, 0)
}

// killObj is 0x40ab: free the record and hide its sprite.
func (g *Game) killObj() {
	g.setX(0, 0)
	g.setY(0, 0)
	g.setY(0x31, 0)
}

// aimLookup is 0x33b8: it returns an approximate heading (0..255) from the
// shadow sprite at IY toward the position pair stored at hl (X at hl, Y at
// hl-1).  The original uses a small 8-step division of the smaller offset by
// the larger one and a table of octant offsets.
func (g *Game) aimLookup(hl uint16) uint8 {
	c := uint8(0)
	b := g.Y(0x31)
	e := g.rb(hl)
	a := g.rb(hl&0xff00 | uint16(uint8(hl)-1)) // DEC L
	if a >= b {
		a -= b
	} else {
		a = b - a
		c |= 1
	}
	d := a
	g.aimD = d
	b = g.Y(0)
	a = e
	if a >= b {
		a -= b
	} else {
		a = b - a
		c |= 2
	}
	e = a
	g.aimE = e
	if e == d {
		return g.rb(0x341d + uint16(c))
	}
	if e < d {
		c |= 4
	}
	var h, den uint8
	if c&4 == 0 {
		h, den = d, e
	} else {
		h, den = e, d
	}
	r := uint16(h) << 8
	carry := false
	for i := 0; i < 8; i++ {
		v := uint32(r) << 1
		if carry {
			v |= 1
		}
		cout := v > 0xffff
		r = uint16(v)
		a := uint8(r >> 8)
		if cout || a >= den {
			a -= den
			r = uint16(a)<<8 | r&0xff
			carry = true
		} else {
			carry = false
		}
	}
	q := uint8(r)
	tab := g.rb(0x3415 + uint16(c))
	res := q >> 2 & 0x1f
	if tab&0x20 != 0 {
		res = 0x1f - res
	}
	return res + tab
}
