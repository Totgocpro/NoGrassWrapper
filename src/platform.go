package main

import (
	"os"
	"os/exec"
	"runtime"
)

// isWindows returns true on Windows.
func isWindows() bool {
	return runtime.GOOS == "windows"
}

// isDarwin returns true on macOS.
func isDarwin() bool {
	return runtime.GOOS == "darwin"
}

// isWayland returns true if the session is Wayland.
func isWayland() bool {
	return os.Getenv("WAYLAND_DISPLAY") != "" || os.Getenv("XDG_SESSION_TYPE") == "wayland"
}

// execCommand runs a command with arguments, ignoring output.
func execCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Start()
}
