package main

import (
	"log"

	"golang.design/x/clipboard"
)

func initClipboard() {
	if err := clipboard.Init(); err != nil {
		log.Printf("[clipboard] init error: %v", err)
	}
}

func copyImageToClipboard(data []byte) error {
	<-clipboard.Write(clipboard.FmtImage, data)
	return nil
}
