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
	m.WriteWatch = func(a uint16, v uint8, pc uint16) {
		if a == 0xa9eb && m.Frame >= 233 && m.Frame < 237 {
			fmt.Printf("frame %d line %d pc=%04x write a9eb=%02x\n", m.Frame, m.Line, pc, v)
		}
	}
	for i := 0; i < 237; i++ {
		m.RunFrame()
	}
}
