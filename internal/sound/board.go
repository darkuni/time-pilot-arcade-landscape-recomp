// Package sound is the native rewrite of the Time Pilot sound board: the
// command sequencer that the original ran on a second Z80, plus a synthesiser
// for the two AY-3-8910 chips and the board's per-voice RC filters.
//
// The sequencer keeps the original's work-RAM layout (addresses 0x3000-0x33ff)
// so that its behaviour can be compared byte for byte with the reference.
package sound

// TickCycles is the interval between sequencer passes, in sound-CPU clock
// cycles (the board's timer chip releases the main loop once every 5120
// cycles of the 1.789772 MHz clock).
const TickCycles = 5120

// ClockHz is the sound board clock (14.31818 MHz / 8).
const ClockHz = 1789772.727

// Board is the sound sequencer state.
type Board struct {
	rom [0x1000]byte // sound program ROM, used for note / effect tables
	ram [0x400]byte  // 0x3000-0x33ff

	// AY-3-8910 register files (two chips).
	Reg [2][16]uint8
	// Filter holds the last value written to the filter-select latch (the
	// low 12 bits: two bits per voice choosing the RC filter capacitor).
	Filter uint16

	pending []uint8 // commands waiting to be taken by the sequencer

	// OnAY, when set, is told about every AY register write (testing).
	OnAY func(chip int, reg, val uint8)
	// OnFilter, when set, is told about every filter latch write.
	OnFilter func(addr uint16)
}

// New creates a board and runs the program's start-up code.
func New(rom []byte) *Board {
	b := &Board{}
	copy(b.rom[:], rom)
	b.reset()
	return b
}

func (b *Board) rb(a uint16) uint8    { return b.ram[a&0x3ff] }
func (b *Board) wb(a uint16, v uint8) { b.ram[a&0x3ff] = v }
func (b *Board) rw(a uint16) uint16   { return uint16(b.rb(a)) | uint16(b.rb(a+1))<<8 }
func (b *Board) ww(a uint16, v uint16) {
	b.wb(a, uint8(v))
	b.wb(a+1, uint8(v>>8))
}
func (b *Board) dec(a uint16) uint8 { v := b.rb(a) - 1; b.wb(a, v); return v }
func (b *Board) inc(a uint16) uint8 { v := b.rb(a) + 1; b.wb(a, v); return v }

// RAM exposes the sequencer RAM image (testing).
func (b *Board) RAM() []byte { return b.ram[:] }

// ayw writes a register of one of the AY chips.
func (b *Board) ayw(chip int, reg, val uint8) {
	b.Reg[chip][reg&15] = val
	if b.OnAY != nil {
		b.OnAY(chip, reg, val)
	}
}

// ---- start-up (0x0000..0x00be) -------------------------------------------

func (b *Board) reset() {
	b.ram = [0x400]byte{}
	b.Reg = [2][16]uint8{}
	b.ww(0x300c, 0x8000)
	b.filterWrite(0x8000)
	for d := uint8(6); d >= 1; d-- {
		b.volv(d, 0)
	}
	b.ayw(0, 7, 0x38)
	b.ayw(1, 7, 0x38)
}

func (b *Board) filterWrite(addr uint16) {
	b.Filter = addr & 0xfff
	if b.OnFilter != nil {
		b.OnFilter(addr)
	}
}

// ---- primitives shared by every effect ------------------------------------

// vol sets the volume of the voice being processed (0x0199).
func (b *Board) vol(c uint8) { b.volv(b.rb(0x300e), c) }

// volv sets the volume of voice v (1..6) (0x019c).
func (b *Board) volv(v, c uint8) {
	if v < 4 {
		b.ayw(0, v+7, c)
	} else {
		b.ayw(1, v+4, c)
	}
}

// tone sets the tone period of the voice being processed (0x01aa).
func (b *Board) tone(hl uint16) {
	v := b.rb(0x300e)
	if v < 4 {
		r := (v - 1) * 2
		b.ayw(0, r, uint8(hl))
		b.ayw(0, r+1, uint8(hl>>8))
		return
	}
	r := (v - 4) * 2
	b.ayw(1, r, uint8(hl))
	b.ayw(1, r+1, uint8(hl>>8))
}

// mix selects the output filter mode (0..3) of the voice being processed and
// updates the board's filter latch (0x01c8).
func (b *Board) mix(mode uint8) {
	v := b.rb(0x300e)
	var shift uint
	switch v {
	case 1:
		shift = 6
	case 2:
		shift = 8
	case 3:
		shift = 10
	case 4:
		shift = 0
	case 5:
		shift = 2
	case 6:
		shift = 4
	}
	// the two mode bits are stored bit-reversed by the original code
	m := (mode&1)<<1 | (mode>>1)&1
	mask := ^(uint16(3) << shift)
	hl := b.rw(0x300c)
	hl = hl&mask | uint16(m)<<shift
	b.ww(0x300c, hl)
	b.filterWrite(hl)
}

// ---- command interface (the interrupt handler, 0x0038..0x0099) -------------

// Command delivers a command byte from the main board.
func (b *Board) Command(v uint8) { b.pending = append(b.pending, v) }

func (b *Board) takeCommand(cmd uint8) {
	if cmd == 0 {
		for i := uint16(0); i < 12; i++ {
			b.wb(0x3000+i, 0)
		}
		return
	}
	id := cmd & 0x7f
	if id >= 0x21 {
		return
	}
	if cmd&0x80 != 0 { // stop
		if s := b.findSlot(id); s >= 0 {
			b.wb(0x3000+uint16(s)*2, 0)
			b.wb(0x3001+uint16(s)*2, 0)
		}
		return
	}
	if s := b.findSlot(id); s >= 0 { // already running: restart it
		b.wb(0x3001+uint16(s)*2, 0)
		return
	}
	if s := b.findSlot(0); s >= 0 { // free slot
		b.wb(0x3000+uint16(s)*2, cmd)
		b.wb(0x3001+uint16(s)*2, 0)
		return
	}
	// all busy: replace the lowest-numbered effect if the new one outranks it
	min := b.rb(0x3000)
	for i := uint16(1); i < 6; i++ {
		if v := b.rb(0x3000 + 2*i); v < min {
			min = v
		}
	}
	if min >= cmd {
		return
	}
	s := b.findSlot(min)
	b.wb(0x3000+uint16(s)*2, cmd)
	b.wb(0x3001+uint16(s)*2, 0)
}

// findSlot returns the first slot holding id, or -1 (0x0080).
func (b *Board) findSlot(id uint8) int {
	for i := 0; i < 6; i++ {
		if b.rb(0x3000+uint16(i)*2) == id {
			return i
		}
	}
	return -1
}

// ---- one pass of the main loop (0x00bf..0x0166) ----------------------------

// Tick runs one sequencer pass: pending commands are taken first (the
// original's interrupt handler runs while it waits for the timer), then each
// of the six voices is serviced in turn.
func (b *Board) Tick() {
	for _, c := range b.pending {
		b.takeCommand(c)
	}
	b.pending = b.pending[:0]
	for v := uint8(1); v <= 6; v++ {
		b.wb(0x300e, v)
		s := 0x3000 + uint16(v-1)*2
		id, state := b.rb(s), b.rb(s+1)
		if state == 0 {
			// first pass for this slot: run the effect's start routine
			r, abort := b.start(id)
			if abort {
				return
			}
			if r != 0 {
				b.wb(s, 0)
				b.wb(s+1, 0)
			} else {
				b.wb(s+1, 1)
			}
			continue
		}
		if id == 0 {
			continue
		}
		if b.step(id) != 0 {
			b.wb(s, 0)
			b.wb(s+1, 0)
		}
	}
}

// NewTrace is New with an AY-write observer installed before start-up, so
// the start-up writes are reported too (testing).
func NewTrace(rom []byte, f func(chip int, reg, val uint8)) *Board {
	b := &Board{OnAY: f}
	copy(b.rom[:], rom)
	b.reset()
	return b
}
