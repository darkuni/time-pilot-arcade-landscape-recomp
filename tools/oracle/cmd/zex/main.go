// zex runs the zexdoc/zexall CP/M exercisers against the dev-only Z80 core.
package main

import (
	"fmt"
	"os"
	"timepilot/tools/oracle/z80"
)

type bus struct {
	mem [65536]uint8
}

func (b *bus) Read(a uint16) uint8     { return b.mem[a] }
func (b *bus) Write(a uint16, v uint8) { b.mem[a] = v }
func (b *bus) In(p uint16) uint8       { return 0xff }
func (b *bus) Out(p uint16, v uint8)   {}

func main() {
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	b := &bus{}
	copy(b.mem[0x100:], data)
	b.mem[5] = 0xc9 // RET at BDOS entry
	c := &z80.CPU{Bus: b}
	c.Reset()
	c.PC = 0x100
	c.SP = 0xf000
	for {
		if c.PC == 0 {
			fmt.Println("\n-- exit")
			return
		}
		if c.PC == 5 {
			switch c.C {
			case 2:
				fmt.Printf("%c", c.E)
			case 9:
				a := uint16(c.D)<<8 | uint16(c.E)
				for b.mem[a] != '$' {
					fmt.Printf("%c", b.mem[a])
					a++
				}
			}
		}
		c.Step()
	}
}
