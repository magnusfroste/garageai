package tray

import (
	"errors"
	"os"
)

// Do runs a menu action; it reports whether the icon should quit. Every action is fixed here:
// nothing a file says is opened or run.
func Do(action string) (quit bool) {
	switch action {
	case "open-garage":
		_ = openURL(MyGarageURL)
	case "connect":
		_ = openURL(ConnectURL)
	case "doctor":
		_ = openDoctor()
	case "login":
		if LoginItemInstalled() {
			_ = RemoveLoginItem(false)
		} else {
			_ = InstallLoginItem(false)
		}
	case "hide":
		_ = RemoveLoginItem(false)
		return true
	}
	return false
}

// self is the binary the login item and the doctor window run.
func self() string {
	p, err := os.Executable()
	if err != nil {
		return "garageai"
	}
	return p
}

// ErrRunning: the icon already runs for this user.
var ErrRunning = errors.New("the GarageAI icon is already running")
