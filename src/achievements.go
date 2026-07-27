package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

//go:embed assets/Achievements.svg
var achievementsIconSVG []byte

//go:embed achievements.json
var achievementsFS embed.FS

type AchievementDef struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Hint        string `json:"hint"`
}

type Achievement struct {
	AchievementDef
	Unlocked bool
}

type achievementsData struct {
	Achievements []AchievementDef `json:"achievements"`
}

func loadAchievements() ([]AchievementDef, error) {
	data, err := achievementsFS.ReadFile("achievements.json")
	if err != nil {
		return nil, err
	}
	var ad achievementsData
	if err := json.Unmarshal(data, &ad); err != nil {
		return nil, err
	}
	return ad.Achievements, nil
}

func checkAchievements(total *DailyRecord, streak int, records map[string]*DailyRecord) []Achievement {
	defs, err := loadAchievements()
	if err != nil {
		return nil
	}

	totalActiveHours := float64(total.ActiveSeconds) / 3600.0

	uniqueApps := len(total.Apps)
	maxAppsInDay := 0
	hasWeekendWarrior := false
	hasSpeedrunner := false
	hasLongActiveDay := false
	hasDyingPC := false
	hasHighRAM := false
	hasUnemployed := false

	for date, day := range records {
		appsInDay := len(day.Apps)
		if appsInDay > maxAppsInDay {
			maxAppsInDay = appsInDay
		}

		t, parseErr := time.Parse("2006-01-02", date)
		if parseErr == nil {
			weekday := t.Weekday()
			if (weekday == time.Saturday || weekday == time.Sunday) && day.ActiveSeconds >= 43200 {
				hasWeekendWarrior = true
			}
		}

		// Speedrunner only counts completed days (not today)
		if day.ActiveSeconds > 0 && day.ActiveSeconds < 3600 && date != time.Now().Format("2006-01-02") {
			hasSpeedrunner = true
		}

		if day.ActiveSeconds >= 21600 && day.AFKSeconds < 600 {
			hasLongActiveDay = true
		}

		if day.ActiveSeconds >= 86400 {
			hasUnemployed = true
		}

		if day.CPUSamples > 0 && day.CPUAvg > 70 {
			hasDyingPC = true
		}

		if day.RAMSamples > 0 && day.RAMAvg > 70 {
			hasHighRAM = true
		}
	}

	afkRatio := 1.0
	if total.TotalSeconds > 0 {
		afkRatio = float64(total.AFKSeconds) / float64(total.TotalSeconds)
	}

	result := make([]Achievement, 0, len(defs))
	for _, def := range defs {
		unlocked := false
		switch def.ID {
		case "speedrunner":
			unlocked = hasSpeedrunner
		case "forget_the_toilets":
			unlocked = hasLongActiveDay
		case "unemployed":
			unlocked = hasUnemployed
		case "button_masher":
			unlocked = uniqueApps >= 10
		case "multitasker":
			unlocked = maxAppsInDay >= 10
		case "weekend_warrior":
			unlocked = hasWeekendWarrior
		case "touch_grass_denied":
			unlocked = streak >= 7
		case "chair_potato":
			unlocked = afkRatio < 0.05
		case "night_owl":
			unlocked = totalActiveHours >= 100
		case "keyboard_destroyer":
			unlocked = totalActiveHours >= 250
		case "the_grind":
			unlocked = totalActiveHours >= 500
		case "carbon_monster":
			unlocked = totalActiveHours >= 1000
		case "veteran":
			unlocked = streak >= 30
		case "dying_pc":
			unlocked = hasDyingPC
		case "write_once_trash_everywhere":
			unlocked = hasHighRAM
		}
		if unlocked {
			result = append(result, Achievement{
				AchievementDef: def,
				Unlocked:       true,
			})
		}
	}

	return result
}

func sendNotification(title, message string) {
	switch runtime.GOOS {
	case "linux":
		sent := false
		if os.Getenv("KDE_FULL_SESSION") != "" {
			if err := exec.Command("kdialog", "--passivepopup", fmt.Sprintf("%s\n%s", title, message), "5").Run(); err == nil {
				sent = true
			} else {
				log.Printf("[notif] kdialog failed: %v", err)
			}
		}
		if !sent {
			if err := exec.Command("notify-send", "-a", "NoGrassWrapper", title, message).Run(); err != nil {
				log.Printf("[notif] notify-send failed: %v", err)
			}
		}
	case "darwin":
		exec.Command("osascript", "-e",
			fmt.Sprintf(`display notification "%s" with title "%s"`, message, title)).Run()
	case "windows":
		// Save grass icon to temp file for the notification
		iconPath := filepath.Join(os.TempDir(), "ngw-notif-icon.png")
		if len(iconPNG) > 0 {
			os.WriteFile(iconPath, iconPNG, 0644)
			defer os.Remove(iconPath)
		}
		ps := fmt.Sprintf(
			`Add-Type -AssemblyName System.Windows.Forms; `+
				`$iconPath = '%s'; `+
				`$icon = if (Test-Path $iconPath) { `+
				`[System.Drawing.Icon]::FromHandle([System.Drawing.Bitmap]::FromFile($iconPath).GetHicon()) } `+
				`else { [System.Drawing.SystemIcons]::Information }; `+
				`$n = New-Object System.Windows.Forms.NotifyIcon; `+
				`$n.Icon = $icon; `+
				`$n.BalloonTipTitle = '%s'; `+
				`$n.BalloonTipText = '%s'; `+
				`$n.Visible = $true; `+
				`$n.ShowBalloonTip(3000); `+
				`Start-Sleep -Seconds 3; `+
				`$n.Dispose(); `+
				`if ($icon -ne [System.Drawing.SystemIcons]::Information) { $icon.Dispose() }`,
			strings.ReplaceAll(iconPath, "'", "''"),
			strings.ReplaceAll(title, "'", "''"),
			strings.ReplaceAll(message, "'", "''"),
		)
		cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Sta", "-Command", ps)
		hideWindow(cmd)
		if err := cmd.Run(); err != nil {
			log.Printf("[notif] PowerShell notification failed: %v", err)
		}
	}
}

func sendAchievementNotification(name, description string) {
	sendNotification("Achievement Unlocked!", fmt.Sprintf("%s\n%s", name, description))
}
