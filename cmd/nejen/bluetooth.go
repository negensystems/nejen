package main

// Bluetooth device queries and connect/disconnect, backing the quick-connect
// menu (config/elephant/menus/nejen_bluetooth.lua). Everything here shells out
// to bluetoothctl rather than speaking BlueZ's D-Bus API directly: bluez-utils
// is already a hard dependency, and this keeps the binary free of a D-Bus
// module for what amounts to four property reads.
//
// The full manager (bluetuith) stays one entry away in the menu; this path
// only covers the common case, which is reconnecting a device you already
// paired.

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
)

func init() {
	registerCommand("bluetooth list", runBluetoothList)
	registerCommand("bluetooth toggle", runBluetoothToggle)
	registerCommand("bluetooth power", runBluetoothPower)
}

// device is one paired device, as much as the menu needs to draw a row.
type device struct {
	MAC       string
	Alias     string
	Icon      string
	Connected bool
	Battery   string
	Trusted   bool
	Bonded    bool
	Blocked   bool
	Profiles  []string
}

// profileNames maps the BlueZ UUID labels worth surfacing onto the short names
// people actually use for them. Everything else -- the vendor-specific blocks,
// GAP/GATT, PnP -- is plumbing, and listing it would bury the two facts that
// matter: whether a headset can do high-quality audio, and whether a device
// speaks HID.
var profileNames = map[string]string{
	"Audio Sink":                "A2DP",
	"Advanced Audio Distribu..": "A2DP",
	"Handsfree":                 "HFP",
	"Headset":                   "HSP",
	"Human Interface Device":    "HID",
	"A/V Remote Control":        "AVRCP",
}

// btInfo reads `bluetoothctl info <mac>`. Absent properties stay empty rather
// than erroring: an unpaired-but-remembered device reports almost nothing.
func btInfo(mac string) device {
	dev := device{MAC: mac, Icon: "bluetooth"}

	out, err := exec.Command("bluetoothctl", "info", mac).Output()
	if err != nil {
		return dev
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		key, value, found := strings.Cut(strings.TrimSpace(scanner.Text()), ": ")
		if !found {
			continue
		}
		switch key {
		case "Alias":
			dev.Alias = value
		case "Icon":
			dev.Icon = value
		case "Connected":
			dev.Connected = value == "yes"
		case "Trusted":
			dev.Trusted = value == "yes"
		case "Bonded":
			dev.Bonded = value == "yes"
		case "Blocked":
			dev.Blocked = value == "yes"
		case "Battery Percentage":
			// Reported as `0x5f (95)`; the decimal in parens is the readable half.
			if _, pct, ok := strings.Cut(value, "("); ok {
				dev.Battery = strings.TrimSuffix(pct, ")")
			}
		case "UUID":
			// `Audio Sink              (0000110b-...)`: the label is padded out
			// to a fixed column, so trim before matching.
			label, _, _ := strings.Cut(value, "(")
			if name, ok := profileNames[strings.TrimSpace(label)]; ok && !slices.Contains(dev.Profiles, name) {
				dev.Profiles = append(dev.Profiles, name)
			}
		}
	}

	return dev
}

// pairedDevices lists paired devices, connected ones first so the menu opens
// on whatever is currently in use.
func pairedDevices() []device {
	out, err := exec.Command("bluetoothctl", "devices", "Paired").Output()
	if err != nil {
		return nil
	}

	var connected, rest []device
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		// `Device AA:BB:CC:DD:EE:FF Some Name`
		fields := strings.SplitN(scanner.Text(), " ", 3)
		if len(fields) < 3 || fields[0] != "Device" {
			continue
		}
		dev := btInfo(fields[1])
		if dev.Alias == "" {
			dev.Alias = fields[2]
		}
		if dev.Connected {
			connected = append(connected, dev)
		} else {
			rest = append(rest, dev)
		}
	}

	return append(connected, rest...)
}

// runBluetoothList prints one tab-separated record per paired device:
// mac, alias, connected, icon, battery, flags, profiles. The menu provider is
// the only consumer, so the format is fixed rather than pretty; fields are
// always present (possibly empty) so the Lua side can index them positionally.
func runBluetoothList(args []string) {
	yesno := func(b bool) string {
		if b {
			return "yes"
		}
		return "no"
	}

	for _, dev := range pairedDevices() {
		var flags []string
		if dev.Trusted {
			flags = append(flags, "TRUSTED")
		}
		if dev.Bonded {
			flags = append(flags, "BONDED")
		}
		if dev.Blocked {
			flags = append(flags, "BLOCKED")
		}

		fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			dev.MAC, dev.Alias, yesno(dev.Connected), dev.Icon, dev.Battery,
			strings.Join(flags, "+"), strings.Join(dev.Profiles, "+"))
	}
}

// runBluetoothToggle connects a disconnected device and disconnects a
// connected one, so the menu needs a single action per row.
func runBluetoothToggle(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: nejen bluetooth toggle <address>")
		os.Exit(1)
	}
	mac := args[0]

	before := btInfo(mac)

	action, verb := "connect", "Connecting to"
	if before.Connected {
		action, verb = "disconnect", "Disconnecting"
	}

	name := before.Alias
	if name == "" {
		name = mac
	}
	notify(fmt.Sprintf("%s %s", verb, name))

	// Without a timeout, `bluetoothctl connect` blocks forever on a device
	// that is paired but out of range or powered off -- which, fired from a
	// menu, leaves a process hanging around with nothing to show for it.
	exec.Command("bluetoothctl", "--timeout", "15", action, mac).Run()

	// bluetoothctl exits 0 whether or not it managed anything (a timeout is
	// still a clean exit), so the device's own state is the only honest
	// answer about whether this worked.
	want := action == "connect"
	if btInfo(mac).Connected != want {
		notify(fmt.Sprintf("Could not %s %s", action, name))
		os.Exit(1)
	}

	if want {
		notify(fmt.Sprintf("Connected to %s", name))
	} else {
		notify(fmt.Sprintf("Disconnected %s", name))
	}
}

// runBluetoothPower turns the adapter on/off, or toggles when given nothing.
func runBluetoothPower(args []string) {
	state := "toggle"
	if len(args) > 0 {
		state = args[0]
	}

	if state == "toggle" {
		out, _ := exec.Command("bluetoothctl", "show").Output()
		state = "on"
		if strings.Contains(string(out), "Powered: yes") {
			state = "off"
		}
	}

	if state == "on" {
		exec.Command("rfkill", "unblock", "bluetooth").Run()
	}
	if err := exec.Command("bluetoothctl", "power", state).Run(); err != nil {
		notify("Bluetooth power " + state + " failed")
		os.Exit(1)
	}
	notify("Bluetooth " + state)
}

func notify(message string) {
	exec.Command("notify-send", "-a", "NEJEN", message).Run()
}
