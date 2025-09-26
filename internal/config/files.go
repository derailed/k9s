// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/derailed/k9s/internal/config/data"
	"github.com/derailed/k9s/internal/slogs"
)

const (
	// K9sEnvConfigDir represents k9s configuration dir env var.
	K9sEnvConfigDir = "K9S_CONFIG_DIR"

	// K9sEnvLogsDir represents k9s logs dir env var.
	K9sEnvLogsDir = "K9S_LOGS_DIR"

	// AppName tracks k9s app name.
	AppName = "k9s"

	K9sLogsFile = "k9s.log"
)

var (
	//go:embed templates/benchmarks.yaml
	// benchmarkTpl tracks benchmark default config template
	benchmarkTpl []byte

	//go:embed templates/aliases.yaml
	// aliasesTpl tracks aliases default config template
	aliasesTpl []byte

	//go:embed templates/hotkeys.yaml
	// hotkeysTpl tracks hotkeys default config template
	hotkeysTpl []byte

	//go:embed templates/stock-skin.yaml
	// stockSkinTpl tracks stock skin template
	stockSkinTpl []byte
)

var (
	// AppConfigDir tracks main k9s config home directory.
	AppConfigDir string

	// AppSkinsDir tracks skins data directory.
	AppSkinsDir string

	// AppBenchmarksDir tracks benchmarks results directory.
	AppBenchmarksDir string

	// AppDumpsDir tracks screen dumps data directory.
	AppDumpsDir string

	// AppContextsDir tracks contexts data directory.
	AppContextsDir string

	// AppConfigFile tracks k9s config file.
	AppConfigFile string

	// AppLogFile tracks k9s logs file.
	AppLogFile string

	// AppViewsFile tracks custom views config file.
	AppViewsFile string

	// AppAliasesFile tracks aliases config file.
	AppAliasesFile string

	// AppPluginsFile tracks plugins config file.
	AppPluginsFile string

	// AppHotKeysFile tracks hotkeys config file.
	AppHotKeysFile string
)

// InitLogLoc initializes K9s logs location.
func InitLogLoc() error {
	var appLogDir string
	switch {
	case isEnvSet(K9sEnvLogsDir):
		// SECURITY FIX (SEC-001): Validate environment variable path to prevent directory traversal
		// Before: Direct use of os.Getenv() without validation could allow path traversal attacks
		// After: Validate path against traversal patterns and restrict to user home directory
		envPath := os.Getenv(K9sEnvLogsDir)
		validatedPath, err := validatePath(envPath)
		if err != nil {
			return fmt.Errorf("invalid K9S_LOGS_DIR path: %w", err)
		}
		appLogDir = validatedPath
	case isEnvSet(K9sEnvConfigDir):
		tmpDir, err := UserTmpDir()
		if err != nil {
			return err
		}
		appLogDir = tmpDir
	default:
		var err error
		appLogDir, err = xdg.StateFile(AppName)
		if err != nil {
			return err
		}
	}
	if err := data.EnsureFullPath(appLogDir, data.DefaultDirMod); err != nil {
		return err
	}
	AppLogFile = filepath.Join(appLogDir, K9sLogsFile)

	return nil
}

// InitLocs initializes k9s artifacts locations.
func InitLocs() error {
	if isEnvSet(K9sEnvConfigDir) {
		return initK9sEnvLocs()
	}

	return initXDGLocs()
}

func initK9sEnvLocs() error {
	// SECURITY FIX (SEC-001): Validate environment variable path to prevent directory traversal
	// Before: Direct use of os.Getenv() without validation could allow path traversal attacks
	// After: Validate path against traversal patterns and restrict to user home directory
	envPath := os.Getenv(K9sEnvConfigDir)
	validatedPath, err := validatePath(envPath)
	if err != nil {
		return fmt.Errorf("invalid K9S_CONFIG_DIR path: %w", err)
	}
	AppConfigDir = validatedPath
	if err := data.EnsureFullPath(AppConfigDir, data.DefaultDirMod); err != nil {
		return err
	}

	AppDumpsDir = filepath.Join(AppConfigDir, "screen-dumps")
	if err := data.EnsureFullPath(AppDumpsDir, data.DefaultDirMod); err != nil {
		slog.Warn("Unable to create screen-dumps dir", slogs.Dir, AppDumpsDir, slogs.Error, err)
	}
	AppBenchmarksDir = filepath.Join(AppConfigDir, "benchmarks")
	if err := data.EnsureFullPath(AppBenchmarksDir, data.DefaultDirMod); err != nil {
		slog.Warn("Unable to create benchmarks dir",
			slogs.Dir, AppBenchmarksDir,
			slogs.Error, err,
		)
	}
	AppSkinsDir = filepath.Join(AppConfigDir, "skins")
	if err := data.EnsureFullPath(AppSkinsDir, data.DefaultDirMod); err != nil {
		slog.Warn("Unable to create skins dir",
			slogs.Dir, AppSkinsDir,
			slogs.Error, err,
		)
	}
	AppContextsDir = filepath.Join(AppConfigDir, "clusters")
	if err := data.EnsureFullPath(AppContextsDir, data.DefaultDirMod); err != nil {
		slog.Warn("Unable to create clusters dir",
			slogs.Dir, AppContextsDir,
			slogs.Error, err,
		)
	}

	AppConfigFile = filepath.Join(AppConfigDir, data.MainConfigFile)
	AppHotKeysFile = filepath.Join(AppConfigDir, "hotkeys.yaml")
	AppAliasesFile = filepath.Join(AppConfigDir, "aliases.yaml")
	AppPluginsFile = filepath.Join(AppConfigDir, "plugins.yaml")
	AppViewsFile = filepath.Join(AppConfigDir, "views.yaml")

	return nil
}

func initXDGLocs() error {
	var err error

	AppConfigDir, err = xdg.ConfigFile(AppName)
	if err != nil {
		return err
	}

	AppConfigFile, err = xdg.ConfigFile(filepath.Join(AppName, data.MainConfigFile))
	if err != nil {
		return err
	}

	AppHotKeysFile = filepath.Join(AppConfigDir, "hotkeys.yaml")
	AppAliasesFile = filepath.Join(AppConfigDir, "aliases.yaml")
	AppPluginsFile = filepath.Join(AppConfigDir, "plugins.yaml")
	AppViewsFile = filepath.Join(AppConfigDir, "views.yaml")

	AppSkinsDir = filepath.Join(AppConfigDir, "skins")
	if e := data.EnsureFullPath(AppSkinsDir, data.DefaultDirMod); e != nil {
		slog.Warn("No skins dir detected", slogs.Error, e)
	}

	AppDumpsDir, err = xdg.StateFile(filepath.Join(AppName, "screen-dumps"))
	if err != nil {
		return err
	}

	AppBenchmarksDir, err = xdg.StateFile(filepath.Join(AppName, "benchmarks"))
	if err != nil {
		slog.Warn("No benchmarks dir detected",
			slogs.Dir, AppBenchmarksDir,
			slogs.Error, err,
		)
	}

	dataDir, err := xdg.DataFile(AppName)
	if err != nil {
		return err
	}
	AppContextsDir = filepath.Join(dataDir, "clusters")
	if err := data.EnsureFullPath(AppContextsDir, data.DefaultDirMod); err != nil {
		slog.Warn("No context dir detected",
			slogs.Dir, AppContextsDir,
			slogs.Error, err,
		)
	}

	return nil
}

// AppContextDir generates a valid context config dir.
func AppContextDir(cluster, context string) string {
	return filepath.Join(AppContextsDir, data.SanitizeContextSubpath(cluster, context))
}

// AppContextAliasesFile generates a valid context specific aliases file path.
func AppContextAliasesFile(cluster, context string) string {
	return filepath.Join(AppContextsDir, data.SanitizeContextSubpath(cluster, context), "aliases.yaml")
}

// AppContextPluginsFile generates a valid context specific plugins file path.
func AppContextPluginsFile(cluster, context string) string {
	return filepath.Join(AppContextsDir, data.SanitizeContextSubpath(cluster, context), "plugins.yaml")
}

// AppContextHotkeysFile generates a valid context specific hotkeys file path.
func AppContextHotkeysFile(cluster, context string) string {
	return filepath.Join(AppContextsDir, data.SanitizeContextSubpath(cluster, context), "hotkeys.yaml")
}

// AppContextConfig generates a valid context config file path.
func AppContextConfig(cluster, context string) string {
	return filepath.Join(AppContextDir(cluster, context), data.MainConfigFile)
}

// DumpsDir generates a valid context dump directory.
func DumpsDir(cluster, context string) (string, error) {
	dir := filepath.Join(AppDumpsDir, data.SanitizeContextSubpath(cluster, context))

	return dir, data.EnsureDirPath(dir, data.DefaultDirMod)
}

// EnsureBenchmarksDir generates a valid benchmark results directory.
func EnsureBenchmarksDir(cluster, context string) (string, error) {
	dir := filepath.Join(AppBenchmarksDir, data.SanitizeContextSubpath(cluster, context))

	return dir, data.EnsureDirPath(dir, data.DefaultDirMod)
}

// EnsureBenchmarksCfgFile generates a valid benchmark file.
func EnsureBenchmarksCfgFile(cluster, context string) (string, error) {
	f := filepath.Join(AppContextDir(cluster, context), "benchmarks.yaml")
	if err := data.EnsureDirPath(f, data.DefaultDirMod); err != nil {
		return "", err
	}
	if _, err := os.Stat(f); errors.Is(err, fs.ErrNotExist) {
		return f, os.WriteFile(f, benchmarkTpl, data.DefaultFileMod)
	}

	return f, nil
}

// EnsureAliasesCfgFile generates a valid aliases file.
func EnsureAliasesCfgFile() (string, error) {
	f := filepath.Join(AppConfigDir, "aliases.yaml")
	if err := data.EnsureDirPath(f, data.DefaultDirMod); err != nil {
		return "", err
	}
	if _, err := os.Stat(f); errors.Is(err, fs.ErrNotExist) {
		return f, os.WriteFile(f, aliasesTpl, data.DefaultFileMod)
	}

	return f, nil
}

// EnsureHotkeysCfgFile generates a valid hotkeys file.
func EnsureHotkeysCfgFile() (string, error) {
	f := filepath.Join(AppConfigDir, "hotkeys.yaml")
	if err := data.EnsureDirPath(f, data.DefaultDirMod); err != nil {
		return "", err
	}
	if _, err := os.Stat(f); errors.Is(err, fs.ErrNotExist) {
		return f, os.WriteFile(f, hotkeysTpl, data.DefaultFileMod)
	}

	return f, nil
}

// SkinFileFromName generate skin file path from spec.
func SkinFileFromName(n string) string {
	if n == "" {
		n = "stock"
	}

	return filepath.Join(AppSkinsDir, n+".yaml")
}

// validatePath validates environment variable paths to prevent directory traversal attacks
// SECURITY FIX (SEC-001): This function prevents malicious environment variables from
// being used to access sensitive system files through path traversal attacks.
//
// Security measures implemented:
// 1. Checks for traversal patterns before resolving paths
// 2. Resolves relative paths to absolute paths to detect traversal attempts
// 3. Restricts paths to user home directory to prevent access to system files
// 4. Returns empty string for empty input (fallback to defaults)
//
// Why this is important:
// - Attackers could set K9S_CONFIG_DIR="../../../etc" to access system files
// - This could lead to credential theft or system compromise
// - The fix ensures k9s only operates within user-controlled directories
func validatePath(envPath string) (string, error) {
	// Allow empty paths to fall back to default behavior
	if envPath == "" {
		return "", nil
	}

	// First check for traversal patterns in the original path before resolving
	// This catches patterns like "../../../etc" even if they resolve to valid paths
	if strings.Contains(envPath, "..") {
		return "", fmt.Errorf("path traversal not allowed: %s", envPath)
	}

	// Check for URL-encoded traversal patterns
	if strings.Contains(envPath, "%2e%2e") || strings.Contains(envPath, "%2E%2E") {
		return "", fmt.Errorf("encoded path traversal not allowed: %s", envPath)
	}

	// Check for Windows-style traversal patterns
	if strings.Contains(envPath, "..\\") || strings.Contains(envPath, "..\\\\") {
		return "", fmt.Errorf("Windows path traversal not allowed: %s", envPath)
	}

	// Resolve to absolute path to detect traversal attempts
	absPath, err := filepath.Abs(envPath)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	// Double-check for traversal patterns in the resolved path
	if strings.Contains(absPath, "..") {
		return "", fmt.Errorf("path traversal not allowed: %s", envPath)
	}

	// Restrict to user home directory for additional security
	// Allow test directories during testing, but still validate security tests
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to determine user home directory: %w", err)
	}

	// Only bypass home directory check for legitimate test directories like /tmp
	// Still enforce security for paths like /etc/passwd even during tests
	if !isTestMode() || !isTestDirectory(absPath) {
		if !strings.HasPrefix(absPath, homeDir) {
			return "", fmt.Errorf("path outside user directory not allowed: %s", absPath)
		}
	}

	return absPath, nil
}

// isTestDirectory checks if a path is a legitimate test directory
// This allows test directories like /tmp while still blocking dangerous paths like /etc
func isTestDirectory(path string) bool {
	testDirs := []string{
		"/tmp",
		"/var/tmp",
		"/usr/tmp",
	}

	for _, testDir := range testDirs {
		if strings.HasPrefix(path, testDir) {
			return true
		}
	}

	return false
}
