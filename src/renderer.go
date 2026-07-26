package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fogleman/gg"
	"golang.org/x/image/font/gofont/goregular"
)

// Version is set at build time via -ldflags. Default "dev".
var Version = "dev"

// fontFace tries to load a font from system paths, falling back to embedded Go font.
func fontFace(dc *gg.Context, size float64) error {
	// Try system fonts
	paths := []string{
		"/usr/share/fonts/TTF/DejaVuSans.ttf",
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
		"/System/Library/Fonts/Helvetica.ttc",
		"/Library/Fonts/Arial.ttf",
		"C:\\Windows\\Fonts\\arial.ttf",
		"C:\\Windows\\Fonts\\segoeui.ttf",
	}
	for _, p := range paths {
		if err := dc.LoadFontFace(p, size); err == nil {
			return nil
		}
	}
	// Use embedded Go font as fallback — write to temp file
	f, err := os.CreateTemp("", "nograsswrapper-font-*.ttf")
	if err == nil {
		f.Write(goregular.TTF)
		path := f.Name()
		f.Close()
		err = dc.LoadFontFace(path, size)
		os.Remove(path)
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("no font found")
}

// fontFaceBold tries to load a bold font from system paths, falling back to embedded.
func fontFaceBold(dc *gg.Context, size float64) error {
	paths := []string{
		"/usr/share/fonts/TTF/DejaVuSans-Bold.ttf",
		"/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf",
		"/System/Library/Fonts/Helvetica.ttc",
		"C:\\Windows\\Fonts\\arialbd.ttf",
		"C:\\Windows\\Fonts\\segoeuib.ttf",
	}
	for _, p := range paths {
		if err := dc.LoadFontFace(p, size); err == nil {
			return nil
		}
	}
	return fontFace(dc, size) // fallback to regular
}

// WrapperImage generates the "No Grass Wrapper" summary image.
type WrapperImage struct {
	width  int
	height int
}

// NewWrapperImage creates a new renderer with sensible defaults.
func NewWrapperImage() *WrapperImage {
	return &WrapperImage{
		width:  1000,
		height: 800,
	}
}

type appBar struct {
	Name  string
	Value int64  // seconds
	Pct   float64 // 0-1
	Color color.RGBA
}

var appColors = []color.RGBA{
	{R: 52, G: 211, B: 153, A: 255},  // mint
	{R: 56, G: 178, B: 172, A: 255},  // teal
	{R: 99, G: 102, B: 241, A: 255},  // indigo
	{R: 129, G: 140, B: 248, A: 255}, // periwinkle
	{R: 167, G: 139, B: 250, A: 255}, // purple
	{R: 196, G: 181, B: 253, A: 255}, // lavender
	{R: 244, G: 114, B: 182, A: 255}, // pink
	{R: 251, G: 146, B: 60, A: 255},  // orange
	{R: 250, G: 204, B: 21, A: 255},  // yellow
	{R: 34, G: 197, B: 94, A: 255},   // green
}

// GenerateBytes generates the PNG image and returns it as bytes.
func (w *WrapperImage) GenerateBytes(data *DailyRecord, streak, longestStreak int, avatarPath string, achievements []Achievement, username, lastGrassDay string, weekAvg int64, weekChange int, hiddenApps []string) ([]byte, error) {
	dc := gg.NewContext(w.width, w.height)

	w.drawBackground(dc)
	w.drawAvatar(dc, avatarPath)
	w.drawHeader(dc, username)

	score := data.PCScore(streak)
	grade := Grade(score)
	tier := Tier(score, float64(data.ActiveSeconds)/3600.0)
	w.drawStats(dc, data, streak, longestStreak, score, grade, tier, lastGrassDay, weekAvg, weekChange)
	w.drawAchievements(dc, achievements)
	w.drawAppBars(dc, data, hiddenApps)
	w.drawFooter(dc, streak, longestStreak, score, data.PCDyingScore())

	var buf bytes.Buffer
	if err := png.Encode(&buf, dc.Image()); err != nil {
		return nil, fmt.Errorf("encode: %w", err)
	}
	return buf.Bytes(), nil
}

// Generate creates a PNG image at the given path with usage stats.
func (w *WrapperImage) Generate(path string, today *DailyRecord, streak, longestStreak int, avatarPath string, achievements []Achievement, username, lastGrassDay string, weekAvg int64, weekChange int, hiddenApps []string) error {
	data, err := w.GenerateBytes(today, streak, longestStreak, avatarPath, achievements, username, lastGrassDay, weekAvg, weekChange, hiddenApps)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func (w *WrapperImage) drawBackground(dc *gg.Context) {
	// Dark gradient background
	dc.SetColor(color.RGBA{15, 15, 25, 255})
	dc.Clear()

	// Subtle grid pattern
	dc.SetColor(color.RGBA{30, 30, 50, 255})
	for x := 0; x < w.width; x += 40 {
		dc.SetLineWidth(1)
		dc.DrawLine(float64(x), 0, float64(x), float64(w.height))
		dc.Stroke()
	}
	for y := 0; y < w.height; y += 40 {
		dc.DrawLine(0, float64(y), float64(w.width), float64(y))
		dc.Stroke()
	}

	// Top accent bar
	grad := gg.NewLinearGradient(0, 0, float64(w.width), 0)
	grad.AddColorStop(0, color.RGBA{52, 211, 153, 255})   // mint
	grad.AddColorStop(0.5, color.RGBA{99, 102, 241, 255}) // indigo
	grad.AddColorStop(1, color.RGBA{244, 114, 182, 255})  // pink
	dc.SetFillStyle(grad)
	dc.DrawRectangle(0, 0, float64(w.width), 6)
	dc.Fill()

	// Watermark: grass logo
	if wm, err := renderSVG(iconSVG, 350, 350); err == nil {
		bounds := wm.Bounds()
		faded := image.NewRGBA(bounds)
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				r, g, b, a := wm.At(x, y).RGBA()
				na := uint8(float64(a>>8) * 0.04)
				faded.Set(x, y, color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), na})
			}
		}
		dc.DrawImageAnchored(faded, w.width/2, w.height/2, 0.5, 0.5)
	}
}

func (w *WrapperImage) drawHeader(dc *gg.Context, username string) {
	fontFaceBold(dc, 36)
	dc.SetColor(color.RGBA{255, 255, 255, 255})
	dc.DrawStringAnchored("NO GRASS WRAPPER", float64(w.width)/2, 30, 0.5, 0.5)

	subtitle := "Your PC Usage Dashboard — All Time Stats"
	if username != "" {
		subtitle = fmt.Sprintf("%s's PC Stats — All Time", username)
	}
	fontFace(dc, 14)
	dc.SetColor(color.RGBA{150, 150, 180, 255})
	dc.DrawStringAnchored(subtitle, float64(w.width)/2, 55, 0.5, 0.5)

	dc.SetColor(color.RGBA{60, 60, 90, 255})
	dc.SetLineWidth(1)
	dc.DrawLine(60, 68, float64(w.width)-60, 68)
	dc.Stroke()
}

func (w *WrapperImage) drawStats(dc *gg.Context, today *DailyRecord, streak, longestStreak, score int, grade, tier, lastGrassDay string, weekAvg int64, weekChange int) {
	// Stats cards
	changeSign := "+"
	if weekChange < 0 {
		changeSign = ""
	}
	stats := []struct {
		Label string
		Value string
		Desc  string
	}{
		{"Screen Time", formatDuration(today.ActiveSeconds), "total active"},
		{"Streak", fmt.Sprintf("%d days", streak), fmt.Sprintf("best: %d", longestStreak)},
		{"Score", formatScore(score), fmt.Sprintf("Grade %s · %s", grade, tier)},
		{"Last Touched Grass", lastGrassDay, "day with <2h active"},
		{"Daily Avg (Week)", formatDuration(weekAvg), fmt.Sprintf("%s%d%% vs last week", changeSign, weekChange)},
	}

	cardW := 170.0
	cardH := 92.0
	gap := 12.0
	totalW := float64(len(stats))*cardW + float64(len(stats)-1)*gap
	startX := (float64(w.width) - totalW) / 2
	y := 85.0

	for i, s := range stats {
		x := startX + float64(i)*(cardW+gap)

		// Card background
		dc.SetColor(color.RGBA{25, 25, 45, 220})
		dc.DrawRoundedRectangle(x, y, cardW, cardH, 12)
		dc.Fill()

		// Border
		dc.SetColor(color.RGBA{50, 50, 80, 200})
		dc.SetLineWidth(1)
		dc.DrawRoundedRectangle(x, y, cardW, cardH, 12)
		dc.Stroke()

		// Label
		fontFace(dc, 11)
		dc.SetColor(color.RGBA{150, 150, 180, 255})
		dc.DrawStringAnchored(s.Label, x+cardW/2, y+22, 0.5, 0.5)

		// Value
		fontFaceBold(dc, 22)
		dc.SetColor(color.RGBA{255, 255, 255, 255})
		dc.DrawStringAnchored(s.Value, x+cardW/2, y+54, 0.5, 0.5)

		// Description
		fontFace(dc, 10)
		dc.SetColor(color.RGBA{120, 120, 160, 255})
		dc.DrawStringAnchored(s.Desc, x+cardW/2, y+78, 0.5, 0.5)
	}
}

func (w *WrapperImage) drawAchievements(dc *gg.Context, achievements []Achievement) {
	dc.SetColor(color.RGBA{200, 200, 230, 255})
	fontFaceBold(dc, 16)
	dc.DrawStringAnchored("Achievements", 250, 205, 0.5, 0.5)

	if len(achievements) == 0 {
		fontFace(dc, 12)
		dc.SetColor(color.RGBA{120, 120, 160, 255})
		dc.DrawStringAnchored("No achievements yet", 250, 250, 0.5, 0.5)
		return
	}

	maxShow := len(achievements)
	if maxShow > 8 {
		maxShow = 8
	}

	icon, _ := renderSVG(bytes.ReplaceAll(achievementsIconSVG, []byte("#000000"), []byte("#34d399")), 14, 14)

	for i := 0; i < maxShow; i++ {
		a := achievements[i]
		y := 230.0 + float64(i)*40.0

		if icon != nil {
			dc.DrawImage(icon, 37, int(y) + 6)
		} else {
			dc.SetColor(color.RGBA{52, 211, 153, 255})
			dc.DrawCircle(45.0, y+11, 5)
			dc.Fill()
		}

		fontFaceBold(dc, 12)
		dc.SetColor(color.RGBA{220, 220, 240, 255})
		dc.DrawStringAnchored(a.Name, 58.0, y+11, 0, 0.5)

		fontFace(dc, 10)
		dc.SetColor(color.RGBA{140, 140, 175, 255})
		dc.DrawStringAnchored(a.Description, 58.0, y+26, 0, 0.5)
	}
}

func (w *WrapperImage) drawAppBars(dc *gg.Context, today *DailyRecord, hiddenApps []string) {
	// Build bar data
	type appEntry struct {
		Name     string
		Seconds  int64
		ColorIdx int
	}

	merged := make(map[string]int64)
	for name, usage := range today.Apps {
		display := shortAppName(name)
		merged[display] += usage.TotalSeconds
	}

	var apps []appEntry
	i := 0
	for name, sec := range merged {
		if sec < 60 {
			continue
		}
		// Filter out hidden apps (case-insensitive substring match)
		skip := false
		for _, hidden := range hiddenApps {
			if strings.Contains(strings.ToLower(name), strings.ToLower(hidden)) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		apps = append(apps, appEntry{
			Name:     name,
			Seconds:  sec,
			ColorIdx: i,
		})
		i++
	}

	// Sort by usage descending
	sort.Slice(apps, func(a, b int) bool {
		return apps[a].Seconds > apps[b].Seconds
	})

	// Take top 8
	const maxBars = 8
	if len(apps) > maxBars {
		apps = apps[:maxBars]
	}

	dc.SetColor(color.RGBA{200, 200, 230, 255})
	fontFaceBold(dc, 16)
	dc.DrawStringAnchored("Top Applications (All Time)", 740, 205, 0.5, 0.5)

	if len(apps) == 0 {
		fontFace(dc, 14)
		dc.SetColor(color.RGBA{120, 120, 160, 255})
		dc.DrawStringAnchored("No data yet", 740, 250, 0.5, 0.5)
		return
	}

	barStartY := 230.0
	barH := 30.0
	barGap := 6.0
	barStartX := 520.0
	barMaxW := 420.0

	maxSec := apps[0].Seconds
	if maxSec == 0 {
		maxSec = 1
	}

	fontFace(dc, 11)

	for i, app := range apps {
		y := barStartY + float64(i)*(barH+barGap)
		pct := float64(app.Seconds) / float64(maxSec)
		barW := pct * barMaxW
		if barW < 30 {
			barW = 30
		}

		// Bar background
		dc.SetColor(color.RGBA{30, 30, 55, 200})
		dc.DrawRoundedRectangle(barStartX, y, barMaxW, barH, 5)
		dc.Fill()

		// Bar fill
		clr := appColors[app.ColorIdx%len(appColors)]
		dc.SetColor(clr)
		dc.DrawRoundedRectangle(barStartX, y, barW, barH, 5)
		dc.Fill()

		// App name on bar (left)
		dc.SetColor(color.RGBA{255, 255, 255, 220})
		dc.DrawStringAnchored(app.Name, barStartX+6, y+barH/2, 0, 0.5)

		// Duration text on bar (right)
		dc.DrawStringAnchored(formatDuration(app.Seconds), barStartX+barW-6, y+barH/2, 1, 0.5)
	}
}

func (w *WrapperImage) drawAvatar(dc *gg.Context, avatarPath string) {
	if avatarPath == "" {
		return
	}
	img, err := gg.LoadImage(avatarPath)
	if err != nil {
		log.Printf("[renderer] avatar load error: %v", err)
		return
	}

	cx := 48.0
	cy := 44.0
	r := 22.0

	bounds := img.Bounds()
	scale := (r * 2) / float64(min(bounds.Dx(), bounds.Dy()))

	// Temporary context for circular crop
	ts := int(r*2) + 4
	tc := gg.NewContext(ts, ts)
	tc.DrawCircle(float64(ts)/2, float64(ts)/2, r)
	tc.Clip()
	tc.ScaleAbout(scale, scale, float64(ts)/2, float64(ts)/2)
	tc.DrawImageAnchored(img, ts/2, ts/2, 0.5, 0.5)
	tc.Identity()

	dc.DrawImageAnchored(tc.Image(), int(cx), int(cy), 0.5, 0.5)

	// Circular border
	dc.SetColor(color.RGBA{99, 102, 241, 200})
	dc.SetLineWidth(2)
	dc.DrawCircle(cx, cy, r)
	dc.Stroke()
}

func (w *WrapperImage) drawFooter(dc *gg.Context, streak, longestStreak, score int, dyingScore float64) {
	// Bottom panel with gradient
	panelY := 600.0
	panelH := 120.0
	panelPad := 60.0
	grad := gg.NewLinearGradient(panelPad, panelY, float64(w.width)-panelPad, panelY)
	grad.AddColorStop(0, color.RGBA{20, 40, 30, 180})
	grad.AddColorStop(0.5, color.RGBA{25, 25, 55, 180})
	grad.AddColorStop(1, color.RGBA{40, 20, 40, 180})
	dc.SetFillStyle(grad)
	dc.DrawRoundedRectangle(panelPad, panelY, float64(w.width)-panelPad*2, panelH, 16)
	dc.Fill()

	dc.SetColor(color.RGBA{60, 60, 100, 120})
	dc.SetLineWidth(1)
	dc.DrawRoundedRectangle(panelPad, panelY, float64(w.width)-panelPad*2, panelH, 16)
	dc.Stroke()

	// Left: PC Usage Score
	leftX := 120.0
	scoreStr := formatScore(score)
	fontSize := 44.0
	if len(scoreStr) >= 5 {
		fontSize = 26.0
	} else if len(scoreStr) >= 4 {
		fontSize = 32.0
	} else if len(scoreStr) >= 3 {
		fontSize = 38.0
	}
	fontFaceBold(dc, fontSize)
	dc.SetColor(color.RGBA{255, 255, 255, 255})
	dc.DrawStringAnchored(scoreStr, leftX, panelY+32, 0, 0.5)

	fontFace(dc, 12)
	dc.SetColor(color.RGBA{150, 150, 190, 255})
	dc.DrawStringAnchored("PC USAGE SCORE", leftX, panelY+54, 0, 0.5)

	if msg := randomScoreMessage(score); msg != "" {
		fontFaceBold(dc, 15)
		dc.SetColor(color.RGBA{200, 200, 230, 255})
		dc.DrawStringAnchored(msg, leftX, panelY+80, 0, 0.5)
	}

	// Right: PC Dying Score
	rightX := float64(w.width) - 120.0
	dyingStr := fmt.Sprintf("%.0f%%", dyingScore)
	fontFaceBold(dc, 36)
	dc.SetColor(color.RGBA{244, 114, 182, 230})
	dc.DrawStringAnchored(dyingStr, rightX, panelY+32, 1, 0.5)

	fontFaceBold(dc, 12)
	dc.SetColor(color.RGBA{200, 130, 170, 220})
	dc.DrawStringAnchored("PC DYING SCORE", rightX, panelY+54, 1, 0.5)

	// Separator
	dc.SetColor(color.RGBA{60, 60, 90, 255})
	dc.SetLineWidth(1)
	dc.DrawLine(60, 745, float64(w.width)-60, 745)
	dc.Stroke()

	fontFace(dc, 10)
	dc.SetColor(color.RGBA{100, 100, 140, 255})
	dc.DrawStringAnchored(fmt.Sprintf("Generated %s · %s · NoGrassWrapper %s", time.Now().Format("2006-01-02 15:04"), "github.com/Totgocpro/NoGrassWrapper", Version), float64(w.width)/2, 770, 0.5, 0.5)
}

// shortAppName shortens a window/app name for display in the bar chart.
func shortAppName(title string) string {
	name := extractAppName(title)
	if len(name) > 12 {
		return name[:10] + ".."
	}
	return name
}
