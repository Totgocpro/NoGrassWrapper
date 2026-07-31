package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/getlantern/systray"
)

// Tray manages the system tray icon and menu.
type Tray struct {
	store   *Storage
	tracker *Tracker
	quitCh  chan struct{}
	mStats  *systray.MenuItem
}

// NewTray creates a new tray manager.
func NewTray(store *Storage, tracker *Tracker) *Tray {
	return &Tray{
		store:   store,
		tracker: tracker,
		quitCh:  make(chan struct{}),
	}
}

func (t *Tray) populateAchievements() {
	data := t.store.TotalData()
	streak := t.store.Streak()
	records := t.store.AllDailyRecords()
	allUnlocked := checkAchievements(data, streak, records)
	prevUnlocked := t.store.GetUnlockedAchievements()
	for _, a := range allUnlocked {
		if !prevUnlocked[a.ID] {
			t.store.AddUnlockedAchievement(a.ID)
		}
	}
}

func (t *Tray) checkNewAchievements() {
	data := t.store.TotalData()
	streak := t.store.Streak()
	records := t.store.AllDailyRecords()
	allUnlocked := checkAchievements(data, streak, records)
	prevUnlocked := t.store.GetUnlockedAchievements()
	for _, a := range allUnlocked {
		if !prevUnlocked[a.ID] {
			t.store.AddUnlockedAchievement(a.ID)
			sendAchievementNotification(a.Name, a.Description)
		}
	}
}

// Run starts the system tray. This is a blocking call.
func (t *Tray) Run() {
	systray.Run(t.onReady, t.onExit)
}

// Quit signals the tray to exit.
func (t *Tray) Quit() {
	systray.Quit()
}

func (t *Tray) onReady() {
	systray.SetIcon(getIcon())
	systray.SetTitle("NGW")
	systray.SetTooltip("NoGrassWrapper — tracking your screen time")

	initClipboard()

	// Current stats in menu
	t.mStats = systray.AddMenuItem("Tracking...", "Current stats")
	t.mStats.Disable()

	systray.AddSeparator()

	// Copy wrapper image to clipboard
	mCopy := systray.AddMenuItem("Copy Wrapper", "Generate and copy the wrapper image to clipboard")

	systray.AddSeparator()

	mSettings := systray.AddMenuItem("Settings", "Open settings")

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit", "Stop tracking and exit")

	// Start a goroutine to check achievements periodically
	go func() {
		time.Sleep(5 * time.Second)
		t.populateAchievements()
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				t.checkNewAchievements()
			case <-t.quitCh:
				return
			}
		}
	}()

	// Start a goroutine to periodically save data
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := t.store.Save(); err != nil {
					log.Printf("[tray] periodic save error: %v", err)
				}
			case <-t.quitCh:
				return
			}
		}
	}()

	// Start a goroutine to update stats display
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				data := t.store.TotalData()
				streak := t.store.Streak()
				score := data.PCScore(streak)
				tier := Tier(score, float64(data.ActiveSeconds)/3600.0)
				t.mStats.SetTitle(fmt.Sprintf("%s — Score: %s (%s)", formatDuration(data.ActiveSeconds), formatScore(score), tier))
			case <-t.quitCh:
				return
			}
		}
	}()

	// Handle menu actions
	go func() {
		for {
			select {
			case <-mCopy.ClickedCh:
				go t.generateAndCopy()
			case <-mSettings.ClickedCh:
				go showSettingsDialog(t.store)
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			case <-t.quitCh:
				return
			}
		}
	}()
}

func (t *Tray) onExit() {
	log.Println("[tray] exiting")
	t.tracker.Stop()
	if err := t.store.Close(); err != nil {
		log.Printf("[tray] save error: %v", err)
	}
	close(t.quitCh)
	os.Exit(0)
}

func (t *Tray) generateAndCopy() {
	sendNotification("NoGrassWrapper", "Wrapper image copied to clipboard!")

	snap := t.store.Snapshot()

	avatarPath := snap.AvatarPath
	if avatarPath == "" {
		avatarPath = findAvatar()
	}
	records := t.store.AllDailyRecords()
	achievements := checkAchievements(snap.Data, snap.Streak, records)
	heatmap := t.store.ActivityHeatmap()

	renderer := NewWrapperImage()
	imgBytes, err := renderer.GenerateBytes(snap.Data, snap.Streak, snap.LongestStreak, avatarPath, achievements, snap.Username, snap.LastGrassDay, snap.WeekAvg, snap.WeekChange, snap.HiddenApps, heatmap, snap.SplitBrowserURLs, snap.HideApps)
	if err != nil {
		log.Printf("[tray] generate error: %v", err)
		return
	}

	// Save image to desktop (skip on Windows — would clutter desktop)
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, "Desktop")
	if runtime.GOOS == "windows" {
		dir = os.TempDir()
	} else if _, err := os.Stat(dir); os.IsNotExist(err) {
		dir = home
	}
	path := filepath.Join(dir, fmt.Sprintf("nograss_wrapper_%s.png", time.Now().Format("2006-01-02")))
	if err := os.WriteFile(path, imgBytes, 0644); err != nil {
		log.Printf("[tray] save error: %v", err)
	}

	// Copy to clipboard
	if err := copyImageToClipboard(imgBytes); err != nil {
		log.Printf("[tray] clipboard error: %v", err)
		t.mStats.SetTitle(fmt.Sprintf("Clipboard error — %s", err.Error()))
		return
	}

	log.Printf("[tray] wrapper image copied to clipboard: %s", path)
	t.mStats.SetTitle("Wrapper copied to clipboard!")
	time.AfterFunc(3*time.Second, func() {
		tier := Tier(snap.Data.PCScore(snap.Streak), float64(snap.Data.ActiveSeconds)/3600.0)
		t.mStats.SetTitle(fmt.Sprintf("%s — Score: %s (%s)", formatDuration(snap.Data.ActiveSeconds), formatScore(snap.Data.PCScore(snap.Streak)), tier))
	})
}

func findAvatar() string {
	if path := os.Getenv("NOGRASS_AVATAR"); path != "" {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	home, _ := os.UserHomeDir()
	user := os.Getenv("USER")
	candidates := []string{
		filepath.Join(home, ".face"),
		filepath.Join(home, ".face.icon"),
		filepath.Join(home, "Pictures", "avatar.png"),
		filepath.Join(home, "Pictures", "avatar.jpg"),
		filepath.Join(home, "Pictures", "avatar.jpeg"),
		filepath.Join(home, "Photos", "avatar.png"),
		filepath.Join(home, "Photos", "avatar.jpg"),
		filepath.Join(home, ".local", "share", "icons", user+".png"),
	}
	// Check config dir for any avatar file (saved via settings upload)
	configDir, _ := configDir()
	exts := []string{".png", ".jpg", ".jpeg", ".gif", ".webp"}
	for _, ext := range exts {
		candidates = append(candidates, filepath.Join(configDir, "avatar"+ext))
	}
	if runtime.GOOS == "linux" {
		candidates = append(candidates,
			fmt.Sprintf("/var/lib/AccountsService/icons/%s", user),
		)
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}
