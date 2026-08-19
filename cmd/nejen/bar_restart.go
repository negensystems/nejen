package main

func init() {
	registerCommand("bar restart", runBarRestart)
}

func runBarRestart(args []string) {
	runAppRestart([]string{"waybar"})
}
