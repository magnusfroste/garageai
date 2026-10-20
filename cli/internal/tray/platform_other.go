//go:build !darwin && !windows

package tray

import "errors"

// Supported: no icon on Linux (servers); maybe AppIndicator later for Linux desktops.
const Supported = false

var errUnsupported = errors.New("the GarageAI icon is for macOS and Windows; on Linux use garageai doctor")

func Run(string) error            { return errUnsupported }
func LoginItemInstalled() bool    { return false }
func InstallLoginItem(bool) error { return errUnsupported }
func RemoveLoginItem(bool) error  { return nil }
func openURL(string) error        { return errUnsupported }
func openDoctor() error           { return errUnsupported }
