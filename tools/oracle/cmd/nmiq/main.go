// nmiq: trace the (a9ab,a9ac) frame-state sequence at every NMI entry.
package main

import (
	"fmt"
	"math/rand"
	"os"
	"timepilot/internal/assets"
	"timepilot/tools/oracle/machine"
)

func main() {
	r, _ := assets.Load(os.Args[1])
	m := machine.New(r)
	last := [2]byte{0xee, 0xee}
	start := 0
	m.MainTrace = func(pc uint16) {
		if pc == 0x66 {
			ram := m.WorkRAM()
			cur := [2]byte{ram[0x1ab], ram[0x1ac]}
			if cur != last {
				if last[0] != 0xee {
					fmt.Printf("  frames %d..%d state %02x/%02x\n", start, m.Frame-1, last[0], last[1])
				}
				last = cur
				start = m.Frame
			}
		}
	}
	rnd := rand.New(rand.NewSource(1))
	for f := 0; f < 3000; f++ {
		m.In0, m.In1, m.In2 = 0xff, 0xff, 0xff
		if f >= 300 && f < 304 {
			m.In0 &^= 1
		}
		if f >= 1400 && f < 1404 {
			m.In0 &^= 8
		}
		if f > 1450 {
			if rnd.Intn(2) == 0 { m.In1 &^= 1 }
			if rnd.Intn(2) == 0 { m.In1 &^= 2 }
			if rnd.Intn(3) == 0 { m.In1 &^= 4 }
			if rnd.Intn(3) == 0 { m.In1 &^= 8 }
			if rnd.Intn(2) == 0 { m.In0 &^= 0x10 }
		}
		m.RunFrame()
	}
}
