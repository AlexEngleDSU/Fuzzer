package screen

import (
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// GetPrimaryScreenSize returns the width and height of the primary monitor's usable area.
func GetPrimaryScreenSize() (float32, float32) {
	switch runtime.GOOS {
	case "windows":
		w, h := getWindowsScreenSize()
		return float32(w), float32(h)
	case "linux":
		w, h := getLinuxScreenSize()
		return float32(w), float32(h)
	case "darwin":
		w, h := getMacScreenSize()
		return float32(w), float32(h)
	default:
		return 1280, 720
	}
}
// --------------------- Windows ---------------------
func getWindowsScreenSize() (int, int) {
	// Windows: use PowerShell to get working area (width x height)
	out, err := exec.Command("powershell", "-Command",
		`Add-Type -AssemblyName System.Windows.Forms; [System.Windows.Forms.Screen]::PrimaryScreen.WorkingArea.Width; [System.Windows.Forms.Screen]::PrimaryScreen.WorkingArea.Height`).Output()
	if err != nil {
		return 1280, 720
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) >= 2 {
		w, _ := strconv.Atoi(strings.TrimSpace(lines[0]))
		h, _ := strconv.Atoi(strings.TrimSpace(lines[1]))
		return w, h
	}
	return 1280, 720
}

// --------------------- Linux ---------------------
func getLinuxScreenSize() (int, int) {
	out, err := exec.Command("xrandr").Output()
	if err != nil {
		return 1280, 720
	}

	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, " connected primary") {
			parts := strings.Fields(line)
			for _, p := range parts {
				if strings.Contains(p, "+") && strings.Contains(p, "x") {
					dims := strings.Split(p, "+")[0] // "1920x1080"
					wh := strings.Split(dims, "x")
					w, _ := strconv.Atoi(wh[0])
					h, _ := strconv.Atoi(wh[1])
					return w, h
				}
			}
		}
	}
	return 1280, 720
}

// --------------------- macOS ---------------------
func getMacScreenSize() (int, int) {
	out, err := exec.Command("osascript", "-e",
		`tell application "Finder" to get bounds of window of desktop`).Output()
	if err != nil {
		return 1280, 720
	}
	bounds := strings.Split(strings.TrimSpace(string(out)), ",")
	if len(bounds) >= 4 {
		w, _ := strconv.Atoi(bounds[2])
		h, _ := strconv.Atoi(bounds[3])
		return w, h
	}
	return 1280, 720
}
