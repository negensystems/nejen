package main

// Audio device queries and default-device switching, backing the quick-select
// menu (config/elephant/menus/nejen_audio.lua).
//
// This reads `pactl -f json`, not `wpctl status`: wpctl prints a box-drawing
// tree meant for humans, so pulling fields out of it means matching a regex
// against line art, and it carries none of the detail the menu shows -- port,
// sample format, mute state. The JSON is the same data PipeWire hands the
// mixer.
//
// The full mixer (wiremix) stays one entry away in the menu; this path only
// covers the common case, which is moving output to a different device.

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func init() {
	registerCommand("audio list", runAudioList)
	registerCommand("audio default", runAudioDefault)
	registerCommand("audio mute", runAudioMute)
}

// paDevice is one sink or source as pactl reports it. Only the fields the menu
// draws are named; pactl emits a great deal more.
type paDevice struct {
	Index       int    `json:"index"`
	Name        string `json:"name"`
	Description string `json:"description"`
	State       string `json:"state"`
	Mute        bool   `json:"mute"`
	SampleSpec  string `json:"sample_specification"`
	ActivePort  string `json:"active_port"`
	Ports       []struct {
		Name         string `json:"name"`
		Description  string `json:"description"`
		Availability string `json:"availability"`
	} `json:"ports"`
	Volume map[string]struct {
		ValuePercent string `json:"value_percent"`
	} `json:"volume"`
}

// volumePercent returns the first channel's level. Per-channel volumes can
// differ in principle, but a balance offset is a mixer concern; the menu wants
// the one number that answers "how loud is this".
func (d paDevice) volumePercent() string {
	for _, channel := range d.Volume {
		return strings.TrimSuffix(channel.ValuePercent, "%")
	}
	return ""
}

// portDescription resolves the active port id to its readable name, so the row
// says "Speakers" rather than "analog-output-speaker".
func (d paDevice) portDescription() string {
	for _, port := range d.Ports {
		if port.Name == d.ActivePort {
			return port.Description
		}
	}
	return ""
}

func paList(kind string) []paDevice {
	out, err := exec.Command("pactl", "-f", "json", "list", kind+"s").Output()
	if err != nil {
		return nil
	}
	var devices []paDevice
	if json.Unmarshal(out, &devices) != nil {
		return nil
	}
	return devices
}

func paDefault(kind string) string {
	out, err := exec.Command("pactl", "get-default-"+kind).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// runAudioList prints one tab-separated record per device:
// kind, name, description, default, volume, muted, port, spec. The menu
// provider is the only consumer, so the format is fixed rather than pretty.
func runAudioList(args []string) {
	yesno := func(b bool) string {
		if b {
			return "yes"
		}
		return "no"
	}

	for _, kind := range []string{"sink", "source"} {
		def := paDefault(kind)
		for _, dev := range paList(kind) {
			// Every sink has a matching `.monitor` source for recording what it
			// plays. Nobody picks one as their microphone, and listing them
			// would double the menu with entries that mirror rows above.
			if strings.HasSuffix(dev.Name, ".monitor") {
				continue
			}

			fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				kind, dev.Name, dev.Description, yesno(dev.Name == def),
				dev.volumePercent(), yesno(dev.Mute),
				dev.portDescription(), dev.SampleSpec)
		}
	}
}

// runAudioDefault makes a device the default for its kind. Existing streams
// are moved across too -- otherwise "switch output" leaves whatever is already
// playing on the old device, which reads as the switch having silently failed.
func runAudioDefault(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: nejen audio default <sink|source> <name>")
		os.Exit(1)
	}
	kind, name := args[0], args[1]

	if err := exec.Command("pactl", "set-default-"+kind, name).Run(); err != nil {
		notify("Could not switch audio " + kind)
		os.Exit(1)
	}

	streamKind := map[string]string{"sink": "sink-input", "source": "source-output"}[kind]
	if out, err := exec.Command("pactl", "list", "short", streamKind+"s").Output(); err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			if id, _, ok := strings.Cut(line, "\t"); ok && id != "" {
				exec.Command("pactl", "move-"+streamKind, id, name).Run()
			}
		}
	}

	for _, dev := range paList(kind) {
		if dev.Name == name {
			notify(dev.Description)
			return
		}
	}
}

// runAudioMute toggles mute on the default device of the given kind.
func runAudioMute(args []string) {
	kind := "sink"
	if len(args) > 0 {
		kind = args[0]
	}
	if kind != "sink" && kind != "source" {
		fmt.Fprintln(os.Stderr, "usage: nejen audio mute [sink|source]")
		os.Exit(1)
	}

	if err := exec.Command("pactl", "set-"+kind+"-mute", "@DEFAULT_"+strings.ToUpper(kind)+"@", "toggle").Run(); err != nil {
		notify("Could not toggle mute")
		os.Exit(1)
	}
}
