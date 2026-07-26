package main

import (
	_ "embed"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
)

//go:embed assets/settings.html
var settingsHTML []byte

var (
	settingsMu sync.Mutex
)

func showSettingsDialog(store *Storage) {
	if !settingsMu.TryLock() {
		log.Println("[settings] already open, ignoring")
		return
	}
	defer settingsMu.Unlock()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Println("[settings] port error:", err)
		return
	}
	port := listener.Addr().(*net.TCPAddr).Port
	log.Printf("[settings] open http://127.0.0.1:%d", port)

	done := make(chan struct{})

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(settingsHTML)
	})

	mux.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"username":"%s","avatar":"%s"}`, store.GetUsername(), store.GetAvatarPath())
	})

	mux.HandleFunc("/save", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, "parse error", 400)
			return
		}

		store.SetUsername(r.FormValue("username"))

		// If a file was uploaded, save it to config dir
		file, header, err := r.FormFile("avatar_file")
		if err == nil {
			defer file.Close()
			ext := filepath.Ext(header.Filename)
			if ext == "" {
				ext = ".png"
			}
			dir, _ := configDir()
			dest := filepath.Join(dir, "avatar"+ext)

			f, err := os.Create(dest)
			if err == nil {
				io.Copy(f, file)
				f.Close()
				store.SetAvatarPath(dest)
			}
		} else {
			path := r.FormValue("avatar_path")
			if path != "" {
				store.SetAvatarPath(path)
			}
		}

		_ = store.Save()
		w.Write([]byte("ok"))
		close(done)
	})

	mux.HandleFunc("/cancel", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
		close(done)
	})

	server := &http.Server{Handler: mux}
	go server.Serve(listener)

	url := fmt.Sprintf("http://127.0.0.1:%d", port)
	switch runtime.GOOS {
	case "darwin":
		exec.Command("open", url).Run()
	case "windows":
		exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Run()
	default:
		exec.Command("xdg-open", url).Run()
	}

	<-done

	server.Close()
	listener.Close()
	log.Println("[settings] closed")
}
