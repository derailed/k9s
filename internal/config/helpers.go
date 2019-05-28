package config

import (
	"os"
	"os/user"
	"path/filepath"

	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
)

const (
	// DefaultDirMod default unix perms for k9s directory.
	DefaultDirMod os.FileMode = 0755
	// DefaultFileMod default unix perms for k9s files.
	DefaultFileMod os.FileMode = 0644
)

// InList check if string is in a collection of strings.
func InList(ll []string, n string) bool {
	for _, l := range ll {
		if l == n {
			return true
		}
	}
	return false
}

// InNSList check if ns is in an ns collection.
func InNSList(nn []interface{}, ns string) bool {
	ss := make([]string, len(nn))
	for i, n := range nn {
		if nsp, ok := n.(v1.Namespace); ok {
			ss[i] = nsp.Name
		}
	}
	return InList(ss, ns)
}

func mustK9sHome() string {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	return usr.HomeDir
}

// MustK9sUser establishes current user identity or fail.
func MustK9sUser() string {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	return usr.Username
}

// EnsurePath ensures a directory exist from the given path.
func EnsurePath(path string, mod os.FileMode) {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, mod); err != nil {
			log.Error().Msgf("Unable to create K9s home config dir: %v", err)
			panic(err)
		}
	}
}
