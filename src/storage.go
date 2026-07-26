package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"time"
)

const dataVersion = 1

// Storage persists usage data to a JSON file.
type Storage struct {
	mu       sync.RWMutex
	filePath string
	data     *Store
}

// NewStorage creates or loads the storage from the user's config directory.
func NewStorage() (*Storage, error) {
	dir, err := configDir()
	if err != nil {
		return nil, fmt.Errorf("config dir: %w", err)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("mkdir: %w", err)
	}
	path := filepath.Join(dir, "usage.json")

	s := &Storage{filePath: path}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func configDir() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "NoGrassWrapper"), nil
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", "NoGrassWrapper"), nil
	default:
		// XDG_CONFIG_HOME or ~/.config
		if d := os.Getenv("XDG_CONFIG_HOME"); d != "" {
			return filepath.Join(d, "nograsswrapper"), nil
		}
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".config", "nograsswrapper"), nil
	}
}

func (s *Storage) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = &Store{
		DailyRecords: make(map[string]*DailyRecord),
	}

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // fresh start
		}
		return fmt.Errorf("read: %w", err)
	}

	if err := json.Unmarshal(data, s.data); err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	// Ensure today's record exists
	today := time.Now().Format("2006-01-02")
	if _, ok := s.data.DailyRecords[today]; !ok {
		s.data.DailyRecords[today] = &DailyRecord{
			Date: today,
			Apps: make(map[string]*AppUsage),
		}
	}

	s.recalculateStreak()
	return nil
}

// recalculateStreak recomputes streak and longestStreak from daily records.
func (s *Storage) recalculateStreak() {
	dates := make([]string, 0, len(s.data.DailyRecords))
	for date := range s.data.DailyRecords {
		dates = append(dates, date)
	}
	sort.Strings(dates) // oldest first

	currentRun := 0
	longestRun := 0

	for _, date := range dates {
		day := s.data.DailyRecords[date]
		if day.TotalSeconds > 0 {
			currentRun++
			if currentRun > longestRun {
				longestRun = currentRun
			}
		} else {
			currentRun = 0
		}
	}

	s.data.Streak = currentRun
	s.data.LongestStreak = longestRun
}

// LastTouchedGrassDay returns the most recent day with <2h active, or "Never".
func (s *Storage) LastTouchedGrassDay() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var lastDate string
	for date, day := range s.data.DailyRecords {
		if day.ActiveSeconds < 7200 {
			if lastDate == "" || date > lastDate {
				lastDate = date
			}
		}
	}
	if lastDate == "" {
		return "Never"
	}
	return lastDate
}

// Save persists data to disk.
func (s *Storage) Save() error {
	s.mu.RLock()
	data, err := json.MarshalIndent(s.data, "", "  ")
	s.mu.RUnlock()
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Write atomically: write to temp file, then rename
	tmp := s.filePath + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	if err := os.Rename(tmp, s.filePath); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("rename: %w", err)
	}
	return nil
}

// RecordTick logs one second of activity for the given app name.
func (s *Storage) RecordTick(app string, afk bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	today := time.Now().Format("2006-01-02")

	// Day rollover
	if s.data.CurrentDay != "" && s.data.CurrentDay != today {
		s.updateStreakLocked()
	}
	s.data.CurrentDay = today

	day, ok := s.data.DailyRecords[today]
	if !ok {
		day = &DailyRecord{
			Date: today,
			Apps: make(map[string]*AppUsage),
		}
		s.data.DailyRecords[today] = day
	}

	day.TotalSeconds++

	if afk {
		day.AFKSeconds++
	} else {
		day.ActiveSeconds++
		if app != "" {
			a, ok := day.Apps[app]
			if !ok {
				a = &AppUsage{Name: app}
				day.Apps[app] = a
			}
			a.TotalSeconds++
			a.LastSeen = time.Now()
		}
	}

	s.data.CurrentApp = app
	return nil
}

// RecordCPUSample records a CPU usage sample for today.
func (s *Storage) RecordCPUSample(pct float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	today := time.Now().Format("2006-01-02")
	day, ok := s.data.DailyRecords[today]
	if !ok {
		day = &DailyRecord{
			Date: today,
			Apps: make(map[string]*AppUsage),
		}
		s.data.DailyRecords[today] = day
	}
	day.CPUAvg = (day.CPUAvg*float64(day.CPUSamples) + pct) / float64(day.CPUSamples+1)
	day.CPUSamples++
}

// RecordRAMSample records a RAM usage percentage sample for today.
func (s *Storage) RecordRAMSample(pct float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	today := time.Now().Format("2006-01-02")
	day, ok := s.data.DailyRecords[today]
	if !ok {
		day = &DailyRecord{
			Date: today,
			Apps: make(map[string]*AppUsage),
		}
		s.data.DailyRecords[today] = day
	}
	day.RAMAvg = (day.RAMAvg*float64(day.RAMSamples) + pct) / float64(day.RAMSamples+1)
	day.RAMSamples++
}

// RecordGPUSample records a GPU usage percentage sample for today.
func (s *Storage) RecordGPUSample(pct float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	today := time.Now().Format("2006-01-02")
	day, ok := s.data.DailyRecords[today]
	if !ok {
		day = &DailyRecord{
			Date: today,
			Apps: make(map[string]*AppUsage),
		}
		s.data.DailyRecords[today] = day
	}
	day.GPUAvg = (day.GPUAvg*float64(day.GPUSamples) + pct) / float64(day.GPUSamples+1)
	day.GPUSamples++
}

func (s *Storage) updateStreakLocked() {
	yesterday := time.Now().Add(-24 * time.Hour).Format("2006-01-02")

	// Only increment streak if yesterday had activity
	if d, ok := s.data.DailyRecords[yesterday]; ok && d.TotalSeconds > 0 {
		s.data.Streak++
	} else {
		s.data.Streak = 0
	}
	if s.data.Streak > s.data.LongestStreak {
		s.data.LongestStreak = s.data.Streak
	}
}

// TodayData returns a copy of today's record.
func (s *Storage) TodayData() *DailyRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	today := time.Now().Format("2006-01-02")
	day, ok := s.data.DailyRecords[today]
	if !ok {
		return &DailyRecord{
			Date: today,
			Apps: make(map[string]*AppUsage),
		}
	}

	// Return a shallow copy
	cp := &DailyRecord{
		Date:          day.Date,
		TotalSeconds:  day.TotalSeconds,
		ActiveSeconds: day.ActiveSeconds,
		AFKSeconds:    day.AFKSeconds,
		CPUAvg:        day.CPUAvg,
		CPUSamples:    day.CPUSamples,
		RAMAvg:        day.RAMAvg,
		RAMSamples:    day.RAMSamples,
		GPUAvg:        day.GPUAvg,
		GPUSamples:    day.GPUSamples,
		Apps:          make(map[string]*AppUsage),
	}
	for k, v := range day.Apps {
		cp.Apps[k] = &AppUsage{
			Name:         v.Name,
			TotalSeconds: v.TotalSeconds,
			LastSeen:     v.LastSeen,
		}
	}
	return cp
}

// GetUnlockedAchievements returns a set of unlocked achievement IDs.
func (s *Storage) GetUnlockedAchievements() map[string]bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]bool, len(s.data.UnlockedAchievements))
	for _, id := range s.data.UnlockedAchievements {
		result[id] = true
	}
	return result
}

// AddUnlockedAchievement appends a newly unlocked achievement ID.
func (s *Storage) AddUnlockedAchievement(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.UnlockedAchievements = append(s.data.UnlockedAchievements, id)
}

// GetUsername returns the stored username.
func (s *Storage) GetUsername() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.Username
}

// SetUsername sets the username.
func (s *Storage) SetUsername(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.Username = name
}

// GetAvatarPath returns the stored avatar path.
func (s *Storage) GetAvatarPath() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.AvatarPath
}

// SetAvatarPath sets the avatar path.
func (s *Storage) SetAvatarPath(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.AvatarPath = path
}

// Streak returns the current streak.
func (s *Storage) Streak() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.Streak
}

// LongestStreak returns the longest streak ever.
func (s *Storage) LongestStreak() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.LongestStreak
}

// CurrentDay returns today's date string.
func (s *Storage) CurrentDay() string {
	today := time.Now().Format("2006-01-02")
	return today
}

// AllDailyRecords returns a deep copy of all daily records.
func (s *Storage) AllDailyRecords() map[string]*DailyRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	records := make(map[string]*DailyRecord, len(s.data.DailyRecords))
	for date, day := range s.data.DailyRecords {
		cp := &DailyRecord{
			Date:          day.Date,
			TotalSeconds:  day.TotalSeconds,
			ActiveSeconds: day.ActiveSeconds,
			AFKSeconds:    day.AFKSeconds,
			CPUAvg:        day.CPUAvg,
			CPUSamples:    day.CPUSamples,
			RAMAvg:        day.RAMAvg,
			RAMSamples:    day.RAMSamples,
			GPUAvg:        day.GPUAvg,
			GPUSamples:    day.GPUSamples,
			Apps:          make(map[string]*AppUsage, len(day.Apps)),
		}
		for k, v := range day.Apps {
			cp.Apps[k] = &AppUsage{
				Name:         v.Name,
				TotalSeconds: v.TotalSeconds,
				LastSeen:     v.LastSeen,
			}
		}
		records[date] = cp
	}
	return records
}

// TotalData aggregates all daily records into a single total record.
func (s *Storage) TotalData() *DailyRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := &DailyRecord{
		Date: "Total",
		Apps: make(map[string]*AppUsage),
	}

	for _, day := range s.data.DailyRecords {
		total.TotalSeconds += day.TotalSeconds
		total.ActiveSeconds += day.ActiveSeconds
		total.AFKSeconds += day.AFKSeconds
		for name, app := range day.Apps {
			if existing, ok := total.Apps[name]; ok {
				existing.TotalSeconds += app.TotalSeconds
				if app.LastSeen.After(existing.LastSeen) {
					existing.LastSeen = app.LastSeen
				}
			} else {
				total.Apps[name] = &AppUsage{
					Name:         app.Name,
					TotalSeconds: app.TotalSeconds,
					LastSeen:     app.LastSeen,
				}
			}
		}
	}

	return total
}

// Snapshot returns all display data atomically under a single read lock.
type Snapshot struct {
	Data          *DailyRecord
	Streak        int
	LongestStreak int
	Username      string
	AvatarPath    string
	LastGrassDay  string
	WeekAvg       int64 // average active seconds/day this week
	WeekChange    int   // percentage change vs previous week
}

func weekBounds(t time.Time) (monday time.Time, sunday time.Time) {
	weekday := t.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	monday = t.AddDate(0, 0, -(int(weekday) - 1))
	sunday = monday.AddDate(0, 0, 6)
	return
}

func parseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

func (s *Storage) Snapshot() *Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := &DailyRecord{
		Date: "Total",
		Apps: make(map[string]*AppUsage),
	}
	for _, day := range s.data.DailyRecords {
		total.TotalSeconds += day.TotalSeconds
		total.ActiveSeconds += day.ActiveSeconds
		total.AFKSeconds += day.AFKSeconds
		for name, app := range day.Apps {
			if existing, ok := total.Apps[name]; ok {
				existing.TotalSeconds += app.TotalSeconds
				if app.LastSeen.After(existing.LastSeen) {
					existing.LastSeen = app.LastSeen
				}
			} else {
				total.Apps[name] = &AppUsage{
					Name:         app.Name,
					TotalSeconds: app.TotalSeconds,
					LastSeen:     app.LastSeen,
				}
			}
		}
	}

	var lastGrassDay string
	for date, day := range s.data.DailyRecords {
		if day.ActiveSeconds < 7200 {
			if lastGrassDay == "" || date > lastGrassDay {
				lastGrassDay = date
			}
		}
	}
	if lastGrassDay == "" {
		lastGrassDay = "Never"
	}

	now := time.Now()
	thisMon, thisSun := weekBounds(now)
	lastMon := thisMon.AddDate(0, 0, -7)
	lastSun := thisMon.AddDate(0, 0, -1)

	var thisTotal, thisDays int64
	var lastTotal, lastDays int64

	for date, day := range s.data.DailyRecords {
		t, err := parseDate(date)
		if err != nil {
			continue
		}
		if !t.Before(thisMon) && !t.After(thisSun) {
			thisTotal += day.ActiveSeconds
			thisDays++
		}
		if !t.Before(lastMon) && !t.After(lastSun) {
			lastTotal += day.ActiveSeconds
			lastDays++
		}
	}

	var weekAvg int64
	if thisDays > 0 {
		weekAvg = thisTotal / thisDays
	}

	var weekChange int
	if lastDays > 0 {
		lastAvg := lastTotal / lastDays
		if lastAvg > 0 {
			weekChange = int(((weekAvg - lastAvg) * 100) / lastAvg)
		}
	}

	return &Snapshot{
		Data:          total,
		Streak:        s.data.Streak,
		LongestStreak: s.data.LongestStreak,
		Username:      s.data.Username,
		AvatarPath:    s.data.AvatarPath,
		LastGrassDay:  lastGrassDay,
		WeekAvg:       weekAvg,
		WeekChange:    weekChange,
	}
}

// Close saves data before shutting down.
func (s *Storage) Close() error {
	// Update streak on close
	s.mu.Lock()
	s.updateStreakLocked()
	s.mu.Unlock()
	return s.Save()
}
