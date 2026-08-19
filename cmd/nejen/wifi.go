package main

// Wi-Fi scanning and connection, backing the quick-connect menu
// (config/elephant/menus/nejen_wifi.lua).
//
// This drives nmcli, not iwctl: NEJEN ships networkmanager (packages/core.txt)
// and never ships iwd, so NetworkManager is the stack actually managing the
// link on a stock install. The TUI this used to open -- impala -- is an iwd
// frontend, which means it had no backend to talk to.

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

func init() {
	registerCommand("wifi list", runWifiList)
	registerCommand("wifi toggle", runWifiToggle)
	registerCommand("wifi rescan", runWifiRescan)
	registerCommand("wifi power", runWifiPower)
}

// network is one visible access point, reduced to what the menu draws.
type network struct {
	SSID string
	// Profile is the NetworkManager connection holding this SSID's
	// credentials, empty when none is saved.
	Profile  string
	InUse    bool
	Signal   int
	Security string
	Freq     int
	Rate     string
	Channel  string
}

// nmcliFields splits one `nmcli -t` record. That format separates fields with
// ':' and backslash-escapes any colon inside a value, so a plain Split would
// tear apart SSIDs and the MHz-suffixed frequency field alike.
func nmcliFields(line string) []string {
	var fields []string
	var current strings.Builder
	escaped := false

	for _, r := range line {
		switch {
		case escaped:
			current.WriteRune(r)
			escaped = false
		case r == '\\':
			escaped = true
		case r == ':':
			fields = append(fields, current.String())
			current.Reset()
		default:
			current.WriteRune(r)
		}
	}
	return append(fields, current.String())
}

// savedProfiles maps SSID to the NetworkManager profile that stores it.
//
// A profile's name is not its SSID: it defaults to one but is free-form, and
// renaming it is common (this machine's is "Home" for SSID "WIFI-A04C"). So
// the SSID has to be read out of each wireless profile rather than assumed
// from its name -- otherwise every renamed network reads as unsaved, and
// connecting to it re-prompts for a password it already has.
func savedProfiles() map[string]string {
	saved := make(map[string]string)

	out, err := exec.Command("nmcli", "-t", "-f", "NAME,TYPE", "connection", "show").Output()
	if err != nil {
		return saved
	}

	for _, line := range strings.Split(string(out), "\n") {
		fields := nmcliFields(line)
		if len(fields) < 2 || !strings.Contains(fields[1], "wireless") {
			continue
		}
		name := fields[0]
		ssidOut, err := exec.Command("nmcli", "-g", "802-11-wireless.ssid", "connection", "show", name).Output()
		if err != nil {
			continue
		}
		if ssid := strings.TrimSpace(string(ssidOut)); ssid != "" {
			saved[ssid] = name
		}
	}
	return saved
}

// scanNetworks lists visible access points, strongest first.
//
// A network broadcasting on both bands shows up once per band, and picking the
// weaker entry from a menu would silently connect you to the worse radio -- so
// entries collapse per SSID, keeping whichever is connected, or failing that
// the strongest.
func scanNetworks() []network {
	// --rescan no: without it nmcli decides for itself whether to force a fresh
	// scan, and when it does the call blocks for the better part of ten seconds
	// -- unusable for something that runs every time the menu opens. This reads
	// NetworkManager's cache, which its own background scanning keeps current;
	// `nejen wifi rescan` is there for when that is not good enough.
	out, err := exec.Command("nmcli", "-t", "-f",
		"IN-USE,SSID,SIGNAL,SECURITY,FREQ,RATE,CHAN", "device", "wifi", "list",
		"--rescan", "no").Output()
	if err != nil {
		return nil
	}

	saved := savedProfiles()
	best := make(map[string]network)

	for _, line := range strings.Split(string(out), "\n") {
		fields := nmcliFields(line)
		if len(fields) < 7 {
			continue
		}

		// Hidden networks report an empty SSID and cannot be joined by name.
		ssid := fields[1]
		if ssid == "" {
			continue
		}

		signal, _ := strconv.Atoi(fields[2])
		freq, _ := strconv.Atoi(strings.Fields(fields[4])[0])

		net := network{
			SSID:     ssid,
			InUse:    strings.TrimSpace(fields[0]) == "*",
			Profile:  saved[ssid],
			Signal:   signal,
			Security: fields[3],
			Freq:     freq,
			Rate:     fields[5],
			Channel:  fields[6],
		}

		if prev, seen := best[ssid]; !seen || net.InUse || (!prev.InUse && net.Signal > prev.Signal) {
			best[ssid] = net
		}
	}

	networks := make([]network, 0, len(best))
	for _, net := range best {
		networks = append(networks, net)
	}
	sort.Slice(networks, func(i, j int) bool {
		if networks[i].InUse != networks[j].InUse {
			return networks[i].InUse
		}
		return networks[i].Signal > networks[j].Signal
	})
	return networks
}

// runWifiList prints one tab-separated record per network:
// ssid, in-use, saved, signal, security, freq, rate, channel. The menu provider
// is the only consumer, so the format is fixed rather than pretty.
func runWifiList(args []string) {
	yesno := func(b bool) string {
		if b {
			return "yes"
		}
		return "no"
	}

	for _, net := range scanNetworks() {
		fmt.Printf("%s\t%s\t%s\t%d\t%s\t%d\t%s\t%s\n",
			net.SSID, yesno(net.InUse), yesno(net.Profile != ""), net.Signal,
			net.Security, net.Freq, net.Rate, net.Channel)
	}
}

// promptPassword asks for a passphrase through walker's password mode, so an
// unsaved network can be joined without leaving the menu for a terminal.
// Returns false when the prompt is dismissed.
func promptPassword(ssid string) (string, bool) {
	cmd := exec.Command(nejenSelf(), "open", "launcher", "--dmenu", "--password",
		"-p", "Password for "+ssid)
	cmd.Stdin = strings.NewReader("")
	out, err := cmd.Output()
	if err != nil {
		return "", false
	}

	password := strings.TrimRight(string(out), "\n")
	return password, password != ""
}

// runWifiToggle joins a network, or leaves it when it is the active one, so the
// menu needs a single action per row.
func runWifiToggle(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: nejen wifi toggle <ssid>")
		os.Exit(1)
	}
	ssid := args[0]

	var target network
	var found bool
	for _, net := range scanNetworks() {
		if net.SSID == ssid {
			target, found = net, true
			break
		}
	}

	if found && target.InUse {
		// Down by profile name: `id <ssid>` misses whenever the profile was
		// renamed, and nmcli then reports an unknown connection.
		profile := target.Profile
		if profile == "" {
			profile = ssid
		}
		notify("Disconnecting from " + ssid)
		if err := exec.Command("nmcli", "connection", "down", "id", profile).Run(); err != nil {
			notify("Could not disconnect from " + ssid)
			os.Exit(1)
		}
		return
	}

	// A saved profile carries its own credentials, and an open network needs
	// none; anything else has to be asked for before nmcli will get anywhere.
	var connect *exec.Cmd
	switch {
	case found && target.Profile != "":
		connect = exec.Command("nmcli", "connection", "up", "id", target.Profile)
	case found && target.Security == "":
		connect = exec.Command("nmcli", "device", "wifi", "connect", ssid)
	default:
		password, ok := promptPassword(ssid)
		if !ok {
			return
		}
		connect = exec.Command("nmcli", "device", "wifi", "connect", ssid, "password", password)
	}

	notify("Connecting to " + ssid)
	if out, err := connect.CombinedOutput(); err != nil {
		// nmcli's own message names the actual problem (bad password, timeout,
		// no secrets), which is more use than a generic failure line.
		message := strings.TrimSpace(string(out))
		if idx := strings.LastIndex(message, "Error:"); idx != -1 {
			message = strings.TrimSpace(message[idx+len("Error:"):])
		}
		if message == "" {
			message = "Could not connect to " + ssid
		}
		notify(message)
		os.Exit(1)
	}
	notify("Connected to " + ssid)
}

// runWifiRescan forces a fresh scan; NetworkManager otherwise serves a cache
// that can be a minute or more stale.
func runWifiRescan(args []string) {
	notify("Rescanning for networks")
	exec.Command("nmcli", "device", "wifi", "rescan").Run()
}

// runWifiPower turns the radio on/off, or toggles when given nothing.
func runWifiPower(args []string) {
	state := "toggle"
	if len(args) > 0 {
		state = args[0]
	}

	if state == "toggle" {
		out, _ := exec.Command("nmcli", "radio", "wifi").Output()
		state = "on"
		if strings.TrimSpace(string(out)) == "enabled" {
			state = "off"
		}
	}

	if state == "on" {
		exec.Command("rfkill", "unblock", "wifi").Run()
	}
	if err := exec.Command("nmcli", "radio", "wifi", state).Run(); err != nil {
		notify("Could not turn Wi-Fi " + state)
		os.Exit(1)
	}
	notify("Wi-Fi " + state)
}
