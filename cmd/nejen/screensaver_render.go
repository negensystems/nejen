package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"golang.org/x/sys/unix"
)

func init() {
	registerCommand("screensaver render", runScreensaverRender)
}

// getTerminalSize asks the terminal directly via TIOCGWINSZ. It reads stdout
// (and then stderr) rather than stdin, because `nejen screensaver` starts this
// renderer as a child with only Stdout/Stderr wired to the pty -- stdin is
// /dev/null there, so anything asking stdin gets no window size and the scene
// would silently fall back to 80x24 and draw itself into the top-left corner.
func getTerminalSize() (int, int) {
	for _, f := range []*os.File{os.Stdout, os.Stderr} {
		ws, err := unix.IoctlGetWinsize(int(f.Fd()), unix.TIOCGWINSZ)
		if err == nil && ws.Col > 0 && ws.Row > 0 {
			return int(ws.Col), int(ws.Row)
		}
	}
	return 80, 24
}

type CanvasCell struct {
	Ch   string
	RGB  [3]int
	Bold bool
}

type Canvas struct {
	W, H  int
	Cells [][]CanvasCell
}

func NewCanvas(w, h int) *Canvas {
	c := &Canvas{W: w, H: h, Cells: make([][]CanvasCell, h)}
	for i := range c.Cells {
		c.Cells[i] = make([]CanvasCell, w)
	}
	return c
}

func (c *Canvas) Put(x, y int, ch string, rgb [3]int, bold bool) {
	if x >= 0 && x < c.W && y >= 0 && y < c.H {
		c.Cells[y][x] = CanvasCell{Ch: ch, RGB: rgb, Bold: bold}
	}
}

func (c *Canvas) Text(x, y int, s string, rgb [3]int, bold bool) {
	runes := []rune(s)
	for i, r := range runes {
		c.Put(x+i, y, string(r), rgb, bold)
	}
}

func (c *Canvas) Flush() {
	var sb strings.Builder
	sb.WriteString("\033[H")
	var lastRGB [3]int
	lastRGBSet := false
	var lastBold bool
	lastBoldSet := false

	for i, row := range c.Cells {
		if i > 0 {
			sb.WriteString("\033[E")
		}
		for _, cell := range row {
			if cell.Ch == "" {
				sb.WriteString(" ")
				continue
			}
			if !lastBoldSet || cell.Bold != lastBold {
				if cell.Bold {
					sb.WriteString("\033[1m")
				} else {
					sb.WriteString("\033[22m")
				}
				lastBold = cell.Bold
				lastBoldSet = true
			}
			if !lastRGBSet || cell.RGB != lastRGB {
				sb.WriteString(fmt.Sprintf("\033[38;2;%d;%d;%dm", cell.RGB[0], cell.RGB[1], cell.RGB[2]))
				lastRGB = cell.RGB
				lastRGBSet = true
			}
			sb.WriteString(cell.Ch)
		}
	}
	sb.WriteString("\033[0m")
	os.Stdout.WriteString(sb.String())
	os.Stdout.Sync()
}

func getBattery() (string, string) {
	dirs, _ := os.ReadDir("/sys/class/power_supply")
	for _, d := range dirs {
		path := filepath.Join("/sys/class/power_supply", d.Name())
		if content, err := os.ReadFile(filepath.Join(path, "type")); err == nil && strings.TrimSpace(string(content)) == "Battery" {
			if scope, err := os.ReadFile(filepath.Join(path, "scope")); err == nil && strings.TrimSpace(string(scope)) == "Device" {
				continue
			}
			capStr := ""
			if capBytes, err := os.ReadFile(filepath.Join(path, "capacity")); err == nil {
				capStr = strings.TrimSpace(string(capBytes))
			} else {
				return "", ""
			}
			statStr := "Unknown"
			if statBytes, err := os.ReadFile(filepath.Join(path, "status")); err == nil {
				statStr = strings.TrimSpace(string(statBytes))
			}
			return capStr, statStr
		}
	}
	return "", ""
}

func parseHex(hex string) [3]int {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) == 6 {
		r, _ := strconv.ParseInt(hex[0:2], 16, 32)
		g, _ := strconv.ParseInt(hex[2:4], 16, 32)
		b, _ := strconv.ParseInt(hex[4:6], 16, 32)
		return [3]int{int(r), int(g), int(b)}
	}
	return [3]int{}
}

func loadPalette() map[string][3]int {
	pal := map[string]string{
		"background": "#08080a",
		"foreground": "#c7d0ff",
		"accent":     "#7f87df",
		"border":     "#4a4f8a",
		"muted":      "#767ca0",
	}

	home, _ := os.UserHomeDir()
	stateDir := filepath.Join(home, ".local", "state", "nejen")
	nejenPath := getNejenPath()

	nameBytes, err := os.ReadFile(filepath.Join(stateDir, "theme.name"))
	if err == nil {
		name := strings.TrimSpace(string(nameBytes))
		paths := []string{
			filepath.Join(home, ".config", "nejen", "themes", name, "theme.toml"),
			filepath.Join(nejenPath, "themes", name, "theme.toml"),
		}
		for _, p := range paths {
			var theme struct {
				Palette map[string]string `toml:"palette"`
			}
			if _, err := toml.DecodeFile(p, &theme); err != nil {
				continue
			}
			// Only take values that are real #rrggbb colors; anything else
			// keeps the built-in default rather than rendering as black.
			for k, v := range theme.Palette {
				if hexColorRegex.MatchString(v) {
					pal[k] = v
				}
			}
			break
		}
	}

	res := make(map[string][3]int)
	for k, v := range pal {
		res[k] = parseHex(v)
	}
	return res
}

func runScreensaverRender(args []string) {
	pal := loadPalette()
	os.Stdout.WriteString("\033[?25l\033[?7l\033[2J") // cursor off, no autowrap, clear screen

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		<-sigCh
		os.Stdout.WriteString("\033[0m\033[?25h\033[?7h")
		os.Stdout.Sync()
		os.Exit(0)
	}()
	defer func() {
		os.Stdout.WriteString("\033[0m\033[?25h\033[?7h")
		os.Stdout.Sync()
	}()

	var lastSize struct{ w, h int }

	for {
		w, h := getTerminalSize()
		if w != lastSize.w || h != lastSize.h {
			lastSize.w = w
			lastSize.h = h
			os.Stdout.WriteString("\033[2J")
		}

		cv := NewCanvas(w, h)
		cx, cy := w/2, h/2

		title := "NEJEN"
		cv.Text(cx-len(title)/2, cy-2, title, pal["accent"], true)

		timeStr := time.Now().Format("15:04:05")
		cv.Text(cx-len(timeStr)/2, cy, timeStr, pal["accent"], false)

		dateStr := time.Now().Format("Monday, January 02")
		cv.Text(cx-len(dateStr)/2, cy+1, dateStr, pal["foreground"], false)

		cap, stat := getBattery()
		if cap != "" {
			batStr := cap + "%"
			if stat == "Charging" {
				batStr += " ⚡"
			} else if stat == "Full" {
				batStr = "Full 󰁹"
			}
			cv.Text(cx-len([]rune(batStr))/2, cy+3, batStr, pal["border"], false)
		}

		cv.Flush()

		now := time.Now()
		sleepDuration := time.Second - time.Duration(now.Nanosecond()) + 10*time.Millisecond
		if sleepDuration <= 0 {
			sleepDuration = time.Second
		}
		time.Sleep(sleepDuration)
	}
}
