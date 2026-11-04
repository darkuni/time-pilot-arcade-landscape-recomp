// Package stats keeps the Time Pilot+ statistics and draws them in the side
// panels. It only listens to the game (game.Event and a few read-only
// accessors); it never changes it.
package stats

import (
	"encoding/json"
	"image"
	"os"

	"timepilot/internal/game"
)

// Stages are the five eras of the game (1910, 1940, 1970, 1982, 2001).
const Stages = 5

// Counts are one set of totals (this game, or all games) for one player.
type Counts struct {
	Games     int
	Shots     int
	Hits      int
	Kills     int
	Stage     [Stages]int // kills per era
	Bosses    int
	Rescued   int
	Deaths    int
	Cleared   int // stages completed
	BestChain int // longest run of kills inside the game's chain window
}

type Tracker struct {
	GameP, LifeP [2]Counts
	Players      int // 1 or 2 players in the running game
	Cur          int // whose turn it is
	Playing      bool
	High         int
	Icons        [Stages]*image.RGBA
	Dirty        bool

	g *game.Game
}

func New(g *game.Game) *Tracker {
	t := &Tracker{g: g, Players: 1}
	g.OnEvent = t.Event
	return t
}

func (t *Tracker) both(p int, f func(c *Counts)) {
	f(&t.GameP[p])
	f(&t.LifeP[p])
	t.Dirty = true
}

// Event is the game's event hook.
func (t *Tracker) Event(e game.Event) {
	if !t.Playing || e.Player < 0 || e.Player > 1 {
		return
	}
	s := e.Stage
	if s < 0 || s >= Stages {
		s = 0
	}
	switch e.Kind {
	case game.EvShot:
		t.both(e.Player, func(c *Counts) { c.Shots++ })
	case game.EvHit:
		t.both(e.Player, func(c *Counts) { c.Hits++ })
	case game.EvKill:
		t.both(e.Player, func(c *Counts) {
			c.Kills++
			c.Stage[s]++
			if e.N > c.BestChain {
				c.BestChain = e.N
			}
		})
	case game.EvBoss:
		t.both(e.Player, func(c *Counts) { c.Bosses++ })
	case game.EvRescue:
		t.both(e.Player, func(c *Counts) { c.Rescued++ })
	case game.EvDeath:
		t.both(e.Player, func(c *Counts) { c.Deaths++ })
	case game.EvStageClear:
		t.both(e.Player, func(c *Counts) { c.Cleared++ })
	}
}

// Frame is called once per emulated frame, after the game has run.
func (t *Tracker) Frame() {
	t.High = t.g.HiScore()
	playing := t.g.ActiveGame()
	if playing && !t.Playing {
		// a new game: both seats start fresh; a second player is present
		// when the second player's plane block holds lives
		t.GameP = [2]Counts{}
		t.Players = 1
		t.Cur = 0
		t.LifeP[0].Games++
		t.GameP[0].Games = 1
		if t.g.Lives(1) > 0 {
			t.Players = 2
			t.LifeP[1].Games++
			t.GameP[1].Games = 1
		}
		t.Dirty = true
	}
	t.Playing = playing
	if playing {
		t.Cur = t.g.CurrentPlayer()
	}
}

// Accuracy is hits per shot as a whole percentage.
func (c *Counts) Accuracy() int {
	if c.Shots == 0 {
		return 0
	}
	return c.Hits * 100 / c.Shots
}

// ---- saving ------------------------------------------------------------

type saved struct {
	Life  Counts // player 1
	Life2 Counts // player 2
}

// Load reads the all-time totals from path (a missing file is fine).
func (t *Tracker) Load(path string) {
	d, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var s saved
	if json.Unmarshal(d, &s) != nil {
		return
	}
	t.LifeP[0], t.LifeP[1] = s.Life, s.Life2
}

// Save writes the all-time totals to path.
func (t *Tracker) Save(path string) error {
	d, err := json.MarshalIndent(saved{Life: t.LifeP[0], Life2: t.LifeP[1]}, "", " ")
	if err != nil {
		return err
	}
	t.Dirty = false
	return os.WriteFile(path, d, 0o644)
}
