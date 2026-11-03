// Package machine is the DEV-ONLY reference system (original ROMs on a Z80
// interpreter) used to verify the native rewrite. Not shipped.
package machine

import (
	"timepilot/internal/assets"
	"timepilot/tools/oracle/z80"
)

const (
	MainClock  = 3072000.0
	SoundClock = 1789772.727
	LinesTotal = 264
	CyclesLine = 192 // 3.072MHz / (6.144MHz/384)
)

// AYWrite is one write to an AY-3-8910 register.
type AYWrite struct {
	Cycle uint64 // sound CPU cycle
	Chip  int
	Reg   uint8
	Val   uint8
}

type Machine struct {
	ROM *assets.ROMs

	Main  *z80.CPU
	Sound *z80.CPU
	mb    *mainBus
	sb    *soundBus

	// Inputs (active-low bytes as read by the CPU)
	In0, In1, In2, Dsw0, Dsw1 uint8

	Line     int
	Frame    int
	soundAcc float64

	// Observability
	AYLog      []AYWrite
	LogAY      bool
	MainTrace  func(pc uint16)
	SoundTrace func(pc uint16)
	WriteWatch func(addr uint16, v uint8, pc uint16)
	ReadWatch  func(addr uint16, v uint8, pc uint16)
	IOLog      []IOEvent

	// Coverage: bit0 = instruction start executed, bit1 = read as data
	Cov   [0x6000]byte
	CovOn bool
	RAMWr [0x800]uint32
	RAMRd [0x800]uint32
}

type IOEvent struct {
	Frame int
	Line  int
	PC    uint16
	Addr  uint16
	Val   uint8
	Write bool
}

type mainBus struct {
	m          *Machine
	rom        []byte
	colorRAM   [0x400]byte
	videoRAM   [0x400]byte
	ram        [0x800]byte
	spr        [0x100]byte // b000-b0ff (spriteram)
	spr2       [0x100]byte // b400-b4ff (spriteram2)
	soundLatch uint8
	nmiEnable  bool
	flip       [2]bool
	sndIRQLast uint8
	latch      [16]uint8
	pc         uint16
}

type soundBus struct {
	m       *Machine
	rom     []byte
	ram     [0x400]byte
	ayAddr  [2]uint8
	ayReg   [2][16]uint8
	FilterW []uint16
}

func New(r *assets.ROMs) *Machine {
	m := &Machine{ROM: r}
	m.mb = &mainBus{m: m, rom: r.Main[:]}
	m.sb = &soundBus{m: m, rom: r.Sound[:]}
	m.Main = &z80.CPU{Bus: m.mb}
	m.Sound = &z80.CPU{Bus: m.sb}
	m.Reset()
	// default inputs: all released, factory-ish DIPs (see notes)
	m.In0, m.In1, m.In2 = 0xff, 0xff, 0xff
	m.Dsw0, m.Dsw1 = 0xff, 0x4b
	return m
}

func (m *Machine) Reset() {
	m.Main.Reset()
	m.Sound.Reset()
	m.Line = 0
}

func (b *mainBus) Read(a uint16) uint8 {
	switch {
	case a < 0x6000:
		if b.m.CovOn && b.m.Frame > 240 && (a < b.pc || a >= b.pc+4) {
			b.m.Cov[a] |= 2
		}
		return b.rom[a]
	case a >= 0xa000 && a < 0xa400:
		return b.colorRAM[a-0xa000]
	case a >= 0xa400 && a < 0xa800:
		return b.videoRAM[a-0xa400]
	case a >= 0xa800 && a < 0xb000:
		if b.m.CovOn {
			b.m.RAMRd[a-0xa800]++
		}
		return b.ram[a-0xa800]
	case a >= 0xb000 && a < 0xc000:
		if a&0x400 == 0 {
			return b.spr[a&0xff]
		}
		return b.spr2[a&0xff]
	case a >= 0xc000:
		return b.ioRead(a)
	}
	return 0xff
}

func (b *mainBus) ioRead(a uint16) uint8 {
	m := b.m
	var v uint8 = 0xff
	switch a & 0xf3ff & ^uint16(0x0c00) { // mirror
	}
	switch {
	case a&0x0f00 == 0x000: // c000
		v = uint8(m.vpos())
	case a&0x0f00 == 0x200: // c200
		v = m.Dsw1
	case a&0x0fe0 == 0x300:
		v = m.In0
	case a&0x0fe0 == 0x320:
		v = m.In1
	case a&0x0fe0 == 0x340:
		v = m.In2
	case a&0x0fe0 == 0x360:
		v = m.Dsw0
	}
	return v
}

func (m *Machine) vpos() int { return m.Line }

func (b *mainBus) Write(a uint16, v uint8) {
	if b.m.WriteWatch != nil {
		b.m.WriteWatch(a, v, b.pc)
	}
	switch {
	case a >= 0xa000 && a < 0xa400:
		b.colorRAM[a-0xa000] = v
	case a >= 0xa400 && a < 0xa800:
		b.videoRAM[a-0xa400] = v
	case a >= 0xa800 && a < 0xb000:
		if b.m.CovOn {
			b.m.RAMWr[a-0xa800]++
		}
		b.ram[a-0xa800] = v
	case a >= 0xb000 && a < 0xc000:
		if a&0x400 == 0 {
			b.spr[a&0xff] = v
		} else {
			b.spr2[a&0xff] = v
		}
	case a >= 0xc000:
		b.ioWrite(a, v)
	}
}

func (b *mainBus) ioWrite(a uint16, v uint8) {
	m := b.m
	switch {
	case a&0x0f00 == 0x000: // c000: sound latch
		b.soundLatch = v
	case a&0x0f00 == 0x200: // c200: watchdog
	case a&0x0f00 == 0x300: // LS259 addressable latch, bit = A1..A3, data = D0
		n := (a >> 1) & 7
		bitv := v & 1
		b.latch[n] = bitv
		switch n {
		case 0:
			b.nmiEnable = bitv != 0
		case 1:
			b.flip[0] = bitv != 0
		case 2:
			// sound IRQ trigger on rising edge
			if b.sndIRQLast == 0 && bitv == 1 {
				m.Sound.IRQ(0xff)
			}
			b.sndIRQLast = bitv
		}
	}
}

func (b *mainBus) In(p uint16) uint8     { return 0xff }
func (b *mainBus) Out(p uint16, v uint8) {}

// ---------------- sound bus ----------------

func (b *soundBus) Read(a uint16) uint8 {
	switch {
	case a < 0x1000:
		return b.rom[a]
	case a >= 0x3000 && a < 0x4000:
		return b.ram[a&0x3ff]
	case a >= 0x4000 && a < 0x5000:
		return b.ayRead(0)
	case a >= 0x6000 && a < 0x7000:
		return b.ayRead(1)
	}
	return 0xff
}

var timerTab = [10]uint8{0x00, 0x10, 0x20, 0x30, 0x40, 0x90, 0xa0, 0xb0, 0xa0, 0xd0}

func (b *soundBus) ayRead(chip int) uint8 {
	r := b.ayAddr[chip] & 0xf
	if chip == 0 {
		switch r {
		case 14:
			return b.m.mb.soundLatch
		case 15:
			return timerTab[(b.m.Sound.Cycles/512)%10]
		}
	}
	return b.ayReg[chip][r]
}

func (b *soundBus) Write(a uint16, v uint8) {
	switch {
	case a >= 0x3000 && a < 0x4000:
		b.ram[a&0x3ff] = v
	case a >= 0x4000 && a < 0x5000:
		b.ayData(0, v)
	case a >= 0x5000 && a < 0x6000:
		b.ayAddr[0] = v
	case a >= 0x6000 && a < 0x7000:
		b.ayData(1, v)
	case a >= 0x7000 && a < 0x8000:
		b.ayAddr[1] = v
	case a >= 0x8000 && a < 0x9000:
		b.FilterW = append(b.FilterW, a&0xfff)
		if len(b.FilterW) > 4096 {
			b.FilterW = b.FilterW[1:]
		}
	}
}

func (b *soundBus) ayData(chip int, v uint8) {
	r := b.ayAddr[chip] & 0xf
	b.ayReg[chip][r] = v
	if b.m.LogAY {
		b.m.AYLog = append(b.m.AYLog, AYWrite{b.m.Sound.Cycles, chip, r, v})
	}
}

func (b *soundBus) In(p uint16) uint8     { return 0xff }
func (b *soundBus) Out(p uint16, v uint8) {}

// ---------------- stepping ----------------

// RunLine executes one scanline of both CPUs.
func (m *Machine) RunLine() {
	// vblank NMI at the start of line 240
	if m.Line == 240 && m.mb.nmiEnable {
		m.Main.NMI()
	}
	target := m.Main.Cycles + CyclesLine
	for m.Main.Cycles < target {
		m.mb.pc = m.Main.PC
		if m.CovOn && m.Main.PC < 0x6000 {
			m.Cov[m.Main.PC] |= 1
		}
		if m.MainTrace != nil {
			m.MainTrace(m.Main.PC)
		}
		m.Main.Step()
	}
	m.soundAcc += CyclesLine * SoundClock / MainClock
	st := m.Sound.Cycles + uint64(m.soundAcc)
	m.soundAcc -= float64(uint64(m.soundAcc))
	for m.Sound.Cycles < st {
		m.Sound.Step()
	}
	m.Line++
	if m.Line == LinesTotal {
		m.Line = 0
		m.Frame++
	}
}

// RunFrame runs until the next frame boundary.
func (m *Machine) RunFrame() {
	f := m.Frame
	for m.Frame == f {
		m.RunLine()
	}
}

// Accessors for verification.
func (m *Machine) WorkRAM() []byte    { return m.mb.ram[:] }
func (m *Machine) VideoRAM() []byte   { return m.mb.videoRAM[:] }
func (m *Machine) ColorRAM() []byte   { return m.mb.colorRAM[:] }
func (m *Machine) SpriteRAM() []byte  { return m.mb.spr[:] }
func (m *Machine) SpriteRAM2() []byte { return m.mb.spr2[:] }
func (m *Machine) SoundRAM() []byte   { return m.sb.ram[:] }
func (m *Machine) FlipScreen() bool   { return m.mb.flip[0] }

// SoundCommand delivers a command byte to the sound board exactly as the
// main CPU does (latch write plus interrupt request).
func (m *Machine) SoundCommand(v uint8) {
	m.mb.soundLatch = v
	m.Sound.IRQ(0xff)
}

// SoundRun steps only the sound CPU for at least n cycles.
func (m *Machine) SoundRun(n uint64) {
	end := m.Sound.Cycles + n
	for m.Sound.Cycles < end {
		if m.SoundTrace != nil {
			m.SoundTrace(m.Sound.PC)
		}
		m.Sound.Step()
	}
}
