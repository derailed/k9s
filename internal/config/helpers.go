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

func isEnvSet(env string) bool {
	return os.Getenv(env) != ""
}

func userTmpDir() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(os.TempDir(), AppName, u.Username)

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
		log.Fatal().Err(err).Msg("Die on retrieving user info")
	}
	return usr.Username
}

// IsBoolSet checks if a bool prt is set.
func IsBoolSet(b *bool) bool {
	return b != nil && *b
}
