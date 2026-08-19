package main

import (
	"os"
	"os/exec"
	"strings"
)

// newtColors themes nmtui, which is a libnewt program and so takes its palette
// from NEWT_COLORS rather than any config file. Only the sixteen ANSI names
// are accepted -- no hex -- but that is enough here, because the theme engine
// already renders the palette into Alacritty's ANSI slots
// (templates/alacritty.toml.tmpl): blue is palette.border, green is
// palette.accent, white is palette.foreground, black is palette.background.
// Naming ANSI slots therefore lands NEJEN's own colours, the same way
// bluetuith inherits them. Left unset, nmtui renders in stock newt blue and
// red, which is the one surface in this flow ignoring the theme.
var newtColors = strings.Join([]string{
	"root=white,black",
	"window=white,black",
	"shadow=black,black",
	"border=blue,black",
	"title=green,black",
	"button=black,green",
	"actbutton=black,white",
	"checkbox=white,black",
	"actcheckbox=black,green",
	"entry=white,black",
	"disentry=gray,black",
	"label=white,black",
	"listbox=white,black",
	"actlistbox=black,green",
	"sellistbox=black,green",
	"actsellistbox=black,green",
	"textbox=white,black",
	"acttextbox=black,green",
	"emptyscale=,black",
	"fullscale=,green",
	"compactbutton=white,black",
	"helpline=white,black",
	"roottext=white,black",
}, ":")

func init() {
	registerCommand("open wifi", runOpenWifi)
}

// runOpenWifi opens the quick-connect menu, which is what you want nine times
// out of ten: join a network you can see. Profile editing, static addressing
// and VPNs live in nmtui, reachable from the menu's own "Manage networks"
// entry or directly with --manager.
//
// nmtui, not impala: impala is an iwd frontend, and NEJEN ships networkmanager
// without ever shipping iwd, so impala had no backend to talk to.
func runOpenWifi(args []string) {
	exec.Command("rfkill", "unblock", "wifi").Run()

	for _, arg := range args {
		if arg == "--manager" {
			// runOpenTui hands the process off with syscall.Exec, which passes
			// os.Environ() straight through, so setting this here reaches nmtui.
			os.Setenv("NEWT_COLORS", newtColors)
			runOpenTui([]string{"--focus", "--app-id=org.nejen.nmtui", "nmtui"})
			return
		}
	}

	// --nosearch: unlike the app launcher this is a short, ranked list you
	// arrow through, and the rows are two lines (name + detail line). Taller
	// than the audio and bluetooth menus because a scan sees far more networks
	// than a machine has speakers.
	runOpenLauncher([]string{
		"-m", "menus:nejenwifi", "--nosearch",
		"--minheight", "1", "--maxheight", "640",
	})
}
