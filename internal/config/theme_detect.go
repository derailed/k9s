// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"os"
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
		if isWSL() {
			// WSL has no OS-level appearance of its own; it inherits the Windows host's.
			return windowsRegDarkMode("reg.exe")
		}
		// ponytail: GNOME only via gsettings; add KDE/xdg-desktop-portal support when needed
		out, err := exec.Command("gsettings", "get", "org.gnome.desktop.interface", "color-scheme").Output()
		return err == nil && strings.Contains(string(out), "dark")
	case "windows":
		return windowsRegDarkMode("reg")
	}
	return false
}

// isWSL reports whether we're running inside Windows Subsystem for Linux.
func isWSL() bool {
	if os.Getenv("WSL_DISTRO_NAME") != "" {
		return true
	}
	b, err := os.ReadFile("/proc/version")
	return err == nil && strings.Contains(strings.ToLower(string(b)), "microsoft")
}

// windowsRegDarkMode queries the Windows dark-mode registry key via bin,
// which is "reg" natively on Windows or "reg.exe" reached through WSL interop.
func windowsRegDarkMode(bin string) bool {
	out, err := exec.Command(bin, "query",
		`HKCU\Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`,
		"/v", "AppsUseLightTheme").Output()
	return err == nil && strings.Contains(string(out), "0x0")
}
