package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func init() {
	registerCommand("update", runUpdate)
}

func handleErrorUpdate(err error) {
	if err != nil {
		fmt.Printf("\n\033[0;31mSomething went wrong during the update!\n\nPlease review the output above carefully, correct the error, and retry the update.\033[0m\n")
		os.Exit(1)
	}
}

func runUpdate(args []string) {
	nejenPath := getNejenPath()

	skipConfirm := false
	if len(args) > 0 && args[0] == "-y" {
		skipConfirm = true
	}

	if !skipConfirm {
		if _, err := exec.LookPath("gum"); err == nil {
			cmdStyle := exec.Command("gum", "style", "--border", "rounded", "--border-foreground", "4", "--padding", "0 2",
				"[ NEJEN // UPDATE ]",
				"",
				"Pulls the repo, then upgrades system and AUR packages.",
				"Interrupting a package upgrade midway can leave pacman in a",
				"broken state - stay on AC power and let it finish.")
			cmdStyle.Stdout = os.Stdout
			cmdStyle.Stderr = os.Stderr
			cmdStyle.Run()

			fmt.Println()

			cmdConfirm := exec.Command("gum", "confirm", "Run the update?")
			cmdConfirm.Stdin = os.Stdin
			cmdConfirm.Stdout = os.Stdout
			cmdConfirm.Stderr = os.Stderr
			if err := cmdConfirm.Run(); err != nil {
				fmt.Println("Update cancelled")
				os.Exit(1)
			}
		} else {
			fmt.Print("Run the update? [y/N] ")
			reader := bufio.NewReader(os.Stdin)
			reply, _ := reader.ReadString('\n')
			reply = strings.TrimSpace(reply)
			if strings.ToLower(reply) != "y" {
				fmt.Println("Update cancelled")
				os.Exit(1)
			}
		}
	}

	fmt.Printf("\033[32m\nUpdate nejen\033[0m\n")
	cmdPull := exec.Command("git", "-C", nejenPath, "pull", "--autostash")
	cmdPull.Stdout = os.Stdout
	cmdPull.Stderr = os.Stderr
	handleErrorUpdate(cmdPull.Run())

	// A conflicted pull (merge or autostash pop) exits non-zero and is already
	// caught above. This used to be followed by `git diff --check` guarding a
	// `git reset --merge`, but that command reports whitespace errors as well
	// as conflict markers, so any local edit carrying trailing whitespace --
	// on an otherwise clean update -- printed diff noise and ran the reset.

	fmt.Printf("\033[32m\nUpdate system packages\033[0m\n")
	cmdPacman := exec.Command("sudo", "pacman", "-Syu", "--noconfirm")
	cmdPacman.Stdout = os.Stdout
	cmdPacman.Stderr = os.Stderr
	handleErrorUpdate(cmdPacman.Run())

	aurHelper := ""
	if _, err := exec.LookPath("paru"); err == nil {
		aurHelper = "paru"
	} else if _, err := exec.LookPath("yay"); err == nil {
		aurHelper = "yay"
	}

	if aurHelper != "" {
		cmdQem := exec.Command("pacman", "-Qem")
		if err := cmdQem.Run(); err == nil {
			fmt.Printf("\033[32m\nUpdate AUR packages\033[0m\n")
			cmdAur := exec.Command(aurHelper, "-Sua", "--noconfirm", "--cleanafter")
			cmdAur.Stdout = os.Stdout
			cmdAur.Stderr = os.Stderr
			handleErrorUpdate(cmdAur.Run())
		}
	}

	// A dev-mode install (the repo itself, installed by install.sh) runs the
	// dispatcher straight out of bin/, so the pull above changed its source
	// but not the binary. Rebuild before invoking it below -- otherwise every
	// re-render, and every bar module, keeps running the pre-update code
	// until someone rebuilds by hand. Packaged installs have no go.mod under
	// NEJEN_PATH and get their binary from pacman, so they skip this.
	if _, err := os.Stat(filepath.Join(nejenPath, "go.mod")); err == nil {
		if _, err := exec.LookPath("go"); err == nil {
			fmt.Printf("\033[32m\nRebuild nejen\033[0m\n")
			cmdBuild := exec.Command("go", "build", "-o", filepath.Join(nejenPath, "bin", "nejen"), "./cmd/nejen")
			cmdBuild.Dir = nejenPath
			cmdBuild.Stdout = os.Stdout
			cmdBuild.Stderr = os.Stderr
			handleErrorUpdate(cmdBuild.Run())
		} else {
			fmt.Println("warning: go not found; bin/nejen not rebuilt after pull")
		}
	}

	fmt.Printf("\033[32m\nRe-render keymap and theme\033[0m\n")
	nejenBin := filepath.Join(nejenPath, "bin", "nejen")
	cmdKeymap := exec.Command(nejenBin, "keymap", "render")
	cmdKeymap.Stdout = os.Stdout
	cmdKeymap.Stderr = os.Stderr
	handleErrorUpdate(cmdKeymap.Run())

	// The theme *id* comes from state, not from `nejen theme current` --
	// that one prints the human-facing display name ("Nejen"), which is
	// not a directory under themes/.
	stateHome := os.Getenv("XDG_STATE_HOME")
	if stateHome == "" {
		home, _ := os.UserHomeDir()
		stateHome = filepath.Join(home, ".local", "state")
	}
	if b, err := os.ReadFile(filepath.Join(stateHome, "nejen", "theme.name")); err == nil {
		currTheme := strings.TrimSpace(string(b))
		if currTheme != "" {
			cmdThemeRender := exec.Command(nejenBin, "theme", "render", currTheme)
			cmdThemeRender.Stdout = os.Stdout
			cmdThemeRender.Stderr = os.Stderr
			cmdThemeRender.Run()
		}
	}
}
