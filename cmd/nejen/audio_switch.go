package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func init() {
	registerCommand("audio switch", runAudioSwitch)
}

func osd(args ...string) {
	focusedCmd := exec.Command("hyprctl", "monitors", "-j")
	out, err := focusedCmd.Output()
	if err != nil {
		return
	}

	var monitors []map[string]interface{}
	if err := json.Unmarshal(out, &monitors); err == nil {
		focusedName := ""
		for _, m := range monitors {
			if focused, ok := m["focused"].(bool); ok && focused {
				if name, ok := m["name"].(string); ok {
					focusedName = name
					break
				}
			}
		}

		if focusedName != "" {
			cmdArgs := append([]string{"--monitor", focusedName}, args...)
			exec.Command("swayosd-client", cmdArgs...).Run()
		}
	}
}

type sink struct {
	id        string
	isDefault bool
	name      string
}

func listSinks() []sink {
	var sinks []sink
	cmd := exec.Command("wpctl", "status")
	out, err := cmd.Output()
	if err != nil {
		return sinks
	}

	lines := strings.Split(string(out), "\n")
	grab := false

	re := regexp.MustCompile(`([│├└─* ]+)([0-9]+)\.\s+([^\[]+)`)

	for _, line := range lines {
		if strings.Contains(line, "Sinks:") {
			grab = true
			continue
		}
		if grab && (strings.Contains(line, "Sources:") || strings.Contains(line, "Filters:") || strings.Contains(line, "Streams:") || strings.Contains(line, "Sink endpoints:")) {
			grab = false
		}

		if grab {
			matches := re.FindStringSubmatch(line)
			if len(matches) > 0 {
				isDefault := strings.Contains(matches[1], "*")
				id := matches[2]
				name := strings.TrimSpace(matches[3])

				// Strip "[vol: ...]"
				if volIdx := strings.Index(name, "[vol:"); volIdx != -1 {
					name = strings.TrimSpace(name[:volIdx])
				}

				sinks = append(sinks, sink{
					id:        id,
					isDefault: isDefault,
					name:      name,
				})
			}
		}
	}

	return sinks
}

func runAudioSwitch(args []string) {
	sinks := listSinks()
	if len(sinks) == 0 {
		osd("--custom-message", "No audio outputs found")
		os.Exit(1)
	}

	current := -1
	for i, s := range sinks {
		if s.isDefault {
			current = i
			break
		}
	}

	if current == -1 {
		current = 0
	}

	nextIdx := (current + 1) % len(sinks)
	nextSink := sinks[nextIdx]

	if nextIdx != current {
		exec.Command("wpctl", "set-default", nextSink.id).Run()
	}

	volCmd := exec.Command("wpctl", "get-volume", nextSink.id)
	volOut, _ := volCmd.Output()
	volStr := string(volOut)

	volPct := 0
	if volStr == "" {
		volStr = "Volume: 0"
	}

	parts := strings.Fields(volStr)
	if len(parts) >= 2 {
		if val, err := strconv.ParseFloat(parts[1], 64); err == nil {
			volPct = int(val * 100)
		}
	}

	level := "high"
	if strings.Contains(volStr, "MUTED") || volPct == 0 {
		level = "muted"
	} else if volPct <= 33 {
		level = "low"
	} else if volPct <= 66 {
		level = "medium"
	}

	osd("--custom-message", nextSink.name, "--custom-icon", fmt.Sprintf("sink-volume-%s-symbolic", level))
}
