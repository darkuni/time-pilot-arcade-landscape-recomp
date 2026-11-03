// Package assets loads the original Time Pilot ROM set (from the user's own
// timeplt.zip) and decodes graphics, palette and lookup PROMs.
package assets

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ROMs holds the raw ROM images.
type ROMs struct {
	Main    [0x6000]byte // tm1,tm2,tm3
	Sound   [0x1000]byte // tm7
	Chars   [0x2000]byte // tm6
	Sprites [0x4000]byte // tm4,tm5
	PalLo   [32]byte     // timeplt.b4
	PalHi   [32]byte     // timeplt.b5
	SprLUT  [256]byte    // timeplt.e9
	CharLUT [256]byte    // timeplt.e12
}

var layout = []struct {
	name string
	dst  func(r *ROMs) []byte
}{
	{"tm1", func(r *ROMs) []byte { return r.Main[0x0000:0x2000] }},
	{"tm2", func(r *ROMs) []byte { return r.Main[0x2000:0x4000] }},
	{"tm3", func(r *ROMs) []byte { return r.Main[0x4000:0x6000] }},
	{"tm7", func(r *ROMs) []byte { return r.Sound[:] }},
	{"tm6", func(r *ROMs) []byte { return r.Chars[:] }},
	{"tm4", func(r *ROMs) []byte { return r.Sprites[0x0000:0x2000] }},
	{"tm5", func(r *ROMs) []byte { return r.Sprites[0x2000:0x4000] }},
	{"timeplt.b4", func(r *ROMs) []byte { return r.PalLo[:] }},
	{"timeplt.b5", func(r *ROMs) []byte { return r.PalHi[:] }},
	{"timeplt.e9", func(r *ROMs) []byte { return r.SprLUT[:] }},
	{"timeplt.e12", func(r *ROMs) []byte { return r.CharLUT[:] }},
}

// Load reads the ROM set from a zip file, or from a directory of loose files.
func Load(path string) (*ROMs, error) {
	st, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	r := &ROMs{}
	if st.IsDir() {
		for _, l := range layout {
			b, err := os.ReadFile(filepath.Join(path, l.name))
			if err != nil {
				return nil, err
			}
			d := l.dst(r)
			if len(b) != len(d) {
				return nil, fmt.Errorf("%s: size %d, want %d", l.name, len(b), len(d))
			}
			copy(d, b)
		}
		return r, nil
	}
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	return loadZip(r, &zr.Reader, path)
}

// Embedded is a ROM zip compiled into the program (see embed.go, build tag
// "embedrom"); nil in the normal build.
var Embedded []byte

// LoadEmbedded reads the ROM set from the compiled-in zip.
func LoadEmbedded() (*ROMs, error) {
	zr, err := zip.NewReader(bytes.NewReader(Embedded), int64(len(Embedded)))
	if err != nil {
		return nil, err
	}
	return loadZip(&ROMs{}, zr, "embedded ROM set")
}

func loadZip(r *ROMs, zr *zip.Reader, path string) (*ROMs, error) {
	files := map[string]*zip.File{}
	for _, f := range zr.File {
		files[filepath.Base(f.Name)] = f
	}
	for _, l := range layout {
		f, ok := files[l.name]
		if !ok {
			return nil, fmt.Errorf("rom file %q missing from %s", l.name, path)
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		b, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, err
		}
		d := l.dst(r)
		if len(b) != len(d) {
			return nil, fmt.Errorf("%s: size %d, want %d", l.name, len(b), len(d))
		}
		copy(d, b)
	}
	return r, nil
}
