//go:build darwin

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
	out, err := d.getMacOSWindowInfo()
	if err != nil {
		return "", err
	}
	// Return just the app name
	parts := strings.SplitN(out, " — ", 2)
	return strings.TrimSpace(parts[0]), nil
}

func (d *osDetector) ActiveWindowTitle() (string, error) {
	out, err := d.getMacOSWindowInfo()
	if err != nil {
		return "", err
	}
	parts := strings.SplitN(out, " — ", 2)
	if len(parts) > 1 {
		return strings.TrimSpace(parts[1]), nil
	}
	return out, nil
}

func (d *osDetector) getMacOSWindowInfo() (string, error) {
	script := `
tell application "System Events"
	set frontApp to first application process whose frontmost is true
	set appName to name of frontApp
	try
		set winTitle to name of front window of frontApp
		return appName & " — " & winTitle
	on error
		return appName
	end try
end tell`
	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// --- Idle detection on macOS ---

type osIdleDetector struct{}

func newOSIdleDetector() IdleDetector {
	return &osIdleDetector{}
}

func (d *osIdleDetector) IdleDuration() (time.Duration, error) {
	// Get idle time in nanoseconds via IOKit
	out, err := exec.Command("ioreg", "-c", "IOHIDSystem", "-r", "-d", "1").Output()
	if err != nil {
		return 0, err
	}
	s := string(out)
	// Parse "HIDIdleTime" = <number> from the plist-ish output
	for _, line := range strings.Split(s, "\n") {
		if strings.Contains(line, "HIDIdleTime") {
			var ns uint64
			if _, err := fmt.Sscanf(line, "\"HIDIdleTime\" = %d", &ns); err == nil {
				return time.Duration(ns), nil
			}
		}
	}
	return 0, fmt.Errorf("idle time not found in ioreg output")
}
