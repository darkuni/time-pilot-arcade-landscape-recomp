package game

// coinHandling is routine 0x48be: coin and credit processing, run every frame.
func (g *Game) coinHandling() {
	g.serviceSwitch()
	g.coinSlot1()
	g.coinSlot2()
	g.coinCounter1()
	g.coinCounter2()
}

// edge shifts the current sample into a 3-deep history register and
// reports a clean 0,0,1 transition.
func (g *Game) edge(reg uint16, bit bool) bool {
	v := g.rb(reg) << 1
	if bit {
		v |= 1
	}
	g.wb(reg, v)
	return v&7 == 1
}

func (g *Game) serviceSwitch() { // 0x48e7
	if g.edge(0xa983, g.rb(vInStart)&0x04 != 0) {
		g.sndPush(g.k(0x322e))
		g.addCredits(1)
	}
}

// coinSlot services one coin input: reg = history register, bit = input
// mask, ctr = per-slot coin counter, st = three-byte state (history, partial
// coins, coins-per-credit).
func (g *Game) coinSlot(reg uint16, mask uint8, ctr uint16) {
	if !g.edge(reg, g.rb(vInStart)&mask != 0) {
		return
	}
	g.sndPush(g.k(0x322e))
	g.wb(ctr, g.rb(ctr)+1)
	hl := reg + 1
	b := g.rb(hl) + 0x10
	g.wb(hl, b)
	hl++
	need := g.rb(hl)
	if need >= b { // SUB B without carry: not enough coins yet
		return
	}
	c := need
	a := -(need&0xf0 + 0x10) + g.rb(hl-1) // NEG of ((need&0xf0)+0x10), plus the counter
	g.wb(hl-1, a)
	g.addCredits(c)
}

func (g *Game) coinSlot1() { g.coinSlot(0xa9c7, 0x01, vCoin1Cnt) } // 0x4941
func (g *Game) coinSlot2() { g.coinSlot(0xa9ca, 0x02, vCoin2Cnt) } // 0x4911

// addCredits is routine 0x496e: award credits unless the machine is on free
// play.
func (g *Game) addCredits(c uint8) {
	if g.rb(0xa9c0) == 0 {
		v, carry := bcdAdd8(c&15, g.rb(vCredits), false)
		if carry {
			v = 0x99
		}
		g.wb(vCredits, v)
		g.drawCredits()
	}
	// 0x496e falls straight into the slot-1 pulse code (0x4984), so a credit
	// frame services that pulse one extra time.
	g.coinCounter1()
}

// coinCounter1 / 2 drive the electromechanical coin counter outputs with a
// short pulse for every coin accepted.
func (g *Game) coinCounter1() { g.pulse(vCoin1Cnt, vCoin1Tmr, 0xc30a) }
func (g *Game) coinCounter2() { g.pulse(vCoin2Cnt, vCoin2Tmr, 0xc30c) }

func (g *Game) pulse(cnt, tmr, port uint16) {
	if g.rb(cnt) == 0 {
		return
	}
	t := g.rb(tmr)
	if t == 0 {
		g.wb(tmr, 0x30)
		g.wb(port, 1)
		return
	}
	t--
	g.wb(tmr, t)
	if t == 0 {
		g.wb(cnt, g.rb(cnt)-1)
		return
	}
	if t == 0x18 {
		g.wb(port, 0)
	}
}
