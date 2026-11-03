package sound

// start runs the start routine of effect id (table at 0x0a92).  It reports
// the routine's result (non-zero means "finished already") and whether the
// routine abandoned the current pass (music start does).
func (b *Board) start(id uint8) (a uint8, abort bool) {
	switch id {
	case 0:
		b.vol(0) // 0x0197: silence the voice
	case 1:
		b.mix(0)
		b.tone(0x6b)
		b.wb(0x3041, 0)
		b.wb(0x3042, 0x47)
		b.wb(0x3043, 9)
		b.vol(9)
	case 2:
		b.mix(0)
		b.vol(10)
		b.wb(0x3044, 0x80)
		b.ww(0x3045, 4)
		b.tone(4)
	case 3:
		b.mix(0)
		b.ww(0x3047, 0xc0)
		b.tone(0xc0)
		b.vol(9)
		b.wb(0x3049, 0)
	case 4:
		b.mix(1)
		b.wb(0x304a, 0)
		b.wb(0x304b, 15)
		b.vol(15)
		b.ww(0x304c, 0x60)
	case 5:
		b.mix(0)
		b.wb(0x304e, 0xb0)
		b.wb(0x304f, 13)
		b.vol(b.dec(0x304f))
		b.wb(0x3050, 0x20)
		b.tone(0x20)
	case 6:
		b.mix(0)
		b.tone(0x50)
		b.reinit6()
	case 7:
		b.mix(0)
		b.wb(0x3056, 8)
		b.wb(0x3055, 0x2c)
		b.step7b()
	case 8:
		b.mix(1)
		b.wb(0x305c, 12)
		b.vol(12)
		b.wb(0x305b, 0)
		b.ww(0x3059, 0x80)
		b.tone(0x80)
	case 9:
		b.mix(1)
		b.wb(0x305d, 0)
		b.wb(0x305e, 0xe0)
		b.reinit9(0x305e)
	case 10:
		b.mix(0)
		b.wb(0x3062, 0)
		b.wb(0x3063, 0xc0)
		b.wb(0x3064, 13)
		b.restart10()
	case 11:
		b.mix(0)
		b.vol(12)
		b.wb(0x3067, 0xb8)
		b.wb(0x3068, 0x1f)
		b.tone(0x1f)
	case 12:
		b.mix(1)
		b.tone(0xc0)
		b.vol(15)
		b.wb(0x3069, 0)
		b.wb(0x306a, 0x53)
		b.wb(0x306b, 9)
		b.wb(0x306c, 0x37)
	case 13:
		b.mix(1)
		b.restart13()
	case 14:
		b.mix(1)
		b.tone(0x100)
		b.wb(0x3070, 0)
		b.wb(0x3071, 15)
		b.vol(15)
		b.wb(0x3072, 0)
		b.wb(0x3073, 0)
	case 15:
		b.mix(1)
		b.tone(0x100)
		b.vol(12)
		b.wb(0x3074, 0)
		b.wb(0x3075, 0)
	case 16:
		b.mix(1)
		b.vol(10)
		b.restart16()
	case 17:
		b.mix(1)
		b.tone(0)
		b.wb(0x307a, 0)
		b.wb(0x307b, 15)
		b.vol(15)
		b.wb(0x307c, 0x45)
		b.wb(0x307d, 0x99)
	case 18:
		b.mix(1)
		b.ww(0x3081, 0)
		b.tone(0)
		b.wb(0x307e, 0)
		b.wb(0x307f, 15)
		b.vol(15)
		b.wb(0x3080, 0)
	case 19:
		b.mix(2)
		b.tone(0)
		b.wb(0x3083, 15)
		b.vol(15)
		b.wb(0x3084, 0x30)
		b.wb(0x3085, 0x95)
		b.wb(0x3086, 0x3d)
		b.ww(0x3087, 0x70)
		b.ww(0x3089, 0x70)
	case 20:
		b.mix(2)
		b.tone(0)
		b.wb(0x308b, 15)
		b.vol(15)
		b.wb(0x308c, 0x30)
		b.wb(0x308d, 0x95)
		b.wb(0x308e, 0x3d)
		b.ww(0x308f, 0x38)
		b.ww(0x3091, 0x38)
	case 21:
		b.mix(1)
		b.wb(0x3093, 0)
		b.reinit21(0x3094)
	case 22:
		b.mix(0)
		b.vol(9)
		b.wb(0x3099, 0xf0)
		b.wb(0x3098, 0x30)
		b.tone(0x30)
	case 23:
		b.mix(1)
		b.vol(15)
		b.wb(0x309a, 0x1d)
		b.step23b()
	case 24:
		b.mix(0)
		b.vol(10)
		b.wb(0x309e, 0x80)
		b.wb(0x309c, 0)
		b.wb(0x309d, 0xc0)
		b.tone(0xc0)
	case 25:
		b.mix(0)
		b.tone(0xc0)
		b.vol(15)
		b.wb(0x309f, 0)
		b.wb(0x30a0, 0x10)
		b.wb(0x30a1, 3)
		b.wb(0x30a2, 5)
	case 26, 27:
		b.musicStart(id - 26)
		return 0, true
	default: // 28..32: music voices
		b.mix(0)
	}
	return 0, false
}

// step runs the per-pass routine of a running effect (table at 0x0ad4) and
// returns non-zero when the effect is over.
func (b *Board) step(id uint8) uint8 {
	switch id {
	case 1:
		return b.step1()
	case 2:
		return b.step2()
	case 3:
		return b.step3()
	case 4:
		return b.step4()
	case 5:
		return b.step5()
	case 6:
		return b.step6()
	case 7:
		return b.step7()
	case 8:
		return b.step8()
	case 9:
		return b.step9()
	case 10:
		return b.step10()
	case 11:
		return b.step11()
	case 12:
		return b.step12()
	case 13:
		return b.step13()
	case 14:
		return b.step14()
	case 15:
		return b.step15()
	case 16:
		return b.step16()
	case 17:
		return b.step17()
	case 18:
		return b.step18()
	case 19:
		return b.step19()
	case 20:
		return b.step20()
	case 21:
		return b.step21()
	case 22:
		return b.step22()
	case 23:
		return b.step23()
	case 24:
		return b.step24()
	case 25:
		return b.step25()
	case 28, 29, 30, 31, 32:
		return b.musicStep(uint16(id - 28))
	}
	return 0
}

// ---- effects 1..25 ----------------------------------------------------------

func (b *Board) step1() uint8 { // 0x0402
	if b.dec(0x3041)&3 != 0 {
		return 0
	}
	v := b.dec(0x3042)
	if v == 0 {
		return 0xff
	}
	if v&7 == 0 {
		b.vol(b.dec(0x3043))
		return 0
	}
	b.tone(uint16(b.rom[0x42f+uint16(uint8(v&3)-1)]))
	return 0
}

func (b *Board) step2() uint8 { // 0x044b
	v := b.dec(0x3044)
	if v == 0 {
		return 0xff
	}
	if v == 0x6d {
		b.ww(0x3045, 4)
		b.tone(4)
		return 0
	}
	hl := b.rw(0x3045) + 0x18
	b.ww(0x3045, hl)
	b.tone(hl)
	return 0
}

func (b *Board) step3() uint8 { // 0x047a
	if b.dec(0x3049)&7 != 0 {
		return 0
	}
	hl := b.rw(0x3047) + 0x40
	if uint8(hl>>8) == 12 {
		return 0xff
	}
	b.ww(0x3047, hl)
	b.tone(hl)
	return 0
}

func (b *Board) step4() uint8 { // 0x04b3
	v := b.dec(0x304a)
	a := v & 15
	if a == 0 {
		c := b.dec(0x304b)
		if c == 0 {
			return 0xff
		}
		b.vol(c)
		b.ww(0x304c, 0x60)
		return 0
	}
	hl := b.rw(0x304c)
	if a >= 12 {
		hl += 0x10
	} else {
		hl += 0xffd0
	}
	b.ww(0x304c, hl)
	b.tone(hl)
	return 0
}

func (b *Board) step5() uint8 { // 0x04f5
	v := b.dec(0x304e)
	if v == 0 {
		return 0xff
	}
	if v&15 == 0 {
		b.vol(b.dec(0x304f))
		b.wb(0x3050, 0x20)
		b.tone(uint16(b.rb(0x3050)))
		return 0
	}
	b.tone(uint16(b.inc(0x3050)))
	return 0
}

func (b *Board) reinit6() { // 0x0517
	b.wb(0x3051, 0x40)
	b.wb(0x3052, 5)
	b.vol(5)
	b.wb(0x3053, 0x17)
	b.wb(0x3054, 0x50)
}

func (b *Board) step6() uint8 { // 0x0528
	v := b.dec(0x3051)
	if v == 0 {
		b.reinit6()
		return 0
	}
	if v&15 == 0 {
		b.vol(b.inc(0x3052))
	}
	b.wb(0x3053, b.rb(0x3053)+0x29)
	a := (b.rb(0x3053)^b.rb(0x3054))&0x3f + 0x50
	b.wb(0x3054, a)
	b.tone(uint16(a))
	return 0
}

// step7b is the body at 0x055a shared by start and step of effect 7.
func (b *Board) step7b() uint8 {
	v := b.dec(0x3056)
	if v == 0 {
		return 0xff
	}
	b.vol(v)
	b.ww(0x3057, 0)
	b.tone(0)
	return 0
}

func (b *Board) step7() uint8 { // 0x056c
	v := b.dec(0x3055)
	if v == 0 {
		b.wb(0x3055, 0x2c)
		return b.step7b()
	}
	if v == 0x16 {
		b.ww(0x3057, 0)
		b.tone(0)
		return 0
	}
	hl := b.rw(0x3057) + 0x10
	b.ww(0x3057, hl)
	b.tone(hl)
	return 0
}

func (b *Board) step8() uint8 { // 0x05a4
	v := b.inc(0x305b)
	if v == 0x59 {
		a := b.rb(0x305c) - 4
		if a == 0 {
			return 0xff
		}
		b.wb(0x305c, a)
		b.vol(a)
		b.wb(0x305b, 0)
		b.ww(0x3059, 0x80)
		b.tone(0x80)
		return 0
	}
	hl := b.rw(0x3059) + uint16(v)
	b.ww(0x3059, hl)
	b.tone(hl)
	return 0
}

// reinit9 is 0x05d5: re-seed effect 9 starting at the byte after hl.
func (b *Board) reinit9(hl uint16) {
	hl++ // 305f
	b.wb(hl, 13)
	b.vol(13)
	b.wb(hl+1, 0x93)
	b.wb(hl+2, 0xd5)
	b.tone(0xc0)
}

func (b *Board) step9() uint8 { // 0x05ea
	if b.dec(0x305d)&1 != 0 {
		return 0
	}
	v := b.dec(0x305e)
	if v == 0 {
		return 0xff
	}
	if v == 0xd0 {
		b.reinit9(0x305e)
		return 0
	}
	if v&15 == 0 {
		b.vol(b.dec(0x305f))
	}
	b.wb(0x3060, b.rb(0x3060)+0x53)
	a := b.rb(0x3060) ^ b.rb(0x3061)
	b.wb(0x3061, a)
	h := uint8(0)
	if a < 0xe0 {
		h = a
	}
	b.tone(uint16(h)<<8 | uint16(a))
	return 0
}

func (b *Board) restart10() { // 0x062a
	c := b.dec(0x3064)
	b.vol(c)
	b.ww(0x3065, 0x20)
	b.tone(0x20)
}

func (b *Board) step10() uint8 { // 0x063a
	if b.dec(0x3062)&1 != 0 {
		return 0
	}
	v := b.dec(0x3063)
	if v == 0 {
		return 0xff
	}
	if v&15 == 0 {
		b.restart10()
		return 0
	}
	hl := b.rw(0x3065) + 0x40
	b.ww(0x3065, hl)
	b.tone(hl)
	return 0
}

func (b *Board) step11() uint8 { // 0x0677
	v := b.dec(0x3067)
	if v == 0 {
		return 0xff
	}
	if v == 0x60 {
		b.wb(0x3068, 0x1f)
		b.tone(0x1f)
		return 0
	}
	if v&3 != 0 {
		return 0
	}
	a := b.rb(0x3068) + 0x20
	b.wb(0x3068, a)
	b.tone(uint16(a))
	return 0
}

func (b *Board) step12() uint8 { // 0x06b4
	d := b.dec(0x3069)
	a := b.rb(0x306a) + 0xd5
	b.wb(0x306a, a)
	if d&1 == 0 {
		a ^= b.rb(0x306b)
		b.wb(0x306b, a)
		a &= 0xef
		b.vol(a)
	}
	a ^= b.rb(0x306c)
	b.wb(0x306c, a)
	b.tone(uint16(a | 0x80))
	return 0
}

func (b *Board) restart13() { // 0x06e1
	b.wb(0x306d, 0x28)
	b.wb(0x306e, 5)
	b.wb(0x306f, 0x30)
	b.tone(0x600)
	b.vol(8)
}

func (b *Board) step13() uint8 { // 0x06f6
	v := b.dec(0x306d)
	if v == 0 {
		b.restart13()
		return 0
	}
	b.vol(v&7 + 8)
	b.wb(0x306e, b.rb(0x306e)+0xb9)
	a := b.rb(0x306e) ^ b.rb(0x306f)
	b.wb(0x306f, a)
	b.tone(0x600 | uint16(a))
	return 0
}

func (b *Board) step14() uint8 { // 0x0732
	v := b.dec(0x3070)
	if v&1 == 0 {
		b.dec(0x3071)
		if v&0x10 == 0 {
			b.inc(0x3071)
			b.inc(0x3071)
		}
		b.vol(b.rb(0x3071))
	}
	b.wb(0x3072, b.rb(0x3072)+0xd3)
	a := b.rb(0x3072) ^ b.rb(0x3073)
	b.wb(0x3073, a)
	a = a&0x7f + 0xa8
	b.tone(uint16(a&7)<<8 | uint16(a))
	return 0
}

func (b *Board) step15() uint8 { // 0x0774
	b.wb(0x3074, b.rb(0x3074)+0x97)
	a := b.rb(0x3074) ^ b.rb(0x3075)
	b.wb(0x3075, a)
	a = a&0x7f + 0xc0
	b.tone(0x100 | uint16(a))
	return 0
}

func (b *Board) restart16() { // 0x0797
	b.wb(0x3076, 0x30)
	b.wb(0x3077, 2)
	b.wb(0x3078, 0xe0)
	b.wb(0x3079, 0xe0)
	b.tone(0xe0)
}

func (b *Board) step16() uint8 { // 0x07aa
	v := b.dec(0x3076)
	if v == 0 {
		b.restart16()
		return 0
	}
	if v&15 == 0 {
		b.wb(0x3077, b.rb(0x3077)+2)
		a := b.rb(0x3078) - 0x20
		b.wb(0x3078, a)
		b.wb(0x3079, a)
		b.tone(uint16(a))
		return 0
	}
	a := b.rb(0x3077) + b.rb(0x3079)
	b.wb(0x3079, a)
	b.tone(uint16(a))
	return 0
}

func (b *Board) step17() uint8 { // 0x07e5
	v := b.dec(0x307a)
	if v&1 != 0 {
		return 0
	}
	if v&0x3f == 0 {
		c := b.dec(0x307b)
		if c == 0 {
			return 0xff
		}
		b.vol(c)
	}
	b.wb(0x307c, b.rb(0x307c)+0xd3)
	a := b.rb(0x307c) ^ b.rb(0x307d)
	b.wb(0x307d, a)
	b.tone(uint16(a)<<8 | uint16(a))
	return 0
}

func (b *Board) step18() uint8 { // 0x0829
	d := b.dec(0x307e)
	if d&0x1f == 0 {
		c := b.dec(0x307f)
		if c == 0 {
			return 0xff
		}
		b.vol(c)
	}
	b.wb(0x3080, b.rb(0x3080)+0xc5)
	a := b.rb(0x3080) ^ d ^ b.rb(0x3081)
	h := a
	if a < 0x80 {
		h = 1
	}
	b.ww(0x3081, uint16(h)<<8|uint16(a))
	b.tone(uint16(h)<<8 | uint16(a))
	return 0
}

func (b *Board) step19() uint8 { // 0x087f
	v := b.dec(0x3084)
	switch {
	case v == 0:
		c := b.dec(0x3083)
		if c == 0 {
			return 0xff
		}
		b.vol(c)
		b.wb(0x3084, b.rb(0x3085)&0x3f+0x30)
		hl := b.rw(0x3087) + 0x60
		b.ww(0x3087, hl)
		b.ww(0x3089, hl)
	case v >= 0x10:
		b.wb(0x3085, b.rb(0x3085)+0xc9)
		a := b.rb(0x3085) ^ b.rb(0x3086)
		b.wb(0x3086, a)
		b.tone(0x100 | uint16(a))
	default:
		c := uint8(0x10-v) * 2
		hl := b.rw(0x3089) + uint16(c)
		b.ww(0x3089, hl)
		b.tone(hl)
	}
	return 0
}

func (b *Board) step20() uint8 { // 0x08f8
	v := b.dec(0x308c)
	switch {
	case v == 0:
		c := b.dec(0x308b)
		if c == 0 {
			return 0xff
		}
		b.vol(c)
		b.wb(0x308c, b.rb(0x308d)&0x3f+0x30)
		hl := b.rw(0x308f) + 6
		b.ww(0x308f, hl)
		b.ww(0x3091, hl)
	case v >= 0x10:
		b.wb(0x308d, b.rb(0x308d)+0xd3)
		a := b.rb(0x308d) ^ b.rb(0x308e)
		b.wb(0x308e, a)
		b.tone(0x100 | uint16(a))
	default:
		hl := b.rw(0x3091) + 0x30
		b.ww(0x3091, hl)
		b.tone(hl)
	}
	return 0
}

// reinit21 is 0x0951: seed effect 21 (hl = address of its second byte).
func (b *Board) reinit21(hl uint16) {
	b.wb(hl, 0xd0)
	b.wb(hl+1, 13)
	b.vol(13)
	b.wb(hl+2, 0x93)
	b.wb(hl+3, 0xd5)
	b.tone(0xc0)
}

func (b *Board) step21() uint8 { // 0x0966
	if b.dec(0x3093)&3 != 0 {
		return 0
	}
	v := b.dec(0x3094)
	if v == 0 {
		return 0xff
	}
	if v == 0x98 || v == 0x80 {
		b.wb(0x3095, 13)
		b.vol(13)
		b.wb(0x3096, 0x93)
		b.wb(0x3097, 0xd5)
		b.tone(0xc0)
		return 0
	}
	if v&15 == 0 {
		b.vol(b.dec(0x3095))
	}
	b.wb(0x3096, b.rb(0x3096)+0x53)
	a := b.rb(0x3096) ^ b.rb(0x3097)
	b.wb(0x3097, a)
	h := uint8(0)
	if a < 0xe0 {
		h = a
	}
	b.tone(uint16(h)<<8 | uint16(a))
	return 0
}

func (b *Board) step22() uint8 { // 0x09bc
	a := b.rb(0x3098) + 6
	b.wb(0x3098, a)
	if a == b.rb(0x3099) {
		a -= 6
		b.wb(0x3099, a)
		if a-0x7e == 0 {
			return 0xff
		}
		b.wb(0x3098, 0x30)
		b.tone(0x30)
		return 0
	}
	b.tone(uint16(a))
	return 0
}

func (b *Board) step23b() uint8 { // 0x09e1
	a := b.dec(0x309a)
	if a == 0 {
		return 0xff
	}
	b.wb(0x309b, a)
	b.tone(uint16(a * 2))
	return 0
}

func (b *Board) step23() uint8 { // 0x09f1
	if b.dec(0x309b) == 0 {
		return b.step23b()
	}
	b.tone(uint16(b.rb(0x309b)) * 2)
	return 0
}

func (b *Board) step24() uint8 { // 0x0a20
	v := b.dec(0x309c)
	a := b.rb(0x309d)
	if v&1 != 0 {
		b.tone(uint16(a >> 1))
		return 0
	}
	a = b.dec(0x309d)
	if a == 0x20 {
		return 0xff
	}
	if a == b.rb(0x309e) {
		b.wb(0x309e, a+5)
		b.wb(0x309d, 0xc0)
		b.tone(0xc0)
		return 0
	}
	b.tone(uint16(a))
	return 0
}

func (b *Board) step25() uint8 { // 0x0a6a
	v := b.dec(0x309f)
	a := v & 0x3f
	if a == 0 {
		if b.dec(0x30a0) == 0 {
			return 0xff
		}
	}
	if a&3 != 0 {
		return 0
	}
	d := b.rb(0x30a0)
	b.wb(0x30a1, b.rb(0x30a1)+5)
	n := (b.rb(0x30a1) ^ b.rb(0x30a2)) & 15
	b.wb(0x30a2, n)
	if n >= d {
		n = 0
	}
	b.vol(n)
	return 0
}
