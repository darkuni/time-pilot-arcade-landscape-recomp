# Time Pilot — native Go rewrite

A from-scratch rewrite of Konami's 1982 *Time Pilot* in Go. It is **not an emulator**: the
game logic (main CPU program) and the sound board's sequencer were analysed and re-implemented
routine by routine as ordinary Go code. The original ROM image is only used as *data* (graphics,
palettes, level tables, music/effect tables), loaded at run time from your own `timeplt.zip`.

## Running

Put `timeplt.zip` next to `timepilot.exe` (or pass its path as the first argument) and run it.

| Key | Action |
|---|---|
| Arrow keys | Move |
| Z | Fire |
| 5 | Insert coin |
| 1 / 2 | Start 1 / 2 player game |
| L | Switch landscape (default) / portrait layout |
| F11 | Full screen |
| Esc | Quit |

## Layout

| Path | Purpose |
|---|---|
| `cmd/timepilot` | Ebitengine window, keyboard, audio output |
| `internal/game` | Game logic rewrite (boot, attract, play, enemies, bosses, collisions, scoring, stage clear, name entry...) |
| `internal/sound` | Sound sequencer rewrite (25 effects, 2 songs) + AY-3-8910 synthesiser with the board's RC filters |
| `internal/video` | Tile / sprite renderer (palette + graphics decoded from ROM) |
| `internal/assets` | ROM loader (zip or folder), graphics and palette decoding |
| `tools/oracle` | **Development-only** reference machine (a Z80 interpreter) used to verify the rewrite |
| `cmd/shot` | Headless PNG screenshot tool |

## How correctness was established

The game state keeps the original memory layout (so a later refactor can be proven equivalent).
Two lockstep harnesses compare the rewrite with the reference machine:

* `tools/oracle/cmd/lockstep` — every frame, all work RAM, video RAM and colour RAM are compared
  byte for byte. Passed 40–60 k frames of randomised play, including cold boot (`COLD=1`),
  all five stages, boss fights, forced stage clears, 2-player games, death, name entry, game over.
* `tools/oracle/cmd/sndcheck` — every AY-3-8910 register write of the sound board is compared:
  each of the 32 commands individually, and 30+ random command mixes of 6000 sequencer passes.

Build and run the checks:

```
go build -o lockstep ./tools/oracle/cmd/lockstep
COLD=1 ./lockstep path/to/rom-dir 40000          # BOSS=1500 STAGE=4 SEED=n P2=1 STAGEEND=2500 are options
go build -o sndcheck ./tools/oracle/cmd/sndcheck
./sndcheck path/to/rom-dir R7 6000
```
(`path/to/rom-dir` is a folder with the unzipped ROM files.)

## Build

```
go build -ldflags "-H=windowsgui -s -w" -o timepilot.exe ./cmd/timepilot      # Windows
```
Pure Go (no cgo); cross-compiles to Windows, macOS and Linux (Linux needs the usual X11/ALSA dev libs).

## Known simplifications

* The sound board's analogue stage (RC filters) is modelled as one-pole low-pass filters with
  component values taken from MAME's description; exact tone may differ slightly from hardware.
* The sequencer runs one pass per 5120 sound-CPU cycles, as the timer chip does.
* Hardware beam-sync sprite "parking" is not reproduced; all sprites are drawn at true position.
* DIP switches are fixed at the defaults (3 lives, 1 coin / 1 credit, upright).

## Roadmap ideas

Replace the address-based state with typed structs (the lockstep harness proves equivalence),
configurable DIPs / difficulty, gamepad, scaling filters, save of hi-scores, new stages.

## macOS

Ebitengine needs cgo on macOS, so the Mac build cannot be cross-compiled from
Windows/Linux. On a Mac with Go 1.24+ and the Xcode command line tools run
`./scripts/build-mac.sh`; it makes a universal, ad-hoc-signed `Time Pilot.app`.
Put `timeplt.zip` next to the app (or in your home folder). First launch:
right-click the app, Open. (`.github/workflows/release.yml` does the same on a
GitHub macOS runner.)

## Releases

Publishing a GitHub release runs `.github/workflows/release.yml`, which builds
`timepilot.exe` (Windows) and `TimePilot-mac.zip` (universal macOS app) and attaches
them to the release. Neither contains the ROM: put your own `timeplt.zip` next to the game.
