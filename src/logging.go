package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"
)

var logFile *os.File

func InitLogging() error {
	dir, err := configDir()
	if err != nil {
		return fmt.Errorf("log dir: %w", err)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("log mkdir: %w", err)
	}

	logPath := filepath.Join(dir, "app.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}

	logFile = f

	multi := io.MultiWriter(os.Stderr, f)
	log.SetOutput(multi)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Printf("[log] initialized: %s", logPath)
	return nil
}

func CloseLogging() {
	if logFile != nil {
		logFile.Close()
	}
}

func RecoverCrash(store *Storage) {
	if r := recover(); r != nil {
		log.Printf("=== CRASH: %v", r)
		log.Printf("=== STACK TRACE:\n%s", debug.Stack())

		if logFile != nil {
			logFile.Sync()
		}

		dir, _ := configDir()
		if dir != "" {
			crashPath := filepath.Join(dir, fmt.Sprintf("crash_%s.log", time.Now().Format("20060102-150405")))
			dump := fmt.Sprintf("Time: %s\nPanic: %v\n\nStack:\n%s\n", time.Now().Format(time.RFC3339), r, debug.Stack())
			os.WriteFile(crashPath, []byte(dump), 0644)
			log.Printf("[crash] dump written: %s", crashPath)
		}

		if store != nil {
			log.Println("[crash] attempting to save data before exit...")
			if err := store.Close(); err != nil {
				log.Printf("[crash] save error: %v", err)
			}
		}

		CloseLogging()
		os.Exit(1)
	}
}
