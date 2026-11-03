// lockstep compares the native Go game against the reference machine frame by
// frame, starting from the reference state at the first vblank interrupt.
package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"timepilot/internal/assets"
	"timepilot/internal/game"
	"timepilot/tools/oracle/machine"
)

var nmis int
var held uint8
var trace int
var bossEvery int
var seed int64 = 1
var p2 bool
var stageEnd int
var startStage int

type script struct {
	rnd *rand.Rand
}

func inputs(f int, rnd *rand.Rand, m *machine.Machine, g *game.Game) {
	in0, in1, in2 := uint8(0xff), uint8(0xff), uint8(0xff)
	if f >= 66 && f < 70 {
		in0 &^= 1
	}
	if p2 && f >= 90 && f < 94 {
		in0 &^= 1
	}
	if f >= 1166 && f < 1170 {
		if p2 {
			in0 &^= 0x10
		} else {
			in0 &^= 8
		}
	}
	if f > 1216 {
		// held controls that change every so often, fire in bursts
		if f%23 == 0 || f == 1217 {
			held = uint8(rnd.Intn(16))
		}
		in1 &^= held & 3
		if held&4 != 0 && f%60 < 20 {
			in1 &^= 4
		}
		if held&8 != 0 && f%80 < 15 {
			in1 &^= 8
		}
		if (f/7)%2 == 0 || rnd.Intn(4) == 0 {
			in1 &^= 0x10
		}
		// player died: press start again to continue sometimes
	}
	m.In0, m.In1, m.In2 = in0, in1, in2
	g.In0, g.In1, g.In2 = in0, in1, in2
}

func main() {
	r, err := assets.Load(os.Args[1])
	if err != nil {
		panic(err)
	}
	maxF := 3000
	if v := os.Getenv("BOSS"); v != "" {
		bossEvery, _ = strconv.Atoi(v)
	}
	if v := os.Getenv("SEED"); v != "" {
		seed, _ = strconv.ParseInt(v, 10, 64)
	}
	p2 = os.Getenv("P2") != ""
	if v := os.Getenv("STAGEEND"); v != "" {
		stageEnd, _ = strconv.Atoi(v)
	}
	if v := os.Getenv("STAGE"); v != "" {
		startStage, _ = strconv.Atoi(v)
	}
	if v := os.Getenv("TRACE"); v != "" {
		trace, _ = strconv.Atoi(v)
	}
	if len(os.Args) > 2 {
		maxF, _ = strconv.Atoi(os.Args[2])
	}
	m := machine.New(r)
	start := 234
	for !(m.Frame == start && m.Line == 240) {
		m.In0, m.In1, m.In2 = 0xff, 0xff, 0xff
		m.RunLine()
	}
	g := game.New(r.Main[:])
	if os.Getenv("COLD") != "" {
		g.Boot()
		bad := 0
		for _, c := range []struct {
			n    string
			a, b []byte
			base int
		}{{"ram", m.WorkRAM(), g.Work(), 0xa800}, {"vram", m.VideoRAM(), g.Video(), 0xa400}, {"cram", m.ColorRAM(), g.Color(), 0xa000}, {"spr", m.SpriteRAM(), g.Sprites1(), 0xb000}, {"spr2", m.SpriteRAM2(), g.Sprites2(), 0xb400}} {
			for i := range c.a {
				if c.n == "ram" && i >= 0x7c0 {
					continue
				}
				if c.a[i] != c.b[i] {
					if bad < 40 {
						fmt.Printf("  cold %s[%04x] oracle=%02x go=%02x\n", c.n, c.base+i, c.a[i], c.b[i])
					}
					bad++
				}
			}
		}
		fmt.Println("cold boot mismatches:", bad, "nmi enabled:", g.NMIEnabled)
	} else {
		g.LoadState(m.WorkRAM(), m.VideoRAM(), m.ColorRAM(), m.SpriteRAM(), m.SpriteRAM2())
	}
	fmt.Printf("snapshot a9eb=%02x a9ab=%02x a9ac=%02x\n", m.WorkRAM()[0x1eb], m.WorkRAM()[0x1ab], m.WorkRAM()[0x1ac])
	rnd := rand.New(rand.NewSource(seed))
	idle := true
	seen := false
	m.MainTrace = func(pc uint16) {
		if pc == 0x66 {
			nmis++
			idle = false
			seen = true
		} else if pc == 0x0b93 {
			w := m.WorkRAM()
			if w[0x1b2] == w[0x1b3] {
				idle = true
			}
		}
	}
	for f := 0; f < maxF; f++ {
		inputs(f, rnd, m, g)
		if stageEnd > 0 && f%stageEnd == stageEnd-1 && m.WorkRAM()[0x1ab] == 3 && m.WorkRAM()[0x1ac] == 7 && m.WorkRAM()[0] == 0xff && m.WorkRAM()[0x530] != 0 {
			for _, w := range [][]byte{m.WorkRAM(), g.Work()} {
				for i := 0x10; i < 0x100; i++ {
					w[i] = 0
				}
				w[0x502] = 0
				w[0x4c6] = 0xff
			}
			fmt.Println("forcing stage end at frame", f)
		}
		if bossEvery > 0 && f%bossEvery == bossEvery-1 && m.WorkRAM()[0x1ab] == 3 && m.WorkRAM()[0x1ac] == 7 && m.WorkRAM()[0x530] != 0 {
			for _, w := range [][]byte{m.WorkRAM(), g.Work()} {
				w[0x502] = 0
				for i := 0x50; i < 0xb0; i += 0x10 {
					w[i] = 0
				}
			}
		}
		if startStage > 0 && m.WorkRAM()[0x1ab] == 3 && m.WorkRAM()[0x1ac] == 4 && m.WorkRAM()[0x504] != byte(startStage) && m.WorkRAM()[0x530] != 0 {
			fmt.Println("poking stage at frame", f)
			m.WorkRAM()[0x504] = byte(startStage)
			g.Work()[0x504] = byte(startStage)
		}
		seen = false
		nmis = 0
		l0 := m.Line
		// run to the next vertical-blank interrupt, then let its handler
		// (which may overrun the frame) and the draw queue finish
		for n := 0; (!seen || !idle) && n < 600; n++ {
			m.RunLine()
		}
		if trace > 0 && f >= trace {
			fmt.Printf("  nmis=%d line %d->%d\n", nmis, l0, m.Line)
		}
		var perr interface{}
		func() {
			defer func() { perr = recover() }()
			for i := 0; i < nmis || i == 0; i++ {
				g.Frame()
			}
		}()
		if perr != nil {
			fmt.Printf("frame %d: Go port stopped: %v\n", f, perr)
			os.Exit(1)
		}
		bad := 0
		cmp := func(name string, a, b []byte, base int, skip func(i int) bool) {
			for i := range a {
				if skip != nil && skip(i) {
					continue
				}
				if a[i] != b[i] {
					if bad < 40 {
						fmt.Printf("  %s[%04x] oracle=%02x go=%02x\n", name, base+i, a[i], b[i])
					}
					bad++
				}
			}
		}
		// stack area at the top of work RAM is not game state
		skipStack := func(i int) bool { return i >= 0x7c0 }
		hdr := fmt.Sprintf("frame %d state %02x/%02x stage=%d ctr=%02x a800=%02x", f, m.WorkRAM()[0x1ab], m.WorkRAM()[0x1ac], m.WorkRAM()[0x504], m.WorkRAM()[0x180], m.WorkRAM()[0])
		if trace > 0 && f >= trace {
			fmt.Println(hdr)
		}
		cmp("ram", m.WorkRAM(), g.Work(), 0xa800, skipStack)
		cmp("vram", m.VideoRAM(), g.Video(), 0xa400, nil)
		cmp("cram", m.ColorRAM(), g.Color(), 0xa000, nil)
		if bad > 0 {
			fmt.Printf("%s: %d mismatches\n", hdr, bad)
			os.Exit(1)
		}
	}
	fmt.Println("lockstep OK through frame", maxF)
}
