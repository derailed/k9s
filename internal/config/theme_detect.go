// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"os/exec"
	"runtime"
	"strings"
)

// IsDarkMode reports whether the OS is currently in dark mode.
func IsDarkMode() bool {
	switch runtime.GOOS {
	case "darwin":
		out, err := exec.Command("defaults", "read", "-g", "AppleInterfaceStyle").Output()
		return err == nil && strings.TrimSpace(string(out)) == "Dark"
	case "linux":
		// ponytail: GNOME only via gsettings; add KDE/xdg-desktop-portal support when needed
		out, err := exec.Command("gsettings", "get", "org.gnome.desktop.interface", "color-scheme").Output()
		return err == nil && strings.Contains(string(out), "dark")
	case "windows":
		out, err := exec.Command("reg", "query",
			`HKCU\Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`,
			"/v", "AppsUseLightTheme").Output()
		return err == nil && strings.Contains(string(out), "0x0")
	}
	return false
}
