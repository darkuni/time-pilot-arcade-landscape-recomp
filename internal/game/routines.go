package game

// routines maps ROM entry addresses (as found in the original jump tables) to
// their Go implementations.
func (g *Game) routines() map[uint16]func() {
	if g.rtab == nil {
		g.rtab = map[uint16]func(){
			// state 0: boot
			0x15e2: g.boot15e2,
			0x15fe: g.boot15fe,
			// state 1: attract
			0x074b: g.att074b,
			0x1734: g.att1734,
			0x2d3f: g.att2d3f,
			0x083e: g.att083e,
			0x1748: g.att1748,
			0x176a: g.att176a,
			0x178c: g.att178c,
			0x17b9: g.att17b9,
			0x3252: g.att3252,
			0x17e2: g.att17e2,
			0x4b19: g.att4b19,
			0x17fb: g.nextSub,
			0x2730: g.att2730,
			// state 3: play
			0x27b1: g.play27b1,
			0x335e: g.play335e,
			0x5bd7: g.play5bd7,
			0x4c75: g.run4c75,
			0x0774: g.play0774,
			0x16af: g.play16af,
			0x5694: g.play5694,
			0x1199: g.playFrame,
			0x330b: g.play330b,
			0x08b4: g.play08b4,
			0x18c3: g.play18c3,
			0x12e2: g.play12e2,
			0x12fb: g.play12fb,
			0x4a0f: g.play4a0f,
			0x1323: g.play1323,
			0x15b5: g.play15b5,
			// state 2: ready
			0x181e: g.rdy181e,
			0x2cdb: g.rdy2cdb,
			0x1830: g.rdy1830,
			0x07e6: g.rdy07e6,
			0x188a: g.rdy188a,
		}
	}
	return g.rtab
}
