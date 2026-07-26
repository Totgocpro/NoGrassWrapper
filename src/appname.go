package main

import "strings"

var knownApps = map[string]string{
	"firefox":             "Firefox",
	"mozilla":             "Firefox",
	"isoufflette":         "Firefox",
	"chrom":               "Chrome",
	"google-chrome":       "Chrome",
	"brave":               "Brave",
	"code":                "VS Code",
	"visual studio":       "VS Code",
	"vscodium":            "VSCodium",
	"terminal":            "Terminal",
	"alacritty":           "Terminal",
	"kitty":               "Kitty",
	"wezterm":             "WezTerm",
	"foot":                "Foot",
	"discord":             "Discord",
	"slack":               "Slack",
	"spotify":             "Spotify",
	"obsidian":            "Obsidian",
	"nvim":                "Neovim",
	"vim":                 "Vim",
	"neovide":             "Neovim",
	"emacs":               "Emacs",
	"libreoffice":         "LibreOffice",
	"writer":              "LibreOffice",
	"calc":                "LibreOffice",
	"impress":             "LibreOffice",
	"draw":                "LibreOffice",
	"base":                "LibreOffice",
	"thunderbird":         "Thunderbird",
	"evolution":           "Evolution",
	"kontact":             "Kontact",
	"kate":                "Kate",
	"kwrite":              "Kate",
	"okular":              "Okular",
	"gwenview":            "Gwenview",
	"dolphin":             "Dolphin",
	"konsole":             "Konsole",
	"yakuake":             "Yakuake",
	"telegram":            "Telegram",
	"signal":              "Signal",
	"whatsapp":            "WhatsApp",
	"steam":               "Steam",
	"lutris":              "Lutris",
	"heroic":              "Heroic",
	"gimp":                "GIMP",
	"inkscape":            "Inkscape",
	"krita":               "Krita",
	"blender":             "Blender",
	"idea":                "IntelliJ",
	"intellij":            "IntelliJ",
	"pycharm":             "PyCharm",
	"goland":              "GoLand",
	"webstorm":            "WebStorm",
	"eclipse":             "Eclipse",
	"android studio":      "Android Studio",
	"thunar":              "Thunar",
	"nautilus":            "Nautilus",
	"nemo":                "Nemo",
	"pcmanfm":             "PCManFM",
	"rhythmbox":           "Rhythmbox",
	"vlc":                 "VLC",
	"mpv":                 "MPV",
	"celluloid":           "Celluloid",
	"obs":                 "OBS",
	"audacity":            "Audacity",
	"virt-manager":        "Virt-Manager",
	"virtualbox":          "VirtualBox",
	"vmware":              "VMware",
	"element":             "Element",
	"ferdium":             "Ferdium",
	"postman":             "Postman",
	"insomnia":            "Insomnia",
	"figma":               "Figma",
	"telegramdesktop":     "Telegram",
	"org.kde.kate":        "Kate",
	"org.kde.konsole":     "Konsole",
	"org.kde.dolphin":     "Dolphin",
	"org.kde.gwenview":    "Gwenview",
	"org.kde.okular":      "Okular",
	"org.kde.kontact":     "Kontact",
	"org.kde.yakuake":     "Yakuake",
	"outlook":             "Outlook",
	"word":                "Microsoft Word",
	"excel":               "Microsoft Excel",
	"powerpoint":          "Microsoft PowerPoint",
	"teams":               "Microsoft Teams",
	"mstsc":               "Remote Desktop",
	"cmd":                 "Command Prompt",
	"powershell":          "PowerShell",
	"explorer":            "File Explorer",
	"winrar":              "WinRAR",
	"7zip":                "7-Zip",
	"notepad++":           "Notepad++",
	"sublime":             "Sublime Text",
	"atom":                "Atom",
	"xcode":               "Xcode",
	"finder":              "Finder",
	"notes":               "Notes",
	"mail":                "Mail",
	"calendar":            "Calendar",
	"safari":              "Safari",
	"messages":            "Messages",
	"music":               "Music",
	"photos":              "Photos",
	"system preferences":  "System Preferences",
	"system settings":     "System Settings",
}

// extractAppName extracts the application name from a window title.
// Handles patterns like "Document - AppName", "AppName: Document", etc.
func extractAppName(title string) string {
	lower := strings.ToLower(title)

	// Direct known app match first
	for pattern, name := range knownApps {
		if strings.Contains(lower, pattern) {
			return name
		}
	}

	// Detect browser tabs: if title looks like a URL or web page, return "Browser"
	if looksLikeWebTab(title) {
		return "Browser"
	}

	// Split on common separators and check each part
	separators := []string{" — ", " – ", " | ", " - ", ": ", " — "}
	for _, sep := range separators {
		parts := strings.Split(title, sep)
		if len(parts) >= 2 {
			for _, part := range parts {
				trimmed := strings.TrimSpace(part)
				lowerPart := strings.ToLower(trimmed)
				for pattern, name := range knownApps {
					if strings.Contains(lowerPart, pattern) {
						return name
					}
				}
			}
			// If no known app in any part, use the last meaningful part
			last := strings.TrimSpace(parts[len(parts)-1])
			if !looksLikeFilename(last) {
				return last
			}
		}
	}

	return title
}

// looksLikeWebTab detects if a window title is likely a browser tab (URL or webpage).
func looksLikeWebTab(title string) bool {
	lower := strings.ToLower(title)
	// Common URL patterns in tab titles
	urlIndicators := []string{
		"www.", ".com", ".org", ".net", ".io", ".fr", ".de", ".uk",
		".app", ".dev", ".me", ".tv", ".co", ".ru", ".jp", ".cn",
		"http://", "https://",
	}
	count := 0
	for _, ind := range urlIndicators {
		if strings.Contains(lower, ind) {
			count++
		}
	}
	// Titles with multiple URL indicators,
	// "save", "download", etc. are likely web tabs
	webActions := []string{
		"download", "save", "upload", "login", "sign in", "sign up",
		"register",
	}
	for _, action := range webActions {
		if strings.Contains(lower, action) {
			return true
		}
	}
	return count >= 1
}

// looksLikeFilename checks if a string looks like a filename rather than an app name.
func looksLikeFilename(s string) bool {
	exts := []string{".txt", ".md", ".go", ".py", ".js", ".ts", ".html", ".css",
		".json", ".xml", ".yaml", ".yml", ".toml", ".ini", ".cfg", ".conf",
		".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".pdf", ".doc", ".docx",
		".xls", ".xlsx", ".ppt", ".pptx", ".odt", ".ods", ".odp",
		".mp3", ".mp4", ".avi", ".mkv", ".mov", ".flac", ".wav",
		".zip", ".tar", ".gz", ".7z", ".rar",
		".sh", ".bat", ".ps1", ".exe", ".msi", ".deb", ".rpm",
		".log", ".tmp", ".swp", ".bak",
		".c", ".cpp", ".h", ".hpp", ".rs", ".rb", ".php", ".pl",
		".sql", ".db", ".sqlite",
	}
	lower := strings.ToLower(s)
	for _, ext := range exts {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return strings.Count(s, ".") > 1 || strings.Count(s, "/") > 0 || strings.Count(s, "\\") > 0
}
