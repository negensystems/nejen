package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	registerCommand("theme bg set", runThemeBgSet)
}

func runThemeBgSet(args []string) {
	if len(args) == 0 || args[0] == "" {
		fmt.Fprintln(os.Stderr, "usage: nejen theme bg set <image-file>")
		fmt.Fprintln(os.Stderr, "       nejen theme bg set --next|--prev|--random")
		os.Exit(2)
	}

	// Cycling flags are accepted here as well as under their own verbs, so
	// `nejen theme bg set --next` and `nejen theme bg next` are the same
	// thing to anyone writing a keybinding or a hub entry.
	switch args[0] {
	case "--next":
		runThemeBgNext(nil)
		return
	case "--prev", "--previous":
		runThemeBgPrev(nil)
		return
	case "--random":
		runThemeBgRandom(nil)
		return
	}

	bg := resolveBg(args[0])
	if bg == "" {
		fmt.Fprintf(os.Stderr, "nejen theme bg set: no such wallpaper: %s\n", args[0])
		fmt.Fprintln(os.Stderr, "see 'nejen theme bg list' for what is available")
		os.Exit(1)
	}

	if err := applyBg(bg); err != nil {
		fmt.Fprintf(os.Stderr, "nejen theme bg set: %v\n", err)
		os.Exit(1)
	}
}

// resolveBg turns whatever the caller typed into a wallpaper path: a real
// file anywhere on disk, or - so menus can show something readable - the
// filename or display label of a wallpaper already in the list. Returns ""
// if nothing matches.
func resolveBg(arg string) string {
	if info, err := os.Stat(arg); err == nil && !info.IsDir() {
		return arg
	}

	for _, p := range bgList() {
		base := filepath.Base(p)
		stem := strings.TrimSuffix(base, filepath.Ext(base))
		if strings.EqualFold(arg, base) || strings.EqualFold(arg, stem) || strings.EqualFold(arg, bgLabel(p)) {
			return p
		}
	}
	return ""
}
