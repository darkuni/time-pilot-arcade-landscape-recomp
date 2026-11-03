package main

import (
	"fmt"
	"os"
	"strconv"
	"timepilot/internal/assets"
	"timepilot/tools/oracle/machine"
)

func main() {
	r, _ := assets.Load(os.Args[1])
	m := machine.New(r)
	m.LogAY = true
	var passes []uint64
	var last uint64
	m.SoundTrace = func(pc uint16) {
		if pc == 0x00c8 {
			passes = append(passes, m.Sound.Cycles)
			last = m.Sound.Cycles
		}
	}
	_ = last
	m.SoundRun(100000)
	_, _ = strconv.Atoi(os.Args[2])
	n0 := len(passes)
	a0 := len(m.AYLog)
	fmt.Println("init passes", n0, "ay writes", a0)
	for i := 1; i < len(passes); i++ {
		fmt.Printf("pass at %d (+%d)\n", passes[i], passes[i]-passes[i-1])
		if i > 30 {
			break
		}
	}
}
