package sound

import "math/bits"

// ---- music ------------------------------------------------------------------

// rec returns the address of music channel record r (10 bytes each).
func rec(r uint16) uint16 { return 0x300f + 10*r }

// musicStart sets up song n (0x037e, 0x0b16/0x0b1a): channel records are
// reset, each channel's stream pointer is loaded from the song header and the
// voice slots are re-assigned to the five music channels.
func (b *Board) musicStart(n uint8) {
	for r := uint16(0); r < 5; r++ {
		a := rec(r)
		for i := uint16(0); i < 10; i++ {
			b.wb(a+i, b.rom[0x3b5+10*r+i])
		}
	}
	hdr := 0xb6e + 10*uint16(n)
	for r := uint16(0); r < 5; r++ {
		b.wb(rec(r)+2, b.rom[hdr+2*r])
		b.wb(rec(r)+3, b.rom[hdr+2*r+1])
	}
	for i := uint16(0); i < 12; i++ {
		b.wb(0x3000+i, b.rom[0xb2f+i])
	}
}

// musicStep advances one music channel by one tick (0x01fc).
func (b *Board) musicStep(r uint16) uint8 {
	ix := rec(r)
	if b.dec(ix+1) != 0 {
		return 0
	}
	b.wb(ix+1, b.rb(ix+8))
	n := b.dec(ix)
	if n != 0 {
		if n&1 == 0 {
			return 0
		}
		slope := b.rb(ix + 9)
		if slope == 0 {
			return 0
		}
		v := slope + b.rb(ix+7)
		if v&0x80 != 0 {
			return 0
		}
		b.wb(ix+7, v)
		b.vol(v)
		return 0
	}
	// the current note has run out: read the next stream item
	for {
		hl := b.rw(ix + 2)
		d := b.rom[hl]
		switch d & 0x1f {
		case 0: // rest
			b.noteLength(ix, hl, d)
			return 0
		case 0x1f: // control item
			cmd := d >> 5
			arg := hl + 1
			b.ww(ix+2, hl+2)
			switch cmd {
			case 0: // select note table
				t := 0x2ce + uint16(b.rom[arg]<<1)
				b.ww(ix+4, uint16(b.rom[t])|uint16(b.rom[t+1])<<8)
			case 1: // tick period
				p := b.rom[0x366+uint16(b.rom[arg])]
				b.wb(ix+8, p)
				b.wb(ix+1, p)
			case 2:
				b.wb(ix+6, b.rom[arg])
			case 3:
				b.wb(ix+9, b.rom[arg])
			case 6: // jump
				b.wb(ix+2, b.rom[arg])
				b.wb(ix+3, b.rom[arg+1])
			default: // end of song
				return 0xff
			}
			continue
		}
		b.noteLength(ix, hl, d)
		idx := uint16(d&0x1f-1) * 2
		tab := b.rw(ix + 4)
		per := uint16(b.rom[tab+idx]) | uint16(b.rom[tab+idx+1])<<8
		b.tone(per)
		b.wb(ix+7, b.rb(ix+6))
		b.vol(b.rb(ix + 6))
		return 0
	}
}

// noteLength consumes the stream byte at hl and sets the note-length counter
// from the top three bits of d (0x0250): 1<<((n-1)&7).
func (b *Board) noteLength(ix, hl uint16, d uint8) {
	b.ww(ix+2, hl+1)
	b.wb(ix, bits.RotateLeft8(0x80, int(d>>5)))
}
