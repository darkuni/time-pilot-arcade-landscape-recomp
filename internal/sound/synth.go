package sound

import (
	"math"
	"sync"
)

// ayVolume is the AY-3-8910 amplitude curve (4-bit level -> linear gain).
var ayVolume = [16]float64{
	0, 0.0100, 0.0145, 0.0211, 0.0307, 0.0455, 0.0644, 0.1073,
	0.1288, 0.2047, 0.2913, 0.3722, 0.4921, 0.6349, 0.8175, 1.0,
}

// Voice -> bit position of its filter selection in the filter latch.
var filterShift = [6]uint{6, 8, 10, 0, 2, 4}

// CapScale scales the filter capacitors. Tuned against a recording of the
// original: the nominal values (1.0) sounded too muffled.
var CapScale = 0.3

const (
	filterR      = 1000.0 * 5100.0 / (1000.0 + 5100.0) // ohms
	channelGain  = 0.60
	capBitLow    = 220e-9 // latch bit 0
	capBitHigh   = 47e-9  // latch bit 1
	toneClockDiv = 8.0
)

type toneGen struct {
	counter float64 // AY ticks until the next edge
	level   float64 // 0 or 1
}

// Synth renders the board's output as audio.  It is safe to call Command
// from one goroutine while another calls Render.
type Synth struct {
	mu       sync.Mutex
	b        *Board
	rate     float64
	tickAcc  float64
	gen      [6]toneGen
	lp       [6]float64 // low-pass states
	hpX, hpY float64
}

// NewSynth creates a synthesiser producing samples at the given rate.
func NewSynth(rom []byte, rate int) *Synth {
	s := &Synth{b: New(rom), rate: float64(rate)}
	for i := range s.gen {
		s.gen[i].level = 1
	}
	return s
}

// Command forwards a command byte from the main board.
func (s *Synth) Command(v uint8) {
	s.mu.Lock()
	s.b.Command(v)
	s.mu.Unlock()
}

// Render fills out with mono samples in the range [-1, 1].
func (s *Synth) Render(out []float32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cyclesPer := ClockHz / s.rate
	ticksPer := ClockHz / toneClockDiv / s.rate // AY tone-clock ticks per sample
	b := s.b
	for i := range out {
		s.tickAcc += cyclesPer
		for s.tickAcc >= TickCycles {
			s.tickAcc -= TickCycles
			b.Tick()
		}
		var mix float64
		for ch := 0; ch < 6; ch++ {
			chip, c := ch/3, ch%3
			reg := &b.Reg[chip]
			period := float64(uint16(reg[2*c]) | uint16(reg[2*c+1]&15)<<8)
			if period < 1 {
				period = 1
			}
			toneOn := reg[7]>>uint(c)&1 == 0
			amp := reg[8+c]
			gain := ayVolume[amp&15]
			if amp&0x10 != 0 { // envelope mode is never used by the ROM
				gain = ayVolume[15]
			}
			// average level over this sample
			var avg float64
			if toneOn {
				g := &s.gen[ch]
				remain := ticksPer
				acc := 0.0
				for remain > 0 {
					if g.counter <= 0 {
						g.counter += period
						g.level = 1 - g.level
					}
					step := math.Min(remain, g.counter)
					acc += g.level * step
					g.counter -= step
					remain -= step
				}
				avg = acc / ticksPer
			} else {
				avg = 1
			}
			x := avg * gain
			// per-voice RC filter selected through the filter latch
			voice := ch
			if chip == 0 {
				voice = c
			} else {
				voice = 3 + c
			}
			pair := b.Filter >> filterShift[voice] & 3
			var cap float64
			if pair&1 != 0 {
				cap += capBitLow
			}
			if pair&2 != 0 {
				cap += capBitHigh
			}
			cap *= CapScale
			if cap > 0 {
				dt := 1 / s.rate
				a := dt / (filterR*cap + dt)
				s.lp[ch] += a * (x - s.lp[ch])
				x = s.lp[ch]
			} else {
				s.lp[ch] = x
			}
			mix += x
		}
		mix *= channelGain
		// AC coupling
		y := mix - s.hpX + 0.995*s.hpY
		s.hpX, s.hpY = mix, y
		if y > 1 {
			y = 1
		} else if y < -1 {
			y = -1
		}
		out[i] = float32(y)
	}
}
