package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	os.Setenv("LIBAPPINDICATOR_DISABLE_DEPRECATION_WARNINGS", "1")
	log.SetFlags(log.Ltime | log.Lshortfile)
	log.Println("[app] NoGrassWrapper starting...")

	// Initialize storage
	store, err := NewStorage()
	if err != nil {
		log.Fatalf("[app] storage init: %v", err)
	}

	// Initialize tracker
	tracker := NewTracker(store)

	// Initialize tray (system tray icon)
	tray := NewTray(store, tracker)

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("[app] shutting down...")
		tray.Quit()
		os.Exit(0)
	}()

	// Start tracking
	tracker.Start()

	// Run the system tray (blocking)
	tray.Run()

	// Cleanup after tray exits
	tracker.Stop()
	if err := store.Close(); err != nil {
		log.Printf("[app] save error on exit: %v", err)
	}
	log.Println("[app] goodbye!")
}
