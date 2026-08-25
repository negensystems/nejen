package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func init() {
	registerCommand("theme bg", runThemeBgHelp)
	registerCommand("theme bg list", runThemeBgList)
	registerCommand("theme bg current", runThemeBgCurrent)
	registerCommand("theme bg dir", runThemeBgDir)
}

// runThemeBgHelp catches a bare `nejen theme bg` and anything the dispatcher
// could not match to a subcommand, which would otherwise fall through to the
// generic "not a nejen command" error.
func runThemeBgHelp(args []string) {
	out, status := os.Stdout, 0
	if len(args) > 0 && args[0] != "help" && args[0] != "--help" && args[0] != "-h" {
		fmt.Fprintf(os.Stderr, "nejen theme bg: unknown subcommand '%s'\n\n", args[0])
		out, status = os.Stderr, 2
	}

	fmt.Fprintln(out, "Usage: nejen theme bg <command>")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "  pick                 Browse wallpapers with previews (Super+Shift+W)")
	fmt.Fprintln(out, "  next | prev          Step through the wallpapers (Super+Ctrl+[Shift+]W)")
	fmt.Fprintln(out, "  random               Switch to a different wallpaper at random")
	fmt.Fprintln(out, "  set <name|path>      Set a wallpaper by path, filename, or display name")
	fmt.Fprintln(out, "  add [--link] <file>  Add images to your collection")
	fmt.Fprintln(out, "  list [--names]       List available wallpapers")
	fmt.Fprintln(out, "  current [--name]     Show the wallpaper in use")
	fmt.Fprintln(out, "  dir                  Print your wallpaper folder")
	fmt.Fprintln(out)
	fmt.Fprintf(out, "Your wallpapers live in %s and are kept across theme changes.\n", bgUserDir())

	os.Exit(status)
}

// runThemeBgList prints one wallpaper per line, in cycling order, so hub
// entries and shell pipelines can consume it directly. --names swaps full
// paths for display labels, which is what menus want to show; `theme bg
// set` accepts either form back.
func runThemeBgList(args []string) {
	names := len(args) > 0 && args[0] == "--names"
	for _, p := range bgList() {
		if names {
			fmt.Println(bgLabel(p))
		} else {
			fmt.Println(p)
		}
	}
}

func runThemeBgCurrent(args []string) {
	cur := currentBg()
	if cur == "" {
		fmt.Fprintln(os.Stderr, "nejen theme bg current: no wallpaper set")
		os.Exit(1)
	}
	if len(args) > 0 && args[0] == "--name" {
		fmt.Println(bgLabel(cur))
		return
	}
	fmt.Println(cur)
}

// runThemeBgDir prints the user's wallpaper folder, creating it on the way
// out so "open this in a file manager" always has somewhere to land.
func runThemeBgDir(args []string) {
	dir := bgUserDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "nejen theme bg dir: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(filepath.Clean(dir))
}
