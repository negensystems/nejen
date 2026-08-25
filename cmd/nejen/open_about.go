package main

func init() {
	registerCommand("open about", runOpenAbout)
}

func runOpenAbout(args []string) {
	// Its own app-id: the program is `bash`, so without one this panel
	// would share org.nejen.bash with every other bash-hosted panel and
	// --focus would raise the wrong window instead of opening About.
	runOpenTui([]string{"--focus", "--app-id=org.nejen.about", "bash", "-c", "fastfetch; read -n 1 -s"})
}
