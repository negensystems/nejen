package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func init() {
	registerCommand("update available", runUpdateAvailable)
}

func getGitRev(path string, rev string) string {
	cmd := exec.Command("git", "-C", path, "rev-parse", rev)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func getGitRevShort(path string, rev string) string {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--short", rev)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func runUpdateAvailable(args []string) {
	nejenPath := getNejenPath()

	exec.Command("git", "-C", nejenPath, "fetch", "--quiet").Run()

	localRev := getGitRev(nejenPath, "HEAD")
	remoteRev := getGitRev(nejenPath, "@{u}")

	// No remote/upstream (a local-only checkout) simply means "nothing to
	// report". Exit non-zero so the waybar module stays hidden, and keep
	// stdout clean -- whatever lands there is what the bar would render.
	if localRev == "" || remoteRev == "" {
		fmt.Fprintln(os.Stderr, "nejen update available: no upstream configured for this checkout")
		os.Exit(1)
	}

	if localRev != remoteRev {
		fmt.Printf("nejen update available (%s)\n", getGitRevShort(nejenPath, remoteRev))
		os.Exit(0)
	} else {
		fmt.Printf("nejen is up to date (%s)\n", getGitRevShort(nejenPath, localRev))
		os.Exit(1)
	}
}
