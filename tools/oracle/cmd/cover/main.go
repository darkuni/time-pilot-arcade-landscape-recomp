// cover plays the original ROM with pseudo-random input and records code coverage.
package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"timepilot/internal/assets"
	"timepilot/tools/oracle/machine"
)

func main() {
	r, _ := assets.Load(os.Args[1])
	frames, _ := strconv.Atoi(os.Args[3])
	seed, _ := strconv.Atoi(os.Args[4])
	rng := rand.New(rand.NewSource(int64(seed)))
	m := machine.New(r)
	m.CovOn = true
	var dir uint8 = 0xff
	var fire bool
	for f := 0; f < frames; f++ {
		m.In0 = 0xff
		if f%2400 >= 300 && f%2400 < 304 {
			m.In0 &^= 1
		}
		if f%2400 >= 400 && f%2400 < 404 {
			if (f/2400)%2 == 0 {
				m.In0 &^= 8
			} else {
				m.In0 &^= 16
			}
		}
		if f%20 == 0 {
			dir = 0xff
			for b := uint(0); b < 4; b++ {
				if rng.Intn(3) == 0 {
					dir &^= 1 << b
				}
			}
			fire = rng.Intn(2) == 0
		}
		m.In1, m.In2 = dir, dir
		if fire && f%4 < 2 {
			m.In1 &^= 16
			m.In2 &^= 16
		}
		m.RunFrame()
	}
	exe, dat := 0, 0
	out := make([]byte, 0x6000)
	copy(out, m.Cov[:])
	for _, c := range out {
		if c&1 != 0 {
			exe++
		}
		if c&2 != 0 {
			dat++
		}
	}
	os.WriteFile(os.Args[2], out, 0644)
	fmt.Printf("frames=%d exec-bytes=%d data-bytes=%d\n", frames, exe, dat)
}
