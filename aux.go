package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

const linuxDesktopTemplate = `
	[Desktop Entry]
	Name=badger-gui
	Exec=%s
	Icon=badger-gui
	Type=Application
	Categories=Network;Social;
`

func setLinuxDesktopIcon(iconData []byte) {
	if runtime.GOOS != "linux" {
		return
	}
	if os.Getenv("SNAP") != "" { // snap package
		return
	}

	currentUser, err := user.Current()
	if err != nil {
		panic(err)
	}
	homeDir := currentUser.HomeDir

	desktopDir := filepath.Join(homeDir, ".local", "share", "applications")
	iconDir := filepath.Join(homeDir, ".local", "share", "icons", "hicolor", "512x512", "apps")

	_ = os.MkdirAll(desktopDir, 0755)
	_ = os.MkdirAll(iconDir, 0755)

	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("setting icon: unable to determine executable path: %v", err)
	}

	desktopFile := filepath.Join(desktopDir, "badger-gui.desktop")
	content := fmt.Sprintf(linuxDesktopTemplate, execPath)
	if err := os.WriteFile(desktopFile, []byte(content), 0644); err != nil {
		log.Fatalf("setting icon: write .desktop file fail: %v", err)
	}

	iconPath := filepath.Join(iconDir, "badger-gui.png")
	if err := os.WriteFile(iconPath, iconData, 0644); err != nil {
		log.Fatalf("setting icon: write icon file fail: %v", err)
	}
}

func getFile(filePath string) ([]byte, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return io.ReadAll(f)
}
