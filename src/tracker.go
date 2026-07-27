package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

// Tracker polls the active window and reports activity to storage.
type Tracker struct {
	store      *Storage
	stopCh     chan struct{}
	stopOnce   sync.Once
	detector   WindowDetector
	idleDetect IdleDetector

	pollInterval time.Duration
	afkTimeout   time.Duration // inactivity before considered AFK
}

// WindowDetector returns the name of the currently focused window.
type WindowDetector interface {
	ActiveWindow() (name string, err error)
}

// IdleDetector returns the system idle duration.
type IdleDetector interface {
	IdleDuration() (time.Duration, error)
}

// NewTracker creates a new tracker.
func NewTracker(store *Storage) *Tracker {
	return &Tracker{
		store:        store,
		stopCh:       make(chan struct{}),
		detector:     newOSDetector(),
		idleDetect:   newOSIdleDetector(),
		pollInterval: 1 * time.Second,
		afkTimeout:   5 * time.Minute,
	}
}

// Start begins polling for active window changes and CPU usage.
func (t *Tracker) Start() {
	log.Println("[tracker] started (poll every", t.pollInterval, ")")
	go t.loop()
	go t.monitorCPU()
}

func (t *Tracker) monitorCPU() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	// First CPU sample immediately (blocks 1s to measure)
	pcts, err := cpu.Percent(1*time.Second, false)
	if err == nil && len(pcts) > 0 {
		t.store.RecordCPUSample(pcts[0])
	}
	// First RAM sample
	if v, err := mem.VirtualMemory(); err == nil {
		t.store.RecordRAMSample(v.UsedPercent)
	}
	// First GPU sample
	t.store.RecordGPUSample(getGPUPercent())

	for {
		select {
		case <-t.stopCh:
			return
		case <-ticker.C:
			pcts, err := cpu.Percent(3*time.Second, false)
			if err != nil {
				log.Printf("[tracker] cpu sample error: %v", err)
			} else if len(pcts) > 0 {
				t.store.RecordCPUSample(pcts[0])
			}
			if v, err := mem.VirtualMemory(); err == nil {
				t.store.RecordRAMSample(v.UsedPercent)
			}
			t.store.RecordGPUSample(getGPUPercent())
		}
	}
}

// Stop signals the tracker to stop.
func (t *Tracker) Stop() {
	t.stopOnce.Do(func() {
		close(t.stopCh)
	})
}

func (t *Tracker) loop() {
	ticker := time.NewTicker(t.pollInterval)
	defer ticker.Stop()

	// For AFK detection: if same window for > afkTimeout without user input
	lastActiveApp := ""
	lastChangeTime := time.Now()

	for {
		select {
		case <-t.stopCh:
			return
		case <-ticker.C:
			t.tick(&lastActiveApp, &lastChangeTime)
		}
	}
}

func (t *Tracker) tick(lastActiveApp *string, lastChangeTime *time.Time) {
	app, err := t.detector.ActiveWindow()
	if err != nil {
		app = "Unknown"
	}

	// Determine if AFK
	afk := false
	if app == *lastActiveApp {
		// Same app — check idle duration
		idle, err := t.idleDetect.IdleDuration()
		if err == nil && idle > t.afkTimeout {
			afk = true
		} else if time.Since(*lastChangeTime) > t.afkTimeout {
			// Fallback: if no window change for > afkTimeout
			afk = true
		}
	} else {
		// Window changed — activity detected
		*lastActiveApp = app
		*lastChangeTime = time.Now()
	}

	// Clean up app name (remove path prefixes, window titles from browsers, etc.)
	app = sanitizeAppName(app)

	if err := t.store.RecordTick(app, afk); err != nil {
		log.Printf("[tracker] record error: %v", err)
	}
}

// sanitizeAppName cleans up a window title to get the logical app name.
func sanitizeAppName(title string) string {
	return extractAppName(title)
}

// getGPUPercent tries to read GPU utilization via nvidia-smi or platform fallbacks.
func getGPUPercent() float64 {
	cmd := exec.Command("nvidia-smi", "--query-gpu=utilization.gpu", "--format=csv,noheader,nounits")
	hideWindow(cmd)
	out, err := cmd.Output()
	if err == nil {
		s := strings.TrimSpace(string(out))
		lines := strings.Split(s, "\n")
		if len(lines) > 0 {
			var pct float64
			if _, err := fmt.Sscanf(strings.TrimSpace(lines[0]), "%f", &pct); err == nil {
				return pct
			}
		}
	}

	// Linux: AMD / Intel from sysfs
	if runtime.GOOS == "linux" {
		files, _ := filepath.Glob("/sys/class/drm/card*/device/gpu_busy_percent")
		for _, f := range files {
			data, err := os.ReadFile(f)
			if err != nil {
				continue
			}
			var pct float64
			if _, err := fmt.Sscanf(strings.TrimSpace(string(data)), "%f", &pct); err == nil {
				return pct
			}
		}
	}

	return 0
}
