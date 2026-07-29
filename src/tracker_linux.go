//go:build linux

package main

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type osDetector struct{}

func newOSDetector() WindowDetector {
	return &osDetector{}
}

func (d *osDetector) ActiveWindow() (string, error) {
	if isWayland() {
		return getWaylandActiveWindow()
	}
	return getX11ActiveWindow()
}

func (d *osDetector) ActiveWindowTitle() (string, error) {
	if isWayland() {
		return getWaylandWindowTitle()
	}
	// X11: get window title via xdotool
	return getX11WindowTitle()
}

func getX11WindowTitle() (string, error) {
	out, err := exec.Command("xdotool", "getactivewindow", "getwindowname").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func getWaylandActiveWindow() (string, error) {
	// Prefer class/process name over window title
	if name, err := getWaylandAppClass(); err == nil {
		return name, nil
	}
	// Fall back to window title
	if name, err := getWaylandWindowTitle(); err == nil {
		return name, nil
	}
	return "", fmt.Errorf("no wayland method (install kdotool)")
}

func getWaylandAppClass() (string, error) {
	// Try kdotool getclass (KDE Wayland)
	if out, err := exec.Command("kdotool", "getactivewindow", "getclass").Output(); err == nil {
		if name := strings.TrimSpace(string(out)); name != "" {
			return name, nil
		}
	}

	// Try KWin D-Bus: get active window ID, then getWindowInfo for resourceClass
	for _, tool := range []string{"qdbus-qt6", "qdbus"} {
		// Step 1: get active window ID
		out, err := exec.Command(tool, "org.kde.KWin", "/KWin", "org.kde.KWin.activeWindow").Output()
		if err != nil {
			continue
		}
		wid := strings.TrimSpace(string(out))
		if wid == "" || wid == "0" {
			continue
		}
		// Step 2: get window info (fields: id, resourceName, resourceClass, caption...)
		out, err = exec.Command(tool, "org.kde.KWin", "/KWin", "org.kde.KWin.getWindowInfo", "uint:"+wid).Output()
		if err != nil {
			continue
		}
		lines := strings.Split(string(out), "\n")
		if len(lines) >= 3 {
			resourceClass := strings.TrimSpace(lines[2])
			if resourceClass != "" {
				return resourceClass, nil
			}
		}
	}

	return "", fmt.Errorf("no app class available")
}

func getWaylandWindowTitle() (string, error) {
	// Try kdotool getwindowname
	if out, err := exec.Command("kdotool", "getactivewindow", "getwindowname").Output(); err == nil {
		return strings.TrimSpace(string(out)), nil
	}

	// Try qdbus window caption
	for _, tool := range []string{"qdbus-qt6", "qdbus"} {
		out, err := exec.Command(tool, "org.kde.KWin", "/KWin", "org.kde.KWin.activeWindow").Output()
		if err != nil {
			continue
		}
		wid := strings.TrimSpace(string(out))
		if wid == "" || wid == "0" {
			continue
		}
		out, err = exec.Command(tool, "org.kde.KWin", "/KWin", "org.kde.KWin.getWindowInfo", "uint:"+wid).Output()
		if err != nil {
			continue
		}
		// caption is typically on line 4 or 5 depending on KWin version
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			title := strings.TrimSpace(line)
			if title != "" && !isNumeric(title) && !strings.Contains(title, "org.kde.") {
				return title, nil
			}
		}
	}

	return "", fmt.Errorf("no window title available")
}

func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func getX11ActiveWindow() (string, error) {
	// Prefer window class name (more reliable than window titles)
	if out, err := exec.Command("xdotool", "getactivewindow", "getwindowclassname").Output(); err == nil {
		if name := strings.TrimSpace(string(out)); name != "" {
			return name, nil
		}
	}
	// Fall back to window title
	out, err := exec.Command("xdotool", "getactivewindow", "getwindowname").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// --- Idle detection on Linux ---

type osIdleDetector struct{}

func newOSIdleDetector() IdleDetector {
	return &osIdleDetector{}
}

func (d *osIdleDetector) IdleDuration() (time.Duration, error) {
	if isWayland() {
		return getWaylandIdle()
	}
	return getX11Idle()
}

func getWaylandIdle() (time.Duration, error) {
	// Try logind idle hint via logind D-Bus
	cmd := exec.Command("dbus-send", "--session",
		"--dest=org.freedesktop.ScreenSaver",
		"--print-reply=literal",
		"/org/freedesktop/ScreenSaver",
		"org.freedesktop.ScreenSaver.GetActiveTime")
	if out, err := cmd.Output(); err == nil {
		var ms uint32
		if _, err := fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &ms); err == nil {
			return time.Duration(ms) * time.Millisecond, nil
		}
	}
	return 0, fmt.Errorf("wayland idle not available")
}

func getX11Idle() (time.Duration, error) {
	out, err := exec.Command("xprintidle").Output()
	if err != nil {
		return 0, err
	}
	var ms int
	if _, err := fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &ms); err == nil {
		return time.Duration(ms) * time.Millisecond, nil
	}
	return 0, fmt.Errorf("xprintidle parse error")
}

