package game

// Events let a front end (the Time Pilot+ statistics) watch what happens in
// the game without changing it. They are emitted from the routines that
// already perform the action; the game never reads anything back.

// EventKind says what happened.
type EventKind int

const (
	EvShot       EventKind = iota + 1 // the player fired a bullet
	EvHit                             // a bullet struck an enemy or a boss
	EvKill                            // something was destroyed; N is the kill-chain length
	EvBoss                            // the stage boss was destroyed; Points is its award
	EvRescue                          // a parachutist was picked up; N is how many in this stage
	EvDeath                           // the player lost a plane
	EvStageClear                      // the stage was completed
	EvScore                           // points were added to the current player's score
)

// Event is one thing that happened. Stage is the game's stage number 0..4
// (1910, 1940, 1970, 1982, 2001), Player 0 or 1.
type Event struct {
	Kind   EventKind
	Stage  int
	Player int
	N      int
	Points int
}

func (g *Game) emit(k EventKind, n, points int) {
	if g.OnEvent == nil {
		return
	}
	g.OnEvent(Event{Kind: k, Stage: int(g.rb(vStage)), Player: int(g.rb(vPlayer2) & 1), N: n, Points: points})
}

// scoreValue decodes entry idx of the game's 3-byte BCD points table.
func (g *Game) scoreValue(idx uint8) int {
	a := 0x0d27 + 3*uint16(idx)
	bcd := func(v uint8) int { return int(v>>4)*10 + int(v&15) }
	return bcd(g.rb(a+2))*10000 + bcd(g.rb(a+1))*100 + bcd(g.rb(a))
}

// ActiveGame reports whether a real (not demo) game is in progress.
func (g *Game) ActiveGame() bool { return g.rb(vGameActive) != 0 }

// CurrentPlayer is 0 or 1.
func (g *Game) CurrentPlayer() int { return int(g.rb(vPlayer2) & 1) }

// Stage is the current stage number 0..4.
func (g *Game) Stage() int { return int(g.rb(vStage)) }

// HiScore is the current high score.
func (g *Game) HiScore() int {
	bcd := func(v uint8) int { return int(v>>4)*10 + int(v&15) }
	return bcd(g.rb(vHiScore+2))*10000 + bcd(g.rb(vHiScore+1))*100 + bcd(g.rb(vHiScore))
}

// Score is player p's score.
func (g *Game) Score(p int) int {
	a := uint16(vScoreP1)
	if p == 1 {
		a = vScoreP2
	}
	bcd := func(v uint8) int { return int(v>>4)*10 + int(v&15) }
	return bcd(g.rb(a+2))*10000 + bcd(g.rb(a+1))*100 + bcd(g.rb(a))
}

// Lives is the number of planes player p has left (including the one in
// play), 0 when that player is out or not playing.
func (g *Game) Lives(p int) int {
	if g.rb(vGameActive) == 0 {
		return 0
	}
	blk := uint16(0xad10)
	if p == 1 {
		blk = 0xad20
	}
	if (g.rb(vPlayer2) != 0) == (p == 1) {
		blk = 0xad00
	}
	return int(g.rb(blk))
}

// EnemyArt returns, for each of the five stages, the sprite code and raw
// attribute byte the stage's enemy planes use when flying level (used only to
// picture them in the statistics panel).
func (g *Game) EnemyArt() (code, attr [5]uint8) {
	code[0], attr[0] = g.rb(0x2a77), g.rb(0x2a77+0x10)
	code[1], attr[1] = g.rb(0x2a77)+0x10, g.rb(0x2a77+0x10)+0x35
	code[2], attr[2] = g.rb(0x2abc), g.rb(0x2abc+1)
	code[3], attr[3] = g.rb(0x2b18), g.rb(0x2b18+0x10)
	code[4], attr[4] = 0xd8, 0x61
	return
}
