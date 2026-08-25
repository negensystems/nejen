package main

import (
	"os"
	"strconv"
	"strings"
)

func init() {
	registerCommand("hibernation available", runHibernationAvailable)
}

func runHibernationAvailable(args []string) {
	resumeConf := "/etc/mkinitcpio.conf.d/nejen_resume.conf"

	state, err := os.ReadFile("/sys/power/state")
	if err != nil || !strings.Contains(string(state), "disk") {
		os.Exit(1)
	}

	imageSizeData, err := os.ReadFile("/sys/power/image_size")
	if err != nil {
		os.Exit(1)
	}

	if _, err := os.Stat(resumeConf); err != nil {
		os.Exit(1)
	}

	swapsData, err := os.ReadFile("/proc/swaps")
	if err != nil {
		os.Exit(1)
	}

	var swapKib int64
	lines := strings.Split(string(swapsData), "\n")
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 3 && !strings.Contains(fields[0], "zram") {
			size, _ := strconv.ParseInt(fields[2], 10, 64)
			swapKib += size
		}
	}

	imageBytes, err := strconv.ParseInt(strings.TrimSpace(string(imageSizeData)), 10, 64)
	if err != nil {
		os.Exit(1)
	}

	if swapKib*1024 <= imageBytes {
		os.Exit(1)
	}
}
