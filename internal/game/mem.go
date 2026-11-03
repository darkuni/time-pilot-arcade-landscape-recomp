// Package game is the native Go rewrite of Konami's Time Pilot game logic.
//
// The original program keeps all of its state in 2KB of work RAM plus the
// video, colour and sprite RAMs, and several of its algorithms depend on the
// exact byte layout (16-byte object records, shadow sprite tables, ring-buffer
// draw queue).  The rewrite therefore keeps the same state layout – exposed
// through typed accessors and named addresses – so that it can be verified
// byte-for-byte against the reference machine, and so that a later refactor
// into plain structs can be proven equivalent.
package game

// Memory map of the main board (only the parts the game logic touches).
const (
	addrColorRAM = 0xa000 // 0x400 bytes, one colour byte per tile
	addrVideoRAM = 0xa400 // 0x400 bytes, one tile code per tile
	addrWorkRAM  = 0xa800 // 0x800 bytes
	addrSprRAM   = 0xb000 // 0x100 bytes (hardware sprite RAM #1)
	addrSprRAM2  = 0xb400 // 0x100 bytes (hardware sprite RAM #2)
)

// Game holds the complete game state.
type Game struct {
	rom [0x6000]byte // program ROM: used only for constant tables

	ram  [0x800]byte
	vram [0x400]byte
	cram [0x400]byte
	spr  [0x100]byte
	spr2 [0x100]byte

	// Raw input ports, active low, exactly as the hardware presents them.
	// In0: coin/start, In1: P1 controls, In2: P2 controls, Dsw0/Dsw1: DIPs.
	In0, In1, In2, Dsw0, Dsw1 uint8

	// Outputs to the sound board / video hardware.
	SoundLatch   uint8 // last byte written to the sound command port
	SoundIRQ     int   // count of sound-CPU interrupt requests raised
	FlipScreen   bool  // cocktail flip flag as written to the hardware latch
	NMIEnabled   bool
	VPos         uint8 // current raster line, as read back by the program
	soundIRQLast bool
	rtab         map[uint16]func()
	aimD, aimE   uint8  // |dy| and |dx| left behind by aimLookup (D and E)
	ix, iy       uint16 // current object record and its sprite pair
}

// New creates a game from the main program ROM image (0x6000 bytes).
func New(mainROM []byte) *Game {
	g := &Game{}
	copy(g.rom[:], mainROM)
	g.In0, g.In1, g.In2, g.Dsw0, g.Dsw1 = 0xff, 0xff, 0xff, 0xff, 0x4b
	return g
}

// ---- raw memory access --------------------------------------------------

func (g *Game) rb(a uint16) uint8 {
	switch {
	case a < 0x6000:
		return g.rom[a]
	case a >= 0xa000 && a < 0xa400:
		return g.cram[a-0xa000]
	case a >= 0xa400 && a < 0xa800:
		return g.vram[a-0xa400]
	case a >= 0xa800 && a < 0xb000:
		return g.ram[a-0xa800]
	case a >= 0xb000 && a < 0xc000:
		if a&0x400 == 0 {
			return g.spr[a&0xff]
		}
		return g.spr2[a&0xff]
	case a >= 0xc000:
		return g.ioRead(a)
	}
	return 0xff
}

func (g *Game) wb(a uint16, v uint8) {
	switch {
	case a >= 0xa000 && a < 0xa400:
		g.cram[a-0xa000] = v
	case a >= 0xa400 && a < 0xa800:
		g.vram[a-0xa400] = v
	case a >= 0xa800 && a < 0xb000:
		g.ram[a-0xa800] = v
	case a >= 0xb000 && a < 0xc000:
		if a&0x400 == 0 {
			g.spr[a&0xff] = v
		} else {
			g.spr2[a&0xff] = v
		}
	case a >= 0xc000:
		g.ioWrite(a, v)
	}
}

func (g *Game) rw(a uint16) uint16 { return uint16(g.rb(a)) | uint16(g.rb(a+1))<<8 }
func (g *Game) ww(a uint16, v uint16) {
	g.wb(a, uint8(v))
	g.wb(a+1, uint8(v>>8))
}

func (g *Game) ioRead(a uint16) uint8 {
	switch {
	case a&0x0f00 == 0x000:
		return g.VPos
	case a&0x0f00 == 0x200:
		return g.Dsw1
	case a&0x0fe0 == 0x300:
		return g.In0
	case a&0x0fe0 == 0x320:
		return g.In1
	case a&0x0fe0 == 0x340:
		return g.In2
	case a&0x0fe0 == 0x360:
		return g.Dsw0
	}
	return 0xff
}

func (g *Game) ioWrite(a uint16, v uint8) {
	switch {
	case a&0x0f00 == 0x000:
		g.SoundLatch = v
	case a&0x0f00 == 0x200: // watchdog
	case a&0x0f00 == 0x300: // addressable latch: bit number in A1..A3, data in D0
		n := (a >> 1) & 7
		b := v&1 != 0
		switch n {
		case 0:
			g.NMIEnabled = b
		case 1:
			g.FlipScreen = b
		case 2:
			if !g.soundIRQLast && b {
				g.SoundIRQ++
			}
			g.soundIRQLast = b
		}
	}
}

// ---- helpers shared by every routine ----------------------------------------

// tbl16 returns the little-endian word stored at base+2*idx (routine 0x018c).
func (g *Game) tbl16(base uint16, idx uint8) uint16 {
	return g.rw(base + 2*uint16(idx))
}

// bcdAdd8 returns a+b+carry in packed BCD plus the carry out (Z80 ADD/ADC + DAA).
func bcdAdd8(a, b uint8, carry bool) (uint8, bool) {
	c := uint16(0)
	if carry {
		c = 1
	}
	sum := uint16(a) + uint16(b) + c
	r := uint8(sum)
	cf := sum > 0xff
	hf := (a&0x0f)+(b&0x0f)+uint8(c) > 0x0f
	adj := uint8(0)
	if hf || r&0x0f > 9 {
		adj |= 0x06
	}
	if cf || r > 0x99 {
		adj |= 0x60
		cf = true
	}
	return r + adj, cf
}

// bcdSub8 returns a-b in packed BCD (Z80 SUB followed by DAA) and the borrow.
func bcdSub8(a, b uint8) (uint8, bool) {
	r := a - b
	cf := b > a
	hf := (a & 0x0f) < (b & 0x0f)
	diff := uint8(0)
	if hf || r&0x0f > 9 {
		diff |= 0x06
	}
	if cf || r > 0x99 {
		diff |= 0x60
		cf = true
	}
	return r - diff, cf
}

// LoadState overwrites the game's RAM images (used by the verification
// harness to start from a known reference state).
func (g *Game) LoadState(work, video, color, spr1, spr2 []byte) {
	copy(g.ram[:], work)
	copy(g.vram[:], video)
	copy(g.cram[:], color)
	copy(g.spr[:], spr1)
	copy(g.spr2[:], spr2)
}

// Work, Video, Color expose the RAM images for inspection and rendering.
func (g *Game) Work() []byte     { return g.ram[:] }
func (g *Game) Video() []byte    { return g.vram[:] }
func (g *Game) Color() []byte    { return g.cram[:] }
func (g *Game) Sprites1() []byte { return g.spr[:] }
func (g *Game) Sprites2() []byte { return g.spr2[:] }
