// probe: ad-hoc inspection runs of the reference machine.
package main

import (
	"fmt"
	"os"
	"timepilot/internal/assets"
	"timepilot/tools/oracle/machine"
)

func main() {
	r, _ := assets.Load(os.Args[1])
	m := machine.New(r)
	watch := false
	m.WriteWatch = func(a uint16, v uint8, pc uint16) {
		if watch && ((a >= 0xb410 && a < 0xb440) || (a >= 0xb010 && a < 0xb040)) && (a&1 == 1 || a < 0xb400) {
			fmt.Printf("f%d L%3d pc=%04x W %04x=%02x\n", m.Frame, m.Line, pc, a, v)
		}
	}
	for f := 0; f < 1100; f++ {
		m.In0 = 0xff
		if f >= 300 && f < 304 {
			m.In0 &^= 1
		}
		if f >= 400 && f < 404 {
			m.In0 &^= 8
		}
		if f >= 1000 && f < 1003 {
			watch = true
			fmt.Printf("--- frame %d a9ab=%02x a9ac=%02x\n", f, m.WorkRAM()[0x1ab], m.WorkRAM()[0x1ac])
		} else {
			watch = false
		}
		m.RunFrame()
	}
}
