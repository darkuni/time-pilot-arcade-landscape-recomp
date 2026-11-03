// sndcheck drives the sound board of the reference machine and the native
// Go sequencer with the same commands and compares every AY register write.
package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"timepilot/internal/assets"
	"timepilot/internal/sound"
	"timepilot/tools/oracle/machine"
)

type w struct {
	chip     int
	reg, val uint8
}

func run(r *assets.ROMs, cmds []int, gap int, passes int) (ora, got []w, oraRAM, goRAM []byte) {
	// reference
	m := machine.New(r)
	m.LogAY = true
	pass := 0
	inject := 0
	var lastPC uint16
	m.SoundTrace = func(pc uint16) {
		lastPC = pc
		if pc == 0x00c8 {
			pass++
		}
	}
	// run to the first pass
	for pass < 1 {
		m.SoundRun(100)
	}
	startPass := pass
	next := 0
	for pass < startPass+passes {
		// wait until we are in the idle part of a window after pass start
		m.SoundRun(2500)
		if next < len(cmds) && pass-startPass >= gap*(next+1)-1 {
			// deliver only once the pass is over and the CPU idles in its wait loop
			for lastPC < 0xc0 || lastPC > 0xc5 {
				m.SoundRun(1)
			}
			m.SoundCommand(uint8(cmds[next]))
			next++
		}
		// finish this window
		for p := pass; pass == p && pass < startPass+passes; {
			m.SoundRun(50)
		}
	}
	_ = inject
	for _, a := range m.AYLog {
		ora = append(ora, w{a.Chip, a.Reg, a.Val})
	}
	oraRAM = append([]byte(nil), m.SoundRAM()...)

	// native
	b := sound.New(r.Sound[:])
	b.OnAY = func(chip int, reg, val uint8) { got = append(got, w{chip, reg, val}) }
	// the reset writes were captured before OnAY was set: replay them
	got = append(got, w{1, 10, 0}, w{1, 9, 0}, w{1, 8, 0}, w{0, 10, 0}, w{0, 9, 0}, w{0, 8, 0}, w{0, 7, 0x38}, w{1, 7, 0x38})
	_ = got
	got = got[:0]
	b = sound.NewTrace(r.Sound[:], func(chip int, reg, val uint8) { got = append(got, w{chip, reg, val}) })
	next = 0
	for p := 1; p < passes+1; p++ {
		if next < len(cmds) && p-1 >= gap*(next+1) {
			b.Command(uint8(cmds[next]))
			next++
		}
		b.Tick()
	}
	goRAM = b.RAM()
	return
}

func main() {
	r, _ := assets.Load(os.Args[1])
	var list []int
	for _, s := range strings.Split(os.Args[2], ",") {
		if strings.Contains(s, "-") {
			p := strings.Split(s, "-")
			a, _ := strconv.Atoi(p[0])
			b, _ := strconv.Atoi(p[1])
			for i := a; i <= b; i++ {
				list = append(list, i)
			}
		} else {
			n, _ := strconv.ParseInt(s, 0, 32)
			list = append(list, int(n))
		}
	}
	passes := 500
	if len(os.Args) > 3 {
		passes, _ = strconv.Atoi(os.Args[3])
	}
	bad := 0
	if strings.HasPrefix(os.Args[2], "R") {
		seed, _ := strconv.Atoi(os.Args[2][1:])
		rnd := rand.New(rand.NewSource(int64(seed)))
		var cmds []int
		for i := 0; i < passes/9; i++ {
			c := rnd.Intn(0x22)
			switch rnd.Intn(8) {
			case 0:
				c |= 0x80
			case 1:
				c = 26 + rnd.Intn(2)
			}
			cmds = append(cmds, c)
		}
		ora, got, oraRAM, goRAM := run(r, cmds, 9, passes)
		n := len(ora)
		if len(got) < n {
			n = len(got)
		}
		for i := 0; i < n; i++ {
			if ora[i] != got[i] {
				fmt.Printf("seed %d: write %d differs: oracle %+v go %+v\n", seed, i, ora[i], got[i])
				fmt.Println("cmds:", cmds[:12])
				for j := i - 8; j < i+6 && j < n; j++ {
					if j >= 0 {
						fmt.Printf("  %d: oracle %+v go %+v\n", j, ora[j], got[j])
					}
				}
				os.Exit(1)
			}
		}
		ramDiff := 0
		for i := 0; i < 0x300; i++ {
			if oraRAM[i] != goRAM[i] {
				ramDiff++
			}
		}
		fmt.Printf("seed %d: writes oracle=%d go=%d ramDiff=%d OK\n", seed, len(ora), len(got), ramDiff)
		return
	}
	for _, c := range list {
		ora, got, oraRAM, goRAM := run(r, []int{c}, 4, passes)
		n := len(ora)
		if len(got) < n {
			n = len(got)
		}
		fail := false
		for i := 0; i < n; i++ {
			if ora[i] != got[i] {
				fmt.Printf("cmd %d: write %d differs: oracle %+v go %+v\n", c, i, ora[i], got[i])
				fail = true
				break
			}
		}
		d := len(ora) - len(got)
		ramDiff := 0
		for i := 0; i < 0x300; i++ {
			if oraRAM[i] != goRAM[i] {
				ramDiff++
			}
		}
		fmt.Printf("cmd %3d: writes oracle=%d go=%d fail=%v lenDiff=%d ramDiff=%d\n", c, len(ora), len(got), fail, d, ramDiff)
		if fail {
			bad++
		}
	}
	if bad > 0 {
		os.Exit(1)
	}
}
