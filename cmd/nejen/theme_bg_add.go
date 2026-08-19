package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	registerCommand("theme bg add", runThemeBgAdd)
}

// runThemeBgAdd files images into the user's wallpaper folder so they join
// the picker and the cycle. Copying is the default because it makes the
// collection self-contained; --link is there for large libraries the user
// would rather keep in one place.
func runThemeBgAdd(args []string) {
	link := false
	var sources []string
	for _, a := range args {
		switch a {
		case "--link":
			link = true
		case "--copy":
			link = false
		default:
			sources = append(sources, a)
		}
	}

	if len(sources) == 0 {
		fmt.Fprintln(os.Stderr, "usage: nejen theme bg add [--link] <image-file>...")
		os.Exit(2)
	}

	dir := bgUserDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "nejen theme bg add: %v\n", err)
		os.Exit(1)
	}

	var added []string
	failed := 0

	for _, src := range sources {
		abs, err := filepath.Abs(src)
		if err != nil {
			fmt.Fprintf(os.Stderr, "nejen theme bg add: %s: %v\n", src, err)
			failed++
			continue
		}
		info, err := os.Stat(abs)
		if err != nil || info.IsDir() {
			fmt.Fprintf(os.Stderr, "nejen theme bg add: not a file: %s\n", src)
			failed++
			continue
		}
		if !bgImageExt[strings.ToLower(filepath.Ext(abs))] {
			fmt.Fprintf(os.Stderr, "nejen theme bg add: unsupported image type: %s\n", src)
			failed++
			continue
		}

		dest := freeBgName(dir, filepath.Base(abs))
		if link {
			err = os.Symlink(abs, dest)
		} else {
			err = copyFile(abs, dest)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "nejen theme bg add: %s: %v\n", src, err)
			failed++
			continue
		}

		added = append(added, dest)
		fmt.Println(dest)
	}

	// Show the result straight away: adding one wallpaper and then having to
	// go find it in a menu is the friction this command exists to remove.
	if len(added) == 1 {
		if err := applyBg(added[0]); err != nil {
			fmt.Fprintf(os.Stderr, "nejen theme bg add: %v\n", err)
			os.Exit(1)
		}
	}

	if failed > 0 && len(added) == 0 {
		os.Exit(1)
	}
}

// freeBgName returns dir/name, or dir/name-2, dir/name-3 ... when the name
// is taken, so adding never silently overwrites an existing wallpaper.
func freeBgName(dir, name string) string {
	ext := filepath.Ext(name)
	stem := strings.TrimSuffix(name, ext)

	candidate := filepath.Join(dir, name)
	for i := 2; ; i++ {
		if _, err := os.Lstat(candidate); err != nil {
			return candidate
		}
		candidate = filepath.Join(dir, fmt.Sprintf("%s-%d%s", stem, i, ext))
	}
}

func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		os.Remove(dest)
		return err
	}
	return out.Close()
}
