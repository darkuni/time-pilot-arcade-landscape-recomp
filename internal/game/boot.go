package game

// Boot performs the cold start of the board: everything the original program
// does from the reset vector until it first enables the vertical-blank
// interrupt (0x07b1 -> 0x0069 -> 0x5866 -> 0x2511 -> 0x52aa -> 0x00b1 ->
// 0x32eb).  The long busy-wait and the ROM checksum loops of the original are
// pure delays / anti-tamper checks and are not reproduced.
func (g *Game) Boot() {
	// work RAM and sprite RAM cleared
	g.ram = [0x800]byte{}
	g.spr = [0x100]byte{}
	g.spr2 = [0x100]byte{}
	// colour RAM = 0x10, video RAM = 0xf1 (blank)
	for i := range g.cram {
		g.cram[i] = 0x10
	}
	for i := range g.vram {
		g.vram[i] = 0xf1
	}
	// draw-command ring buffer empty
	for a := uint16(0xac00); a < 0xac40; a++ {
		g.wb(a, 0xff)
	}
	// 0x4b67 and 0x4ba5: initial data blocks
	for i := uint16(0); i < 0x11; i++ {
		g.wb(0xab30+i, g.rb(0x4b84+i))
	}
	for i := uint16(0); i < 0x28; i++ {
		g.wb(0xab08+i, g.rb(0x4bb1+i))
	}
	// 0x526a: tile fix-up lists empty
	g.resetFixups()

	// 0x52aa: read the DIP switches
	g.wb(0xa98d, g.k(0x08c9))
	g.wb(0xa9cd, g.k(0x0874))
	sw0 := ^g.Dsw0
	g.wb(0xa9b1, sw0)
	// 0x4acc: coin settings
	if sw0&15 == 15 {
		g.wb(0xa9c0, 0xff)
	}
	g.wb(0xa9c9, g.rb(0x4b95+uint16(sw0&15)))
	if sw0>>4 == 15 {
		g.wb(0xa9c0, 0xff)
	}
	g.wb(0xa9cc, g.rb(0x4b95+uint16(sw0>>4)))

	c := ^g.Dsw1
	lives := c&3 + 3
	if lives == 6 {
		lives = 0xff
	}
	g.wb(0xa9c1, lives)
	ror := func(v uint8, n uint) uint8 { return v>>n | v<<(8-n) }
	c = ror(c, 2)
	g.wb(0xa9c2, c&1)
	c = ror(c, 1)
	g.wb(0xa9c3, c&1)
	c = ror(c, 1)
	g.wb(0xa9c4, c&7)
	c = ror(c, 3)
	g.wb(0xa9c6, c&1)

	// 0x00b1: the opening background pattern
	hl := uint16(0xa420)
	for row := 0; row < 14; row++ {
		hl += 0x20
		for col := 0; col < 16; col++ {
			g.wb(hl, 0x56)
			g.wb(hl+1, 0x83)
			g.wb(hl+0x20, 0xc7)
			g.wb(hl+0x21, 0xef)
			hl += 2
		}
	}
	g.wb(0xa9eb, 0)

	// 0x55f8: silence the sound board, pulse its interrupt line
	g.ioWrite(0xc000, 0)
	g.ioWrite(0xc304, 1)
	g.ioWrite(0xc304, 0)
	// 0x00a8: enable the vertical-blank interrupt as the ROM dictates
	g.ioWrite(0xc300, g.k(0x4c87))
}
