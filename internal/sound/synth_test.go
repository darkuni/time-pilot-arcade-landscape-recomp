package sound

import (
	"math"
	"os"
	"testing"

	"timepilot/internal/assets"
)

func TestSynthRenders(t *testing.T) {
	r, err := assets.Load("/home/claude/tp/rom")
	if err != nil {
		t.Skip(err)
	}
	for _, cmd := range []uint8{1, 19, 26, 27} {
		s := NewSynth(r.Sound[:], 48000)
		s.Command(cmd)
		out := make([]float32, 48000*3)
		s.Render(out)
		var peak, e float64
		for _, v := range out {
			if math.IsNaN(float64(v)) {
				t.Fatal("nan")
			}
			peak = math.Max(peak, math.Abs(float64(v)))
			e += float64(v) * float64(v)
		}
		t.Logf("cmd %d peak %.3f rms %.4f", cmd, peak, math.Sqrt(e/float64(len(out))))
		if peak == 0 {
			t.Errorf("cmd %d silent", cmd)
		}
		if cmd == 27 {
			f, _ := os.Create("/tmp/music.raw")
			b := make([]byte, 0, len(out)*2)
			for _, v := range out {
				x := int16(v * 32767)
				b = append(b, byte(x), byte(uint16(x)>>8))
			}
			f.Write(b)
			f.Close()
		}
	}
}
