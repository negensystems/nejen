package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func init() {
	registerCommand("screenshot", runScreenshot)
}

// hyprMonitor and hyprClient name only the geometry fields the screenshot
// rects are built from; hyprctl reports a great deal more.
type hyprMonitor struct {
	X               int     `json:"x"`
	Y               int     `json:"y"`
	Width           int     `json:"width"`
	Height          int     `json:"height"`
	Scale           float64 `json:"scale"`
	Transform       int     `json:"transform"`
	Focused         bool    `json:"focused"`
	ActiveWorkspace struct {
		ID int `json:"id"`
	} `json:"activeWorkspace"`
}

type hyprClient struct {
	At        [2]int `json:"at"`
	Size      [2]int `json:"size"`
	Workspace struct {
		ID int `json:"id"`
	} `json:"workspace"`
}

// logicalRect renders a monitor the way slurp wants it, "x,y WxH". hyprctl
// reports physical pixels plus a scale factor, and the odd transforms (90 and
// 270 degrees) swap the axes.
func (m hyprMonitor) logicalRect() string {
	w, h := m.Width, m.Height
	if m.Scale > 0 {
		w = int(math.Floor(float64(m.Width) / m.Scale))
		h = int(math.Floor(float64(m.Height) / m.Scale))
	}
	if m.Transform%2 == 1 {
		w, h = h, w
	}
	return fmt.Sprintf("%d,%d %dx%d", m.X, m.Y, w, h)
}

func hyprMonitors() []hyprMonitor {
	out, err := exec.Command("hyprctl", "monitors", "-j").Output()
	if err != nil {
		return nil
	}
	var mons []hyprMonitor
	if json.Unmarshal(out, &mons) != nil {
		return nil
	}
	return mons
}

func runScreenshot(args []string) {
	home, _ := os.UserHomeDir()
	saveDir := os.Getenv("NEJEN_SCREENSHOT_DIR")
	if saveDir == "" {
		saveDir = os.Getenv("XDG_PICTURES_DIR")
		if saveDir == "" {
			userDirsFile := filepath.Join(home, ".config", "user-dirs.dirs")
			if content, err := os.ReadFile(userDirsFile); err == nil {
				for _, line := range strings.Split(string(content), "\n") {
					if strings.HasPrefix(line, "XDG_PICTURES_DIR=") {
						val := strings.TrimPrefix(line, "XDG_PICTURES_DIR=")
						val = strings.Trim(val, `"`)
						val = strings.ReplaceAll(val, "$HOME", home)
						saveDir = val
						break
					}
				}
			}
			if saveDir == "" {
				saveDir = filepath.Join(home, "Pictures")
			}
		}
	}

	if _, err := os.Stat(saveDir); os.IsNotExist(err) {
		exec.Command("notify-send", "-u", "critical", "-t", "3000", fmt.Sprintf("Screenshot directory does not exist: %s", saveDir)).Run()
		os.Exit(1)
	}

	cmdPkill := exec.Command("pkill", "-x", "slurp")
	if err := cmdPkill.Run(); err == nil {
		os.Exit(0)
	}

	editorCmd := os.Getenv("NEJEN_SCREENSHOT_EDITOR")
	if editorCmd == "" {
		editorCmd = "satty"
	}

	var positional []string
	for _, arg := range args {
		if strings.HasPrefix(arg, "--editor=") {
			editorCmd = strings.TrimPrefix(arg, "--editor=")
		} else {
			positional = append(positional, arg)
		}
	}

	mode := "smart"
	if len(positional) > 0 {
		mode = positional[0]
	}
	sink := "slurp"
	if len(positional) > 1 {
		sink = positional[1]
	}

	focusedMonitorRect := func() string {
		for _, m := range hyprMonitors() {
			if m.Focused {
				return m.logicalRect()
			}
		}
		return ""
	}

	visibleRects := func() string {
		out, err := exec.Command("hyprctl", "activeworkspace", "-j").Output()
		if err != nil {
			return ""
		}
		var active struct {
			ID int `json:"id"`
		}
		if json.Unmarshal(out, &active) != nil {
			return ""
		}

		var rects []string
		for _, m := range hyprMonitors() {
			if m.ActiveWorkspace.ID == active.ID {
				rects = append(rects, m.logicalRect())
			}
		}

		if outCli, err := exec.Command("hyprctl", "clients", "-j").Output(); err == nil {
			var clients []hyprClient
			if json.Unmarshal(outCli, &clients) == nil {
				for _, c := range clients {
					if c.Workspace.ID == active.ID {
						rects = append(rects, fmt.Sprintf("%d,%d %dx%d", c.At[0], c.At[1], c.Size[0], c.Size[1]))
					}
				}
			}
		}

		return strings.Join(rects, "\n")
	}

	var freezeCmd *exec.Cmd
	freezeDisplay := func() {
		freezeCmd = exec.Command("hyprpicker", "-r", "-z")
		freezeCmd.Start()
		time.Sleep(100 * time.Millisecond)
	}
	thawDisplay := func() {
		if freezeCmd != nil && freezeCmd.Process != nil {
			freezeCmd.Process.Kill()
		}
	}

	rectRe := regexp.MustCompile(`^(-?[0-9]+),(-?[0-9]+)\s+([0-9]+)x([0-9]+)`)
	snapIfTiny := func(picked, candidates string) string {
		m := rectRe.FindStringSubmatch(picked)
		if m == nil {
			return picked
		}
		x, _ := strconv.Atoi(m[1])
		y, _ := strconv.Atoi(m[2])
		w, _ := strconv.Atoi(m[3])
		h, _ := strconv.Atoi(m[4])

		if w*h >= 20 {
			return picked
		}

		for _, rect := range strings.Split(candidates, "\n") {
			m2 := rectRe.FindStringSubmatch(rect)
			if m2 != nil {
				rx, _ := strconv.Atoi(m2[1])
				ry, _ := strconv.Atoi(m2[2])
				rw, _ := strconv.Atoi(m2[3])
				rh, _ := strconv.Atoi(m2[4])
				if x >= rx && x < rx+rw && y >= ry && y < ry+rh {
					return fmt.Sprintf("%d,%d %dx%d", rx, ry, rw, rh)
				}
			}
		}
		return picked
	}

	annotate := func(target string) {
		if editorCmd == "satty" {
			exec.Command("satty", "--filename", target, "--output-filename", target, "--actions-on-enter", "save-to-clipboard", "--save-after-copy", "--copy-command", "wl-copy").Run()
		} else {
			parts := strings.Split(editorCmd, " ")
			parts = append(parts, target)
			exec.Command(parts[0], parts[1:]...).Run()
		}
	}

	var geometry string
	switch mode {
	case "region":
		freezeDisplay()
		out, _ := exec.Command("slurp").Output()
		geometry = strings.TrimSpace(string(out))
		thawDisplay()
	case "windows":
		freezeDisplay()
		rects := visibleRects()
		cmd := exec.Command("slurp", "-r")
		cmd.Stdin = strings.NewReader(rects)
		out, _ := cmd.Output()
		geometry = strings.TrimSpace(string(out))
		thawDisplay()
	case "fullscreen":
		geometry = focusedMonitorRect()
	case "smart":
		fallthrough
	default:
		rects := visibleRects()
		freezeDisplay()
		cmd := exec.Command("slurp")
		cmd.Stdin = strings.NewReader(rects)
		out, _ := cmd.Output()
		geometry = strings.TrimSpace(string(out))
		thawDisplay()
		geometry = snapIfTiny(geometry, rects)
	}

	if geometry == "" {
		os.Exit(0)
	}

	target := filepath.Join(saveDir, fmt.Sprintf("screenshot-%s.png", time.Now().Format("2006-01-02_15-04-05")))

	if sink == "slurp" {
		if err := exec.Command("grim", "-g", geometry, target).Run(); err != nil {
			os.Exit(1)
		}
		cmdCopy := exec.Command("wl-copy")
		f, _ := os.Open(target)
		cmdCopy.Stdin = f
		cmdCopy.Run()
		f.Close()

		// notify-send -A blocks until the notification is answered or times
		// out, and the answer is what opens the editor -- so wait for it
		// here. Backgrounding it would let the process exit and take the
		// "edit" action with it.
		out, _ := exec.Command("notify-send", "-t", "5000", "-i", target, "-A", "default=edit", "Screenshot captured").Output()
		if strings.TrimSpace(string(out)) == "default" {
			annotate(target)
		}
	} else {
		cmdGrim := exec.Command("grim", "-g", geometry, "-")
		cmdCopy := exec.Command("wl-copy")
		cmdCopy.Stdin, _ = cmdGrim.StdoutPipe()
		cmdCopy.Start()
		cmdGrim.Run()
		cmdCopy.Wait()
	}
}
