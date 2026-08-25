package main

func init() {
	registerCommand("open audio", runOpenAudio)
}

// runOpenAudio opens the quick-select menu, which is what you want nine times
// out of ten: move output to another device, or silence the mic. Per-stream
// routing, levels and card profiles live in wiremix, reachable from the menu's
// own "Mixer" entry or directly with --mixer.
func runOpenAudio(args []string) {
	for _, arg := range args {
		if arg == "--mixer" {
			runOpenTui([]string{"--focus", "wiremix"})
			return
		}
	}

	// --nosearch: the list is the handful of devices attached to this machine,
	// all visible at once, so a query box would only add a step between opening
	// the menu and arrowing to the device. Rows are two lines (name + detail
	// line). Matches the bluetooth menu.
	runOpenLauncher([]string{
		"-m", "menus:nejenaudio", "--nosearch",
		"--minheight", "1", "--maxheight", "480",
	})
}
