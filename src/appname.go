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

// knownSites maps common page titles found in browser tab titles to their domain.
// Browser tabs show the <title> tag content, not the URL, so we map known names to domains.
var knownSites = map[string]string{
	"netflix":         "netflix.com",
	"youtube":         "youtube.com",
	"github":          "github.com",
	"reddit":          "reddit.com",
	"twitter":         "x.com",
	"facebook":        "facebook.com",
	"instagram":       "instagram.com",
	"linkedin":        "linkedin.com",
	"amazon":          "amazon.com",
	"stack overflow":  "stackoverflow.com",
	"stackoverflow":   "stackoverflow.com",
	"wikipedia":       "wikipedia.org",
	"discord":         "discord.com",
	"slack":           "slack.com",
	"figma":           "figma.com",
	"chatgpt":         "chatgpt.com",
	"chat gpt":        "chatgpt.com",
	"outlook":         "outlook.live.com",
	"teams":           "teams.microsoft.com",
	"microsoft teams": "teams.microsoft.com",
	"gmail":           "gmail.com",
	"drive":           "drive.google.com",
	"google drive":    "drive.google.com",
	"google docs":     "docs.google.com",
	"google meet":     "meet.google.com",
	"google maps":     "maps.google.com",
	"maps":            "maps.google.com",
	"spotify":         "open.spotify.com",
	"twitch":          "twitch.tv",
	"whatsapp":        "whatsapp.com",
	"telegram":        "telegram.org",
	"messenger":       "messenger.com",
	"zoom":            "zoom.us",
	"canva":           "canva.com",
	"trello":          "trello.com",
	"jira":            "atlassian.net",
	"confluence":      "atlassian.net",
	"gitlab":          "gitlab.com",
	"bitbucket":       "bitbucket.org",
	"medium":          "medium.com",
	"dev.to":          "dev.to",
	"hacker news":     "news.ycombinator.com",
	"news.ycombinator.com": "news.ycombinator.com",
	"dribbble":        "dribbble.com",
	"behance":         "behance.net",
	"pinterest":       "pinterest.com",
	"tumblr":          "tumblr.com",
	"ebay":            "ebay.com",
	"etsy":            "etsy.com",
	"shopify":         "shopify.com",
	"dropbox":         "dropbox.com",
	"onedrive":        "onedrive.live.com",
	"icloud":          "icloud.com",
	"microsoft 365":   "office.com",
	"office":          "office.com",
	"wordpress":       "wordpress.com",
	"google analytics": "analytics.google.com",
	"analytics":       "analytics.google.com",
	"google ads":      "ads.google.com",
	"google cloud":    "console.cloud.google.com",
	"aws":             "aws.amazon.com",
	"amazon aws":      "aws.amazon.com",
	"azure":           "portal.azure.com",
	"digitalocean":    "digitalocean.com",
	"cloudflare":      "cloudflare.com",
	"netlify":         "netlify.com",
	"vercel":          "vercel.com",
	"heroku":          "heroku.com",
	"railway":         "railway.app",
	"render":          "render.com",
	"fly.io":          "fly.io",
	"supabase":        "supabase.com",
	"mongodb":         "mongodb.com",
	"postman":         "postman.com",
	"linear":          "linear.app",
	"notion":          "notion.so",
	"obsidian":        "obsidian.md",
	"perplexity":      "perplexity.ai",
	"claude":          "claude.ai",
	"deepseek":        "chat.deepseek.com",
	"gemini":          "gemini.google.com",
	"copilot":         "github.com/copilot",
	"duckduckgo":      "duckduckgo.com",
	"bing":            "bing.com",
	"baidu":           "baidu.com",
	"yahoo":           "yahoo.com",
	"stackexchange":   "stackexchange.com",
	"super user":      "superuser.com",
	"server fault":    "serverfault.com",
	"ask ubuntu":      "askubuntu.com",
	"product hunt":    "producthunt.com",
	"indiehackers":    "indiehackers.com",
	"hackerone":       "hackerone.com",
	"bugcrowd":        "bugcrowd.com",
	"intigriti":       "intigriti.com",
	"tryhackme":       "tryhackme.com",
	"hack the box":    "hackthebox.com",
	"leetcode":        "leetcode.com",
	"codewars":        "codewars.com",
	"codecademy":      "codecademy.com",
	"udemy":           "udemy.com",
	"coursera":        "coursera.org",
	"edx":             "edx.org",
	"pluralsight":     "pluralsight.com",
	"frontend mentor": "frontendmentor.com",
	"css-tricks":      "css-tricks.com",
	"codepen":         "codepen.io",
	"jsfiddle":        "jsfiddle.net",
	"replit":          "replit.com",
	"codesandbox":     "codesandbox.io",
	"stackblitz":      "stackblitz.com",
	"glitch":          "glitch.com",
	"colab":           "colab.research.google.com",
	"google colab":    "colab.research.google.com",
	"kaggle":          "kaggle.com",
	"tableau":         "tableau.com",
	"grafana":         "grafana.com",
	"prometheus":      "prometheus.io",
	"datadog":         "datadoghq.com",
	"new relic":       "newrelic.com",
	"sentry":          "sentry.io",
	"rollbar":         "rollbar.com",
	"logz.io":         "logz.io",
	"papertrail":      "papertrailapp.com",
	"doppler":         "doppler.com",
	"1password":       "1password.com",
	"lastpass":        "lastpass.com",
	"bitwarden":       "bitwarden.com",
	"dashlane":        "dashlane.com",
	"authy":           "authy.com",
	"duo":             "duo.com",
	"okta":            "okta.com",
	"stripe":          "stripe.com",
	"paypal":          "paypal.com",
	"square":          "squareup.com",
	"wise":            "wise.com",
	"revolut":         "revolut.com",
	"coinbase":        "coinbase.com",
	"binance":         "binance.com",
	"kraken":          "kraken.com",
	"robinhood":       "robinhood.com",
	"etrade":          "etrade.com",
	"fidelity":        "fidelity.com",
	"vanguard":        "vanguard.com",
	"schwab":          "schwab.com",
	"airbnb":          "airbnb.com",
	"booking":         "booking.com",
	"tripadvisor":     "tripadvisor.com",
	"uber":            "uber.com",
	"lyft":            "lyft.com",
	"doordash":        "doordash.com",
	"grubhub":         "grubhub.com",
	"ubereats":        "ubereats.com",
	"yelp":            "yelp.com",
	"imdb":            "imdb.com",
	"rottentomatoes":  "rottentomatoes.com",
	"myanimelist":     "myanimelist.net",
	"crunchyroll":     "crunchyroll.com",
	"hulu":            "hulu.com",
	"hbomax":          "max.com",
	"max":             "max.com",
	"disney+":         "disneyplus.com",
	"disney plus":     "disneyplus.com",
	"peacock":         "peacocktv.com",
	"paramount+":      "paramountplus.com",
	"paramount plus":  "paramountplus.com",
}

// extractSiteName extracts a site/domain name from a browser window title.
// It returns the site name, or empty string if it cannot be determined.
// The title is the full window title (e.g. "Netflix - Google Chrome").
func extractSiteName(title string) string {
	// Known browser identifiers that might appear in window titles
	browserIDs := []string{
		"google chrome", "chromium", "brave", "mozilla firefox",
		"firefox", "safari", "edge", "opera", "vivaldi", "browser",
	}

	// Try to split on common separators and find browser part vs page title part
	separators := []string{" — ", " – ", " | ", " - ", ": "}
	for _, sep := range separators {
		parts := strings.Split(title, sep)
		if len(parts) < 2 {
			continue
		}
		// Determine which part is the browser name
		browserPart := -1
		for i, part := range parts {
			lowerPart := strings.ToLower(strings.TrimSpace(part))
			for _, bid := range browserIDs {
				if strings.Contains(lowerPart, bid) {
					browserPart = i
					break
				}
			}
			if browserPart >= 0 {
				break
			}
		}
		if browserPart < 0 {
			continue
		}
		// The other part(s) are the page title
		var pageTitleParts []string
		for i, part := range parts {
			if i != browserPart {
				pageTitleParts = append(pageTitleParts, strings.TrimSpace(part))
			}
		}
		pageTitle := strings.Join(pageTitleParts, sep)
		if pageTitle == "" {
			continue
		}
		return cleanSiteName(pageTitle)
	}

	// If no separator found, try to find a domain in the entire title
	if site := extractDomainFromString(title); site != "" {
		return site
	}

	// Last resort: return the first meaningful word
	return ""
}

// cleanSiteName tries to extract a clean site/domain name from a page title.
func cleanSiteName(title string) string {
	// Try to extract domain from the title first
	if site := extractDomainFromString(title); site != "" {
		return site
	}

	lower := strings.ToLower(strings.TrimSpace(title))

	// Check knownSites map (exact or prefix match — the title is often just the site name)
	if domain, ok := knownSites[lower]; ok {
		return domain
	}
	// Check first word only (common pattern: "YouTube Music" -> youtube.com)
	if idx := strings.IndexByte(lower, ' '); idx > 0 {
		firstWord := lower[:idx]
		if domain, ok := knownSites[firstWord]; ok {
			return domain
		}
	}

	// Remove common page title suffixes
	title = strings.TrimSpace(title)
	title = strings.TrimSuffix(title, " - Home")
	title = strings.TrimSuffix(title, " | Home")
	title = strings.TrimSuffix(title, " - Dashboard")
	title = strings.TrimSuffix(title, " | Dashboard")

	// If the remaining title is short enough, use it as-is
	if len(title) <= 25 {
		if title != "" {
			return title
		}
	}

	// Use the first meaningful word(s)
	words := strings.Fields(title)
	if len(words) == 0 {
		return ""
	}
	// Take up to 3 words, max 25 chars
	result := words[0]
	for i := 1; i < len(words) && i < 3 && len(result)+len(words[i])+1 <= 25; i++ {
		result += " " + words[i]
	}
	if len(result) > 25 {
		result = result[:22] + "..."
	}
	return result
}

// extractDomainFromString looks for a domain-like pattern in a string.
func extractDomainFromString(s string) string {
	lower := strings.ToLower(s)
	tlds := []string{".com", ".org", ".net", ".io", ".fr", ".de", ".uk",
		".app", ".dev", ".me", ".tv", ".co", ".ru", ".jp", ".cn",
		".gov", ".edu", ".info", ".biz", ".xyz", ".club", ".online",
		".social", "blog", ".wiki", ".news", ".app", ".dev"}
	for _, tld := range tlds {
		idx := strings.Index(lower, tld)
		if idx < 0 {
			continue
		}
		// Walk backwards to find domain start
		start := idx
		for start > 0 {
			c := lower[start-1]
			if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '.' {
				start--
			} else {
				break
			}
		}
		// Walk forward past the TLD
		end := idx + len(tld)
		for end < len(lower) {
			c := lower[end]
			if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '.' {
				end++
			} else {
				break
			}
		}
		domain := lower[start:end]
		// Clean up the domain
		domain = strings.TrimPrefix(domain, "www.")
		if domain != "" && strings.Contains(domain, ".") {
			return domain
		}
	}
	return ""
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
