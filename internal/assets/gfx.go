package assets

import "math"

// RGB is an 8-bit-per-channel color.
type RGB struct{ R, G, B uint8 }

// bit returns bit n of b.
func bit(b byte, n uint) int { return int(b>>n) & 1 }

// resistorWeights replicates the classic resistor-ladder weight computation
// (R-2R-less parallel ladder with a pulldown), auto-scaled so that all bits on = 255.
func resistorWeights(res []float64, pulldown float64) []float64 {
	n := len(res)
	w := make([]float64, n)
	sum := 0.0
	for i := 0; i < n; i++ {
		// resistor i to Vcc, others to ground, plus pulldown
		g := 0.0
		for j := 0; j < n; j++ {
			if j != i {
				g += 1 / res[j]
			}
		}
		if pulldown > 0 {
			g += 1 / pulldown
		}
		rg := 1 / g
		w[i] = rg / (res[i] + rg)
		sum += w[i]
	}
	for i := range w {
		w[i] = w[i] * 255 / sum
	}
	return w
}

// Palette decodes the 32-entry hardware palette from the two color PROMs.
func (r *ROMs) Palette() [32]RGB {
	w := resistorWeights([]float64{1000, 470, 220, 470, 220}, 470)
	comb := func(bits ...int) uint8 {
		v := 0.0
		for i, b := range bits {
			if b != 0 {
				v += w[i]
			}
		}
		return uint8(math.Round(v))
	}
	var pal [32]RGB
	for i := 0; i < 32; i++ {
		hi, lo := r.PalHi[i], r.PalLo[i]
		pal[i] = RGB{
			R: comb(bit(hi, 1), bit(hi, 2), bit(hi, 3), bit(hi, 4), bit(hi, 5)),
			G: comb(bit(hi, 6), bit(hi, 7), bit(lo, 0), bit(lo, 1), bit(lo, 2)),
			B: comb(bit(lo, 3), bit(lo, 4), bit(lo, 5), bit(lo, 6), bit(lo, 7)),
		}
	}
	return pal
}

// CharTile is an 8x8 2bpp tile, pixel values 0..3.
type CharTile [8][8]uint8

// DecodeChars decodes the 512 character tiles from tm6.
func (r *ROMs) DecodeChars() [512]CharTile {
	var out [512]CharTile
	for c := 0; c < 512; c++ {
		base := c * 16
		for y := 0; y < 8; y++ {
			for x := 0; x < 8; x++ {
				b := r.Chars[base+y+(x/4)*8]
				xx := uint(x & 3)
				out[c][y][x] = uint8(bit(b, 3-xx)<<1 | bit(b, 7-xx))
			}
		}
	}
	return out
}

// SpriteTile is a 16x16 2bpp sprite, pixel values 0..3.
type SpriteTile [16][16]uint8

// DecodeSprites decodes the 256 sprite tiles from tm4/tm5.
func (r *ROMs) DecodeSprites() [256]SpriteTile {
	var out [256]SpriteTile
	for s := 0; s < 256; s++ {
		for y := 0; y < 16; y++ {
			for x := 0; x < 16; x++ {
				off := s*64 + (y & 7) + (y/8)*32 + (x/4)*8
				b := r.Sprites[off]
				xx := uint(x & 3)
				out[s][y][x] = uint8(bit(b, 3-xx)<<1 | bit(b, 7-xx))
			}
		}
	}
	return out
}
