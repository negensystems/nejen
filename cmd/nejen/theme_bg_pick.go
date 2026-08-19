package main

func init() {
	registerCommand("theme bg pick", runThemeBgPick)
}

// runThemeBgPick opens the wallpaper grid in walker. Going through
// `nejen open launcher` means the picker inherits walker's theming, its
// image preview pane, and the elephant autostart the launcher already does.
func runThemeBgPick(args []string) {
	runOpenLauncher([]string{"-m", "menus:nejenbackgrounds"})
}
