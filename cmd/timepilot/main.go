// Time Pilot — a native Go rewrite of Konami's 1982 arcade game.
//
// The game needs the original ROM set (timeplt.zip) for its graphics, level
// tables and sound data; none of Konami's data is bundled with this program.
//
// Controls:  arrows move, Z fire, 5 insert coin, 1 / 2 start one / two player
// game, L switch landscape / portrait layout, P switch Time Pilot+ (side-panel
// statistics) on or off, F3 input/heading overlay, F11 toggle full screen, Esc quit.
//
// Gamepad: D-pad/left stick move, A or B fire, Select coin, Start 1 player,
// X or Y 2 players, right bumper change layout, click the right stick to
// switch Time Pilot+ on or off, hold Select+Start to quit.
//
// Option -classic starts with the original portrait picture and no statistics.
package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"timepilot/internal/assets"
	"timepilot/internal/game"
	"timepilot/internal/sound"
	"timepilot/internal/stats"
	"timepilot/internal/video"
)

type app struct {
	pad        padTracker
	coinFrames int
	g          *game.Game
	r          *video.Renderer
	rgba       *image.RGBA
	screen     *ebiten.Image
	land       *image.RGBA
	hud        *image.RGBA
	landIm     *ebiten.Image
	debug      bool
	in1        uint8
	ph         [2]*image.RGBA
	// landscape is the default layout; L switches to the original portrait one
	landscape bool
	synth     *sound.Synth
	sndIRQ    int
	// plus shows the Time Pilot+ statistics panels in the landscape layout
	plus      bool
	tr        *stats.Tracker
	trFrames  int
	notice    int // frames left to show the Plus message
	noticeMsg string
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
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		a.togglePlus()
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
	pd := a.pad.poll()
	if pd.quit {
		return ebiten.Termination
	}
	if pd.view {
		a.landscape = !a.landscape
		w, h := a.size()
		if !ebiten.IsFullscreen() {
			ebiten.SetWindowSize(w*3, h*3)
		}
	}
	if pd.plus {
		a.togglePlus()
	}
	if pd.coin || a.coinFrames > 0 {
		in0 &^= 0x01
		if pd.coin {
			a.coinFrames = 3
		} else {
			a.coinFrames--
		}
	}
	if pd.start1 {
		in0 &^= 0x08
	}
	if pd.start2 {
		in0 &^= 0x10
	}
	if pd.left {
		in1 &^= 0x01
	}
	if pd.right {
		in1 &^= 0x02
	}
	if pd.up {
		in1 &^= 0x04
	}
	if pd.down {
		in1 &^= 0x08
	}
	if pd.fire {
		in1 &^= 0x10
	}
	a.g.In0, a.g.In1, a.g.In2 = in0, in1, in1
	a.in1 = in1
	a.g.Frame()
	a.tr.Frame()
	a.trFrames++
	if a.trFrames >= 600 && a.tr.Dirty {
		a.trFrames = 0
		a.saveStats()
	}
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
		if a.plus {
			stats.DrawPanels(a.land, a.tr)
		}
		a.drawNotice(a.land, 4, 207)
		a.landIm.WritePixels(a.land.Pix)
		screen.DrawImage(a.landIm, nil)
		return
	}
	a.drawNotice(a.rgba, 2, 2)
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

// togglePlus switches the Time Pilot+ panels on or off and remembers it.
func (a *app) togglePlus() {
	a.plus = !a.plus
	a.notice = 150
	a.noticeMsg = "TIME PILOT+: OFF (ORIGINAL)"
	if a.plus {
		a.noticeMsg = "TIME PILOT+: ON"
	}
	savePlus(a.plus)
}

// drawNotice shows the short Plus message for a couple of seconds.
func (a *app) drawNotice(dst *image.RGBA, x, y int) {
	if a.notice <= 0 {
		return
	}
	a.notice--
	stats.DrawText(dst, x+1, y+1, a.noticeMsg, color.RGBA{0, 0, 0, 255}, 1)
	stats.DrawText(dst, x, y, a.noticeMsg, color.RGBA{255, 255, 73, 255}, 1)
}

// saveStats writes the all-time statistics.
func (a *app) saveStats() {
	if f := configFile("stats.json"); f != "" {
		_ = os.MkdirAll(filepath.Dir(f), 0o755)
		_ = a.tr.Save(f)
	}
}

func configFile(name string) string {
	d, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(d, "timepilot", name)
}

// Time Pilot+ (the statistics panels) is on by default; the choice is remembered.
func loadPlus() bool {
	if f := configFile("plus.txt"); f != "" {
		if b, err := os.ReadFile(f); err == nil {
			return strings.TrimSpace(string(b)) != "classic"
		}
	}
	return true
}

func savePlus(on bool) {
	f := configFile("plus.txt")
	if f == "" {
		return
	}
	v := "classic"
	if on {
		v = "plus"
	}
	_ = os.MkdirAll(filepath.Dir(f), 0o755)
	_ = os.WriteFile(f, []byte(v+"\n"), 0o644)
}

func findROM() string {
	for _, arg := range os.Args[1:] {
		if !strings.HasPrefix(arg, "-") {
			return arg
		}
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
	classic := false
	for _, arg := range os.Args[1:] {
		if arg == "-classic" || arg == "--classic" {
			classic = true
		}
	}
	a := &app{
		g:      g,
		r:      video.New(roms),
		rgba:   image.NewRGBA(image.Rect(0, 0, video.ScreenW, video.ScreenH)),
		screen: ebiten.NewImage(video.ScreenW, video.ScreenH),
		hud:    image.NewRGBA(image.Rect(0, 0, video.ScreenW, video.ScreenH)),
		land:   image.NewRGBA(image.Rect(0, 0, video.LandscapeW, video.LandscapeH)),
		landIm: ebiten.NewImage(video.LandscapeW, video.LandscapeH),

		landscape: !classic,
		plus:      loadPlus() && !classic,
	}
	for i := range a.ph {
		a.ph[i] = image.NewRGBA(image.Rect(0, 0, video.ScreenW, video.ScreenH))
	}
	a.tr = stats.New(g)
	a.tr.Icons = stats.EnemyIcons(g, a.r)
	if f := configFile("stats.json"); f != "" {
		a.tr.Load(f)
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
	ww, wh := a.size()
	ebiten.SetWindowSize(ww*3, wh*3)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetTPS(60)
	if err := ebiten.RunGame(a); err != nil {
		log.Fatal(err)
	}
	a.saveStats()
}
