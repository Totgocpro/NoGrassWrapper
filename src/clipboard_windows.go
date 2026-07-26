//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
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

	ps := fmt.Sprintf(
		`Add-Type -AssemblyName System.Windows.Forms; `+
			`[Windows.Forms.Clipboard]::SetImage([Drawing.Image]::FromFile('%s'))`,
		f.Name(),
	)
	cmd := exec.Command("powershell", "-NoProfile", "-Command", ps)
	return cmd.Run()
}
