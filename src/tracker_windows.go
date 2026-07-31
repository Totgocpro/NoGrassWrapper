//go:build windows

package main

import (
	"fmt"
	"time"
	"unsafe"
	"golang.org/x/sys/windows"
)

type osDetector struct{}

func newOSDetector() WindowDetector {
	return &osDetector{}
}

func (d *osDetector) ActiveWindow() (string, error) {
	return d.ActiveWindowTitle()
}

func (d *osDetector) ActiveWindowTitle() (string, error) {
	mod := windows.NewLazySystemDLL("user32.dll")
	procGetForegroundWindow := mod.NewProc("GetForegroundWindow")
	procGetWindowTextW := mod.NewProc("GetWindowTextW")
	procGetWindowTextLenW := mod.NewProc("GetWindowTextLengthW")

	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return "Unknown", nil
	}

	textLen, _, _ := procGetWindowTextLenW.Call(hwnd)
	if textLen == 0 {
		return "Unknown", nil
	}

	buf := make([]uint16, textLen+1)
	procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(textLen+1))
	return windows.UTF16ToString(buf), nil
}

// --- Idle detection on Windows ---

type osIdleDetector struct{}

func newOSIdleDetector() IdleDetector {
	return &osIdleDetector{}
}

type lastInputInfo struct {
	cbSize uint32
	dwTime uint32
}

func (d *osIdleDetector) IdleDuration() (time.Duration, error) {
	mod := windows.NewLazySystemDLL("user32.dll")
	procGetLastInputInfo := mod.NewProc("GetLastInputInfo")

	lii := lastInputInfo{cbSize: uint32(unsafe.Sizeof(lastInputInfo{}))}
	ret, _, _ := procGetLastInputInfo.Call(uintptr(unsafe.Pointer(&lii)))
	if ret == 0 {
		return 0, fmt.Errorf("GetLastInputInfo failed")
	}

	modKernel := windows.NewLazySystemDLL("kernel32.dll")
	procGetTickCount := modKernel.NewProc("GetTickCount")
	tick, _, _ := procGetTickCount.Call()

	idle := uint32(tick) - lii.dwTime
	return time.Duration(idle) * time.Millisecond, nil
}
