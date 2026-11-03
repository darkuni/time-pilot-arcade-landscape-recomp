// Time Pilot — a native Go rewrite of Konami's 1982 arcade game.
//
// The game needs the original ROM set (timeplt.zip) for its graphics, level
// tables and sound data; none of Konami's data is bundled with this program.
//
// Controls:  arrows move, Z fire, 5 insert coin, 1 / 2 start one / two player
// game, L switch landscape / portrait layout, F3 input/heading overlay, F11 toggle full screen, Esc quit.
package main

import (
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"timepilot/internal/assets"
	"timepilot/internal/game"
	"timepilot/internal/sound"
	"timepilot/internal/video"
)

type app struct {
	g      *game.Game
	r      *video.Renderer
	rgba   *image.RGBA
	screen *ebiten.Image
	land   *image.RGBA
	hud    *image.RGBA
	landIm *ebiten.Image
	debug  bool
	in1    uint8
	ph     [2]*image.RGBA
	// landscape is the default layout; L switches to the original portrait one
	landscape bool
	synth     *sound.Synth
	sndIRQ    int
}

const sampleRate = 48000

// pcm streams the synthesiser as 16-bit stereo little-endian samples.
type pcm struct {
	s   *sound.Synth
	buf []float32
}

func (p *pcm) Read(out []byte) (int, error) {
	n := len(out) / 4
	if cap(p.buf) < n {
		p.buf = make([]float32, n)
	}
	p.buf = p.buf[:n]
	p.s.Render(p.buf)
	for i, v := range p.buf {
		x := int16(v * 32767)
		out[4*i], out[4*i+1] = byte(x), byte(uint16(x)>>8)
		out[4*i+2], out[4*i+3] = out[4*i], out[4*i+1]
	}
	return n * 4, nil
}

func (a *app) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyL) {
		a.landscape = !a.landscape
		w, h := a.size()
		if !ebiten.IsFullscreen() {
			ebiten.SetWindowSize(w*3, h*3)
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF11) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF3) {
		a.debug = !a.debug
	}
	in0, in1 := uint8(0xff), uint8(0xff)
	down := ebiten.IsKeyPressed
	if down(ebiten.Key5) {
		in0 &^= 0x01 // coin 1
	}
	if down(ebiten.Key1) {
		in0 &^= 0x08 // 1 player start
	}
	if down(ebiten.Key2) {
		in0 &^= 0x10 // 2 player start
	}
	if down(ebiten.KeyLeft) {
		in1 &^= 0x01 // port bit 0 is "left" on this board
	}
	if down(ebiten.KeyRight) {
		in1 &^= 0x02
	}
	if down(ebiten.KeyUp) {
		in1 &^= 0x04
	}
	if down(ebiten.KeyDown) {
		in1 &^= 0x08
	}
	if down(ebiten.KeyZ) {
		in1 &^= 0x10 // fire
	}
	a.g.In0, a.g.In1, a.g.In2 = in0, in1, in1
	a.in1 = in1
	a.g.Frame()
	if a.g.SoundIRQ != a.sndIRQ {
		a.sndIRQ = a.g.SoundIRQ
		a.synth.Command(a.g.SoundLatch)
	}
	return nil
}

func (a *app) Draw(screen *ebiten.Image) {
	defer func() {
		if a.debug {
			k := func(bit uint8, n string) string {
				if a.in1&bit == 0 {
					return n
				}
				return "."
			}
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%s%s%s%s %s  heading %02x",
				k(1, "L"), k(2, "R"), k(4, "U"), k(8, "D"), k(0x10, "FIRE"), a.g.Peek(0xa802, 1)[0]), 2, 2)
		}
	}()
	a.r.Draw(a.g, a.rgba)
	if a.landscape {
		a.r.DrawHUD(a.g, a.hud)
		var p [2]*image.RGBA
		for i := range p {
			if a.r.DrawPlayerHUD(a.g, i, a.ph[i]) {
				p[i] = a.ph[i]
			}
		}
		video.Landscape(a.land, a.rgba, a.hud, p[0], p[1])
		a.landIm.WritePixels(a.land.Pix)
		screen.DrawImage(a.landIm, nil)
		return
	}
	a.screen.WritePixels(a.rgba.Pix)
	screen.DrawImage(a.screen, nil)
}

func (a *app) size() (int, int) {
	if a.landscape {
		return video.LandscapeW, video.LandscapeH
	}
	return video.ScreenW, video.ScreenH
}

func (a *app) Layout(_, _ int) (int, int) { return a.size() }

func findROM() string {
	if len(os.Args) > 1 {
		return os.Args[1]
	}
	cands := []string{"timeplt.zip"}
	if exe, err := os.Executable(); err == nil {
		d := filepath.Dir(exe)
		cands = append(cands, filepath.Join(d, "timeplt.zip"))
		// macOS app bundle: Contents/MacOS/timepilot -> look in Resources and
		// in the folder that holds the .app
		cands = append(cands,
			filepath.Join(d, "..", "Resources", "timeplt.zip"),
			filepath.Join(d, "..", "..", "..", "timeplt.zip"))
		if home, err := os.UserHomeDir(); err == nil {
			cands = append(cands, filepath.Join(home, "timeplt.zip"))
		}
	}
	for _, c := range cands {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

func main() {
	path := findROM()
	if path == "" && assets.Embedded == nil {
		log.Fatal("timeplt.zip not found: put it next to timepilot.exe or pass its path as the first argument")
	}
	var roms *assets.ROMs
	var err error
	if path == "" {
		roms, err = assets.LoadEmbedded()
	} else {
		roms, err = assets.Load(path)
	}
	if err != nil {
		log.Fatal(fmt.Errorf("loading %s: %w", path, err))
	}
	g := game.New(roms.Main[:])
	g.Boot()
	a := &app{
		g:      g,
		r:      video.New(roms),
		rgba:   image.NewRGBA(image.Rect(0, 0, video.ScreenW, video.ScreenH)),
		screen: ebiten.NewImage(video.ScreenW, video.ScreenH),
		hud:    image.NewRGBA(image.Rect(0, 0, video.ScreenW, video.ScreenH)),
		land:   image.NewRGBA(image.Rect(0, 0, video.LandscapeW, video.LandscapeH)),
		landIm: ebiten.NewImage(video.LandscapeW, video.LandscapeH),

		landscape: true,
	}
	for i := range a.ph {
		a.ph[i] = image.NewRGBA(image.Rect(0, 0, video.ScreenW, video.ScreenH))
	}
	a.synth = sound.NewSynth(roms.Sound[:], sampleRate)
	a.sndIRQ = g.SoundIRQ
	a.synth.Command(g.SoundLatch)
	ctx := audio.NewContext(sampleRate)
	player, err := ctx.NewPlayer(&pcm{s: a.synth})
	if err != nil {
		log.Fatal(err)
	}
	player.SetBufferSize(60 * time.Millisecond)
	player.Play()
	ebiten.SetWindowTitle("Time Pilot")
	ebiten.SetWindowSize(video.LandscapeW*3, video.LandscapeH*3)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetTPS(60)
	if err := ebiten.RunGame(a); err != nil {
		log.Fatal(err)
	}
}
