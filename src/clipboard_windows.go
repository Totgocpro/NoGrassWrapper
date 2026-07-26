//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func copyImageToClipboard(data []byte) error {
	f, err := os.CreateTemp("", "ngw-clip-*.png")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	f.Close()

	// Escape single quotes in path for PowerShell
	path := strings.ReplaceAll(f.Name(), "'", "''")
	ps := fmt.Sprintf(
		`Add-Type -AssemblyName System.Windows.Forms; `+
			`[System.Windows.Forms.Clipboard]::SetImage([System.Drawing.Image]::FromFile('%s'))`,
		path,
	)
	cmd := exec.Command("powershell", "-NoProfile", "-Sta", "-Command", ps)
	return cmd.Run()
}
