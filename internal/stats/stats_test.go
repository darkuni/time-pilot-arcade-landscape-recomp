package stats

import (
	"path/filepath"
	"testing"

	"timepilot/internal/game"
)

func newT() *Tracker {
	t := &Tracker{Players: 1}
	t.Playing = true
	return t
}

func TestEventsPerPlayer(t *testing.T) {
	tr := newT()
	tr.Event(game.Event{Kind: game.EvKill, Stage: 2, Player: 0, N: 3})
	tr.Event(game.Event{Kind: game.EvKill, Stage: 2, Player: 1, N: 1})
	tr.Event(game.Event{Kind: game.EvShot, Player: 1})
	tr.Event(game.Event{Kind: game.EvHit, Player: 1})
	tr.Event(game.Event{Kind: game.EvBoss, Player: 0})
	tr.Event(game.Event{Kind: game.EvRescue, Player: 0})
	if tr.GameP[0].Kills != 1 || tr.GameP[0].Stage[2] != 1 || tr.GameP[0].BestChain != 3 ||
		tr.GameP[0].Bosses != 1 || tr.GameP[0].Rescued != 1 {
		t.Fatalf("p1 %+v", tr.GameP[0])
	}
	if tr.GameP[1].Kills != 1 || tr.GameP[1].Accuracy() != 100 {
		t.Fatalf("p2 %+v", tr.GameP[1])
	}
}

func TestDemoIgnored(t *testing.T) {
	tr := newT()
	tr.Playing = false
	tr.Event(game.Event{Kind: game.EvKill})
	if tr.GameP[0].Kills != 0 {
		t.Fatal("demo counted")
	}
}

func TestSaveLoad(t *testing.T) {
	tr := newT()
	tr.Event(game.Event{Kind: game.EvKill, Stage: 4, N: 7})
	f := filepath.Join(t.TempDir(), "s.json")
	if err := tr.Save(f); err != nil {
		t.Fatal(err)
	}
	u := newT()
	u.Load(f)
	if u.LifeP[0].Kills != 1 || u.LifeP[0].BestChain != 7 || u.LifeP[0].Stage[4] != 1 {
		t.Fatalf("%+v", u.LifeP[0])
	}
}
