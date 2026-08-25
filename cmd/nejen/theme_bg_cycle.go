package main

import (
	"fmt"
	"math/rand"
	"os"
)

func init() {
	registerCommand("theme bg next", runThemeBgNext)
	registerCommand("theme bg prev", runThemeBgPrev)
	registerCommand("theme bg random", runThemeBgRandom)
}

func runThemeBgNext(args []string) { cycleBg(+1) }
func runThemeBgPrev(args []string) { cycleBg(-1) }

func runThemeBgRandom(args []string) {
	list := bgList()
	if !haveWallpapers(list) {
		return
	}
	if len(list) == 1 {
		selectBg(list[0])
		return
	}

	// Draw from everything except the current wallpaper, so "random" always
	// visibly changes something.
	choices := list
	if i := bgIndexOf(list, currentBg()); i >= 0 {
		choices = make([]string, 0, len(list)-1)
		choices = append(choices, list[:i]...)
		choices = append(choices, list[i+1:]...)
	}
	selectBg(choices[rand.Intn(len(choices))])
}

// cycleBg steps step places through the wallpaper list, wrapping at both
// ends. An unrecognised current wallpaper starts the walk at the top.
func cycleBg(step int) {
	list := bgList()
	if !haveWallpapers(list) {
		return
	}

	next := 0
	if i := bgIndexOf(list, currentBg()); i >= 0 {
		next = ((i+step)%len(list) + len(list)) % len(list)
	}
	selectBg(list[next])
}

func selectBg(path string) {
	if err := applyBg(path); err != nil {
		fmt.Fprintf(os.Stderr, "nejen theme bg: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(path)
}

// haveWallpapers reports whether there is anything to switch to, and
// explains where to put images when there is not.
func haveWallpapers(list []string) bool {
	if len(list) > 0 {
		return true
	}
	fmt.Fprintf(os.Stderr, "nejen theme bg: no wallpapers found. Add images with 'nejen theme bg add <file>' or drop them in %s\n", bgUserDir())
	os.Exit(1)
	return false
}
