// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"log/slog"
	"os"
	"os/user"
	"path/filepath"

	"github.com/derailed/k9s/internal/slogs"
)

const (
	envPFAddress          = "K9S_DEFAULT_PF_ADDRESS"
	defaultPortFwdAddress = "localhost"
)

// IsBoolSet checks if a bool ptr is set.
func IsBoolSet(b *bool) bool {
	return b != nil && *b
}

func isStringSet(s *string) bool {
	return s != nil && *s != ""
}

func isYamlFile(file string) bool {
	ext := filepath.Ext(file)

	return ext == ".yml" || ext == ".yaml"
}

// isEnvSet checks if env var is set.
func isEnvSet(env string) bool {
	return os.Getenv(env) != ""
}

// UserTmpDir returns the temp dir with the current user name.
func UserTmpDir() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(os.TempDir(), u.Username, AppName)

	return dir, nil
}

// MustK9sUser establishes current user identity or fail.
func MustK9sUser() string {
	usr, err := user.Current()
	if err != nil {
		envUsr := os.Getenv("USER")
		if envUsr != "" {
			return envUsr
		}
		envUsr = os.Getenv("LOGNAME")
		if envUsr != "" {
			return envUsr
		}
		slog.Error("Die on retrieving user info", slogs.Error, err)
		os.Exit(1)
	}
	return usr.Username
}

func defaultPFAddress() string {
	if a := os.Getenv(envPFAddress); a != "" {
		return a
	}

	return defaultPortFwdAddress
}
