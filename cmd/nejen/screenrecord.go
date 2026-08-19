package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func init() {
	registerCommand("screenrecord", runScreenrecord)
}

const (
	activeFile      = "/tmp/nejen_screenrecord_filename"
	recorderPattern = "^gpu-screen-recorder"
)

func runScreenrecord(args []string) {
	home, _ := os.UserHomeDir()
	saveDir := os.Getenv("NEJEN_SCREENRECORD_DIR")
	if saveDir == "" {
		saveDir = os.Getenv("XDG_VIDEOS_DIR")
		if saveDir == "" {
			userDirsFile := filepath.Join(home, ".config", "user-dirs.dirs")
			if content, err := os.ReadFile(userDirsFile); err == nil {
				for _, line := range strings.Split(string(content), "\n") {
					if strings.HasPrefix(line, "XDG_VIDEOS_DIR=") {
						val := strings.TrimPrefix(line, "XDG_VIDEOS_DIR=")
						val = strings.Trim(val, `"`)
						val = strings.ReplaceAll(val, "$HOME", home)
						saveDir = val
						break
					}
				}
			}
			if saveDir == "" {
				saveDir = filepath.Join(home, "Videos")
			}
		}
	}

	if _, err := os.Stat(saveDir); os.IsNotExist(err) {
		exec.Command("notify-send", "-u", "critical", "-t", "3000", fmt.Sprintf("Screen recording directory does not exist: %s", saveDir)).Run()
		os.Exit(1)
	}

	wantDesktopAudio := false
	wantMicAudio := false
	wantWebcam := false
	webcamDevice := ""
	forcedResolution := ""
	stopOnly := false

	for _, arg := range args {
		switch {
		case arg == "--audio=system":
			wantDesktopAudio = true
		case arg == "--audio=mic":
			wantMicAudio = true
		case arg == "--audio=all":
			wantDesktopAudio = true
			wantMicAudio = true
		case arg == "--webcam":
			wantWebcam = true
		case strings.HasPrefix(arg, "--webcam="):
			wantWebcam = true
			webcamDevice = strings.TrimPrefix(arg, "--webcam=")
		case strings.HasPrefix(arg, "--size="):
			forcedResolution = strings.TrimPrefix(arg, "--size=")
		case arg == "--stop":
			stopOnly = true
		}
	}

	if recorderRunning() {
		endRecording()
	} else if stopOnly {
		os.Exit(1)
	} else {
		if wantWebcam {
			openWebcamOverlay(webcamDevice)
		}
		if !beginRecording(saveDir, wantDesktopAudio, wantMicAudio, forcedResolution) {
			closeWebcamOverlay()
		}
	}
}

func recorderRunning() bool {
	cmd := exec.Command("pgrep", "-f", recorderPattern)
	return cmd.Run() == nil
}

func refreshBarIndicator() {
	exec.Command("pkill", "-RTMIN+8", "waybar").Run()
}

func captureResolution(forcedResolution string) string {
	if forcedResolution != "" {
		return forcedResolution
	}

	out, err := exec.Command("hyprctl", "monitors", "-j").Output()
	if err == nil {
		var monitors []struct {
			Focused bool `json:"focused"`
			Width   int  `json:"width"`
			Height  int  `json:"height"`
		}
		if json.Unmarshal(out, &monitors) == nil {
			for _, m := range monitors {
				if m.Focused {
					if m.Width > 3840 || m.Height > 2160 {
						return "3840x2160"
					}
					return "0x0"
				}
			}
		}
	}
	return "0x0"
}

func closeWebcamOverlay() {
	exec.Command("pkill", "-f", "nejen-webcam").Run()
}

func openWebcamOverlay(device string) {
	closeWebcamOverlay()

	if device == "" {
		out, _ := exec.Command("v4l2-ctl", "--list-devices").CombinedOutput()
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "/dev/video") {
				device = line
				break
			}
		}
		if device == "" {
			exec.Command("notify-send", "-u", "critical", "-t", "3000", "No webcam devices found").Run()
			return
		}
	}

	out, _ := exec.Command("hyprctl", "monitors", "-j").Output()
	var monitorScale float64 = 1.0
	var monitors []struct {
		Focused bool    `json:"focused"`
		Scale   float64 `json:"scale"`
	}
	if json.Unmarshal(out, &monitors) == nil {
		for _, m := range monitors {
			if m.Focused {
				monitorScale = m.Scale
				break
			}
		}
	}

	overlayWidth := int(math.Floor((360.0 * monitorScale) + 0.5))

	formatsOut, _ := exec.Command("v4l2-ctl", "-d", device, "--list-formats-ext").CombinedOutput()
	formatsStr := string(formatsOut)
	size := ""
	for _, wanted := range []string{"640x360", "1280x720", "1920x1080"} {
		if strings.Contains(formatsStr, wanted) {
			size = wanted
			break
		}
	}

	args := []string{"-f", "v4l2"}
	if size != "" {
		args = append(args, "-video_size", size)
	}
	args = append(args, "-framerate", "30", device, "-vf", fmt.Sprintf("scale=%d:-1", overlayWidth), "-window_title", "nejen-webcam", "-noborder", "-fflags", "nobuffer", "-flags", "low_delay", "-probesize", "32", "-analyzeduration", "0", "-loglevel", "quiet")

	cmd := exec.Command("ffplay", args...)
	cmd.Start()
	go cmd.Wait()
	time.Sleep(1 * time.Second)
}

func beginRecording(saveDir string, wantDesktopAudio, wantMicAudio bool, forcedResolution string) bool {
	target := filepath.Join(saveDir, fmt.Sprintf("screenrecording-%s.mp4", time.Now().Format("2006-01-02_15-04-05")))

	sources := []string{}
	if wantDesktopAudio {
		sources = append(sources, "default_output")
	}
	if wantMicAudio {
		sources = append(sources, "default_input")
	}

	args := []string{"-w", "portal", "-k", "auto", "-s", captureResolution(forcedResolution), "-f", "60", "-fm", "cfr", "-fallback-cpu-encoding", "yes", "-o", target}
	if len(sources) > 0 {
		args = append(args, "-a", strings.Join(sources, "|"), "-ac", "aac")
	}

	cmd := exec.Command("gpu-screen-recorder", args...)
	if err := cmd.Start(); err != nil {
		return false
	}
	go cmd.Wait()

	recorder := cmd.Process

	for {
		if recorder == nil {
			break
		}
		if _, err := os.Stat(target); err == nil {
			break
		}
		if err := cmd.Process.Signal(syscall.Signal(0)); err != nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	if err := cmd.Process.Signal(syscall.Signal(0)); err == nil {
		os.WriteFile(activeFile, []byte(target), 0644)
		refreshBarIndicator()
		return true
	}

	return false
}

func shaveLeadingFrame(video string) {
	scratch := strings.TrimSuffix(video, ".mp4") + "-trim.mp4"
	cmd := exec.Command("ffmpeg", "-y", "-ss", "0.1", "-i", video, "-c", "copy", scratch, "-loglevel", "quiet")
	if cmd.Run() == nil {
		os.Rename(scratch, video)
	} else {
		os.Remove(scratch)
	}
}

func announceSaved() {
	content, err := os.ReadFile(activeFile)
	if err != nil {
		return
	}
	video := strings.TrimSpace(string(content))
	if video == "" {
		return
	}
	if _, err := os.Stat(video); os.IsNotExist(err) {
		return
	}

	shaveLeadingFrame(video)

	thumbnail := strings.TrimSuffix(video, ".mp4") + "-preview.png"
	cmd := exec.Command("ffmpeg", "-y", "-i", video, "-ss", "00:00:00.1", "-vframes", "1", "-q:v", "2", thumbnail, "-loglevel", "quiet")
	cmd.Run()

	if _, err := os.Stat(thumbnail); os.IsNotExist(err) {
		thumbnail = video
	}

	go func() {
		out, _ := exec.Command("notify-send", "-t", "10000", "-i", thumbnail, "-A", "default=open", "Screen recording saved", "Open with Super + Alt + , (or click this)").Output()
		if strings.TrimSpace(string(out)) == "default" {
			exec.Command("mpv", video).Run()
		}
		os.Remove(strings.TrimSuffix(video, ".mp4") + "-preview.png")
	}()
}

func endRecording() {
	exec.Command("pkill", "-SIGINT", "-f", recorderPattern).Run()

	tries := 0
	for recorderRunning() && tries < 50 {
		time.Sleep(100 * time.Millisecond)
		tries++
	}

	refreshBarIndicator()
	closeWebcamOverlay()

	if recorderRunning() {
		exec.Command("pkill", "-9", "-f", recorderPattern).Run()
		exec.Command("notify-send", "-u", "critical", "-t", "5000", "Screen recording error", "Recording process had to be force-killed. Video may be corrupted.").Run()
	} else {
		announceSaved()
		// Give announceSaved a chance to start its goroutine before process exits
		time.Sleep(100 * time.Millisecond)
	}

	os.Remove(activeFile)
}
