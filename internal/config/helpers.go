// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"os"
	"os/user"
	"path/filepath"

	"github.com/derailed/k9s/internal/config/data"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
)

// IsBoolSet checks if a bool ptr is set.
func IsBoolSet(b *bool) bool {
	return b != nil && *b
}

func isStringSet(s *string) bool {
	return s != nil && len(*s) > 0
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

// InNSList check if ns is in an ns collection.
func InNSList(nn []interface{}, ns string) bool {
	ss := make([]string, len(nn))
	for i, n := range nn {
		if nsp, ok := n.(v1.Namespace); ok {
			ss[i] = nsp.Name
		}
	}
	return data.InList(ss, ns)
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
		log.Fatal().Err(err).Msg("Die on retrieving user info")
	}
	return usr.Username
}
