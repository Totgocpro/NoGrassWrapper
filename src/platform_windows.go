//go:build windows

package main

import (
	"os/exec"
	"syscall"
)

// hideWindow configures cmd to execute without opening any console window on Windows.
func hideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
}
