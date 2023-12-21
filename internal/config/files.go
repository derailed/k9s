// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"os"
	"os/user"
	"path/filepath"
	"regexp"

	"github.com/derailed/k9s/internal/config/data"

	"github.com/adrg/xdg"
	"github.com/rs/zerolog/log"
)

const (
	// K9sConfigDir represents k9s configuration dir env var.
	K9sConfigDir = "K9S_CONFIG_DIR"

	// AppName tracks k9s app name.
	AppName = "k9s"

	K9sLogsFile = "k9s.log"
)

var (
	invalidPathCharsRX = regexp.MustCompile(`[:/]+`)

	AppConfigDir     string
	AppSkinsDir      string
	AppBenchmarksDir string
	AppDumpsDir      string
	AppContextsDir   string
	AppConfigFile    string
	AppLogFile       string
	AppViewsFile     string
	AppAliasesFile   string
	AppPluginsFile   string
	AppHotKeysFile   string
)

// InitLogsLoc initializes K9s logs location.
func InitLogLoc() error {
	if hasK9sConfigEnv() {
		tmpDir, err := userTmpDir()
		if err != nil {
			return err
		}
		AppLogFile = filepath.Join(tmpDir, K9sLogsFile)
		return nil
	}

	var err error
	AppLogFile, err = xdg.StateFile(filepath.Join(AppName, K9sLogsFile))

	return err
}

// InitLocs initializes k9s artifacts locations.
func InitLocs() error {
	if hasK9sConfigEnv() {
		return initK9sEnvLocs()
	}

	return initXDGLocs()
}

func initK9sEnvLocs() error {
	AppConfigDir = os.Getenv(K9sConfigDir)
	if err := data.EnsureFullPath(AppConfigDir, data.DefaultDirMod); err != nil {
		return err
	}

	AppDumpsDir = filepath.Join(AppConfigDir, "screen-dumps")
	if err := data.EnsureFullPath(AppDumpsDir, data.DefaultDirMod); err != nil {
		log.Warn().Err(err).Msgf("Unable to create screen-dumps dir: %s", AppDumpsDir)
	}
	AppBenchmarksDir = filepath.Join(AppConfigDir, "benchmarks")
	if err := data.EnsureFullPath(AppBenchmarksDir, data.DefaultDirMod); err != nil {
		log.Warn().Err(err).Msgf("Unable to create benchmarks dir: %s", AppBenchmarksDir)
	}
	AppSkinsDir = filepath.Join(AppConfigDir, "skins")
	if err := data.EnsureFullPath(AppSkinsDir, data.DefaultDirMod); err != nil {
		log.Warn().Err(err).Msgf("Unable to create skins dir: %s", AppSkinsDir)
	}
	AppContextsDir = filepath.Join(AppConfigDir, "clusters")
	if err := data.EnsureFullPath(AppContextsDir, data.DefaultDirMod); err != nil {
		log.Warn().Err(err).Msgf("Unable to create clusters dir: %s", AppContextsDir)
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

	AppDumpsDir, err = xdg.StateFile(filepath.Join(AppName, "screen-dumps"))
	if err != nil {
		return err
	}

	AppBenchmarksDir, err = xdg.StateFile(filepath.Join(AppName, "benchmarks"))
	if err != nil {
		log.Warn().Err(err).Msgf("No benchmarks dir detected")
	}

	AppConfigFile, err = xdg.ConfigFile(filepath.Join(AppName, data.MainConfigFile))
	if err != nil {
		return err
	}

	dataDir, err := xdg.DataFile(AppName)
	if err != nil {
		return err
	}
	AppSkinsDir = filepath.Join(dataDir, "skins")
	if err := data.EnsureFullPath(AppSkinsDir, data.DefaultDirMod); err != nil {
		log.Warn().Err(err).Msgf("No skins dir detected")
	}
	AppContextsDir = filepath.Join(dataDir, "clusters")
	if err := data.EnsureFullPath(AppContextsDir, data.DefaultDirMod); err != nil {
		log.Warn().Err(err).Msgf("No context dir detected")
	}

	AppHotKeysFile, err = xdg.DataFile(filepath.Join(AppName, "hotkeys.yaml"))
	if err != nil {
		log.Warn().Err(err).Msgf("No hotkeys file detected")
	}

	AppAliasesFile, err = xdg.DataFile(filepath.Join(AppName, "aliases.yaml"))
	if err != nil {
		log.Warn().Err(err).Msgf("No aliases file detected")
	}

	AppPluginsFile, err = xdg.DataFile(filepath.Join(AppName, "plugins.yaml"))
	if err != nil {
		log.Warn().Err(err).Msgf("No plugins file detected")
	}

	AppViewsFile, err = xdg.DataFile(filepath.Join(AppName, "views.yaml"))
	if err != nil {
		log.Warn().Err(err).Msgf("No views file detected")
	}

	return nil
}

func SanitizeFileName(name string) string {
	return invalidPathCharsRX.ReplaceAllString(name, "-")
}

func AppContextDir(context string) string {
	return filepath.Join(AppContextsDir, SanitizeFileName(context))
}

func AppContextConfig(context string) string {
	return filepath.Join(AppContextDir(context), "config.yaml")
}

func DumpsDir(context string) (string, error) {
	dir := filepath.Join(AppDumpsDir, SanitizeFileName(context))

	return dir, data.EnsureDirPath(dir, data.DefaultDirMod)
}

func BenchmarksDir(context string) (string, error) {
	dir := filepath.Join(AppBenchmarksDir, SanitizeFileName(context))

	return dir, data.EnsureDirPath(dir, data.DefaultDirMod)
}

func BenchConfigFile(context string) (string, error) {
	path := filepath.Join(AppContextDir(context), "benchmarks.yaml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", err
	}

	return path, nil
}

func SkinFile(skin string) (string, error) {
	path := filepath.Join(AppSkinsDir, "skin.yaml")
	_, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	return path, nil
}

func hasK9sConfigEnv() bool {
	return os.Getenv(K9sConfigDir) != ""
}

func userTmpDir() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(os.TempDir(), AppName, u.Username)
	if err := data.EnsureFullPath(dir, data.DefaultDirMod); err != nil {
		return "", err
	}

	return dir, nil
}
