# NoGrassWrapper 🌿

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/license-GPL3.0-green)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey)](#)
[![Made With](https://img.shields.io/badge/Made%20With-Love%20%26%20Go-red)](#)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen)](#)
[![Coffee Required](https://img.shields.io/badge/Coffee%20Required-Yes-8B4513)](#)
[![Bugs](https://img.shields.io/badge/Bugs-0%20(that%20you%20know%20of)-yellow)](#)

> **Track your screen time, earn achievements, and judge your PC usage with style (or not).**

NoGrassWrapper runs in your system tray, silently tracking which applications you use throughout the day. Generate a stylish "wrapper" image showing your stats, achievements, and a complete breakdown of your PC usage — ready to share, flex, or reflect on. Show the world just how worthwhile it was to give you a PC.

---

## ✨ Features

- **⏱️ All-Time Tracking** — Cumulative stats across every session, not just daily.
- **🏆 Achievements** — From *Speedrunner* to *Dying PC*, unlock them all.
- **📊 Beautiful Wrappers** — 1000×800 PNG with app bars, stats cards, and your avatar.
- **📈 Hardware Monitoring** — CPU, RAM, and GPU usage tracked per day.
- **🖼️ Custom Avatar** — Upload your profile picture via the settings page.
- **🎯 Score System** — PC Usage Score with grades (F → SSS) and fun tiers.
- **💬 Auto-Deprecation** — Random cheeky messages based on your score.
- **🔔 Notifications** — Get notified when you unlock achievements.
- **🌐 Cross-Platform** — Windows, macOS, Linux with native system tray.
- **💾 Privacy** — Nothing leaves your computer (unless you share it).
- **🚀 Performance** — Uses less than 100 MB of RAM and places little load on your processor.

---

## 🚀 Installation

### Pre-built Binaries

Download the latest release from the [Releases page](https://github.com/Totgocpro/NoGrassWrapper/releases).

### Build from Source

```bash
# Clone the repository
git clone https://github.com/Totgocpro/NoGrassWrapper.git
cd NoGrassWrapper

# Build
go build -ldflags="-X main.Version=$(git describe --tags 2>/dev/null || echo dev)" -o nograsswrapper ./src/

# Run
./nograsswrapper
```

### Dependencies (Linux)

```bash
# X11 (required for window detection)
sudo apt install xdotool xprintidle

# Wayland (optional, for KDE Wayland)
sudo apt install kdotool
```

### Windows Installer

Download `NoGrassWrapper-Setup.exe` from releases. The installer will:

1. Ask for your **username** (optional)
2. Let you **choose an avatar image** (optional)
3. Offer to **auto-start** with Windows

---

## ⚙️ Configuration

### Settings Page

Click **Settings** in the system tray menu. A browser tab opens with:

- **Username** — displayed on the wrapper image
- **Avatar** — upload a profile picture (saved to `~/.config/nograsswrapper/avatar.*`)

### Environment Variables

| Variable          | Description                          |
|-------------------|--------------------------------------|
| `NOGRASS_AVATAR`  | Path to a default avatar image       |

### Data Location

| Platform | Path                                                     |
|----------|----------------------------------------------------------|
| Linux    | `~/.config/nograsswrapper/usage.json`                   |
| macOS    | `~/Library/Application Support/NoGrassWrapper/`          |
| Windows  | `%APPDATA%\NoGrassWrapper\usage.json`                   |

---

## 🧑‍💻 Development

```bash
# Run directly
go run ./src/

# Build with version
go build -ldflags="-X main.Version=$(git describe --tags 2>/dev/null || echo dev)" -o nograsswrapper ./src/

# Build for Windows from Linux/macOS
GOOS=windows GOARCH=amd64 go build -ldflags="-X main.Version=$(git describe --tags 2>/dev/null || echo dev)" -o nograsswrapper.exe ./src/

# Build for macOS from Linux/Windows
GOOS=darwin GOARCH=amd64 go build -ldflags="-X main.Version=$(git describe --tags 2>/dev/null || echo dev)" -o nograsswrapper ./src/
```

### Project Structure

```
nograsswrapper/
├── src/
│   ├── main.go              # Entry point
│   ├── types.go             # Data types, score, grades
│   ├── storage.go           # JSON persistence
│   ├── tracker.go           # Window + hardware monitoring
│   ├── tracker_linux.go     # Linux window/idle detection
│   ├── tracker_windows.go   # Windows window/idle detection
│   ├── tracker_darwin.go    # macOS window/idle detection
│   ├── tray.go              # System tray menu
│   ├── renderer.go          # PNG wrapper generation
│   ├── settings.go          # HTTP settings server
│   ├── achievements.go      # Achievement checks + notifications
│   ├── messages.go          # Score-based messages loader
│   ├── icon.go              # SVG icon embedding
│   └── assets/              # Embedded assets (SVGs, HTML)
│       ├── Icon.svg
│       ├── Achievements.svg
│       └── settings.html
├── installer/
│   └── windows.iss          # Inno Setup installer script
├── .github/
│   └── workflows/
│       ├── build.yml        # CI build workflow
│       └── release.yml      # Release workflow
├── go.mod
└── README.md
```

---

## 📜 License

GPL 3.0 — do whatever you want, just don't blame me for your screen time.

---

## 🙏 Credits

- [getlantern/systray](https://github.com/getlantern/systray) — cross-platform system tray
- [fogleman/gg](https://github.com/fogleman/gg) — 2D rendering library
- [shirou/gopsutil](https://github.com/shirou/gopsutil) — hardware monitoring
- [srwiley/oksvg](https://github.com/srwiley/oksvg) — SVG rasterization

---

*Made with ❤️ and too much screen time.*
