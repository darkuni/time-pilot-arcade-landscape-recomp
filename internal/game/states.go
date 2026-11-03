package game

import "fmt"

// The top-level state machine.  vState (0xa9ab) selects one of four phases and
// vSub (0xa9ac) selects the step within it.  The step handlers are found
// through the original jump tables in program ROM so that every table entry
// maps 1:1 onto a routine of the original (see run()).

func (g *Game) dispatch(table uint16, idx uint8) { g.run(g.tbl16(table, idx)) }

// stateBoot: power-up housekeeping (jump table 0x15c8).
func (g *Game) stateBoot() { g.dispatch(0x15c8, g.rb(vSub)&7) }

// stateAttract: title / attract sequence (jump table 0x1659).
func (g *Game) stateAttract() {
	g.dispatch(0x1659, g.rb(vSub))
	// common tail 0x167b
	if g.rb(vCredits) != 0 {
		g.nextState()
		return
	}
	if g.rb(0xa9c0) == 0 {
		return
	}
	if g.rb(vInStart)&0x18 == 0 {
		return
	}
	g.hideSprites()
	g.startFromAttract()
}

// startFromAttract is routine 0x1690: a start button was pressed on free play.
func (g *Game) startFromAttract() {
	a := g.rb(vInStart)
	switch {
	case a&0x10 != 0:
		g.wb(vGameActive, 0xff)
		g.wb(0xad31, 0xff)
		g.wb(0xad10, g.rb(0xa9c1))
		g.wb(0xad20, g.rb(0xa9c1))
	case a&0x08 != 0:
		g.wb(0xad31, 0)
		g.wb(0xad20, 0)
		g.wb(vGameActive, 0xff)
		g.wb(0xad10, g.rb(0xa9c1))
	default:
		return
	}
	g.enterPlay()
}

// enterPlay is the tail at 0x172a.
func (g *Game) enterPlay() {
	g.wb(vState, 3)
	g.wb(vSub, 0)
}

// stateReady: "insert coin / press start" screens (jump table 0x1806).
func (g *Game) stateReady() { g.dispatch(0x1806, g.rb(vSub)) }

// statePlay: the game proper (jump table 0x0f29).
func (g *Game) statePlay() {
	g.dispatch(0x0f29, g.rb(vSub)&15)
	// common tail 0x0f54
	if g.rb(vGameActive) != 0 {
		return
	}
	if g.rb(vCredits) != 0 {
		g.wb(vSub, 0)
		g.wb(vState, g.k(0x1736))
		return
	}
	if g.rb(0xa9c0) == 0 {
		return
	}
	if g.rb(vInStart)&0x18 == 0 {
		return
	}
	g.hideSprites()
	g.startFromAttract()
}

// run executes the original routine at the given ROM address.
func (g *Game) run(addr uint16) {
	f, ok := g.routines()[addr]
	if !ok {
		panic(fmt.Sprintf("unported routine %04x (state %02x/%02x)", addr, g.rb(vState), g.rb(vSub)))
	}
	f()
}
