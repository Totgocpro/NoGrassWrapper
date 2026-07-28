package main

import (
	"fmt"
	"time"
)

// AppUsage tracks time spent on a single application.
type AppUsage struct {
	Name         string    `json:"name"`
	TotalSeconds int64     `json:"total_seconds"`
	LastSeen     time.Time `json:"last_seen"`
}

// ActivityHeatmap stores activity intensity per day-of-week (Mon=0…Sun=6) per hour (0-23).
type ActivityHeatmap [7][24]int64

// DailyRecord holds all usage data for one calendar day.
type DailyRecord struct {
	Date            string              `json:"date"` // YYYY-MM-DD
	Apps            map[string]*AppUsage `json:"apps"`
	TotalSeconds    int64               `json:"total_seconds"`
	ActiveSeconds   int64               `json:"active_seconds"`
	AFKSeconds      int64               `json:"afk_seconds"`
	HourlyActive    [24]int64           `json:"hourly_active"`
	CPUAvg          float64             `json:"cpu_avg"`
	CPUSamples      int64               `json:"cpu_samples"`
	RAMAvg          float64             `json:"ram_avg"`
	RAMSamples      int64               `json:"ram_samples"`
	GPUAvg          float64             `json:"gpu_avg"`
	GPUSamples      int64               `json:"gpu_samples"`
}

// Store is the persistent data we save to disk.
type Store struct {
	Version       int                     `json:"version"`
	DailyRecords  map[string]*DailyRecord `json:"daily_records"`
	CurrentDay    string                  `json:"current_day"`
	CurrentApp    string                  `json:"current_app"`
	Streak               int                     `json:"streak"`
	LongestStreak        int                     `json:"longest_streak"`
	UnlockedAchievements []string                `json:"unlocked_achievements,omitempty"`
	Username             string                  `json:"username,omitempty"`
	AvatarPath           string                  `json:"avatar_path,omitempty"`
	HiddenApps           []string                `json:"hidden_apps,omitempty"`
}

// PCScore calculates an unbounded "PC Usage Score" based on total data.
func (d *DailyRecord) PCScore(streak int) int {
	activeHours := float64(d.ActiveSeconds) / 3600.0

	// Base: 100 points per active hour (was 10)
	score := activeHours * 100

	// 24h total usage milestone bonus
	if d.TotalSeconds >= 86400 {
		score += 500
	}

	// Streak bonus: +25 per consecutive day, no cap (was 5)
	score += float64(streak * 25)

	// Penalty for too much AFK (>10% of total time)
	if d.TotalSeconds > 0 {
		afkRatio := float64(d.AFKSeconds) / float64(d.TotalSeconds)
		if afkRatio > 0.1 {
			penalty := (afkRatio - 0.1) * 100
			score -= penalty
		}
	}

	if score < 0 {
		score = 0
	}
	return int(score)
}

// PCDyingScore returns a weighted hardware stress score (CPU=50%, RAM=35%, GPU=15%).
func (d *DailyRecord) PCDyingScore() float64 {
	weighted := 0.0
	totalWeight := 0.0
	if d.CPUSamples > 0 {
		weighted += d.CPUAvg * 0.50
		totalWeight += 0.50
	}
	if d.RAMSamples > 0 {
		weighted += d.RAMAvg * 0.35
		totalWeight += 0.35
	}
	if d.GPUSamples > 0 {
		weighted += d.GPUAvg * 0.15
		totalWeight += 0.15
	}
	if totalWeight == 0 {
		return 0
	}
	return weighted / totalWeight
}

// formatScore formats a number in monetary style (1.2k, 3.4M, etc.).
func formatScore(n int) string {
	switch {
	case n >= 1_000_000_000:
		return fmt.Sprintf("%.1fB", float64(n)/1_000_000_000)
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1000:
		return fmt.Sprintf("%.1fk", float64(n)/1000)
	default:
		return fmt.Sprintf("%d", n)
	}
}

// Grade returns a letter grade for the score.
func Grade(score int) string {
	switch {
	case score >= 10000:
		return "SSS"
	case score >= 5000:
		return "SS"
	case score >= 2000:
		return "S"
	case score >= 1000:
		return "A"
	case score >= 500:
		return "B"
	case score >= 200:
		return "C"
	case score >= 100:
		return "D"
	default:
		return "F"
	}
}

// Tier returns a fun description for the score.
func Tier(score int, activeHours float64) string {
	if activeHours >= 720 {
		return "No Life"
	}
	switch {
	case score >= 10000:
		return "No Life"
	case score >= 5000:
		return "Legendary"
	case score >= 2000:
		return "Dedicated"
	case score >= 500:
		return "Balanced"
	case score >= 100:
		return "Casual"
	default:
		return "Touch Grass"
	}
}

func formatDuration(seconds int64) string {
	d := time.Duration(seconds) * time.Second
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}
