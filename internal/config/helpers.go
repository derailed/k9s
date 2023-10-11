package config

import (
	"os"
	"os/user"
	"path/filepath"
	"regexp"

	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
)

const (
	// DefaultDirMod default unix perms for k9s directory.
	DefaultDirMod os.FileMode = 0755
	// DefaultFileMod default unix perms for k9s files.
	DefaultFileMod os.FileMode = 0600
)

var invalidPathCharsRX = regexp.MustCompile(`[:]+`)

// SanitizeFilename sanitizes the dump filename.
func SanitizeFilename(name string) string {
	return invalidPathCharsRX.ReplaceAllString(name, "-")
}

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

// MustK9sUser establishes current user identity or fail.
func MustK9sUser() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal().Err(err).Msg("Die on retrieving user info")
	}
	return usr.Username
}

// EnsureDirPath ensures a directory exist from the given path.
func EnsureDirPath(path string, mod os.FileMode) error {
	return EnsureFullPath(filepath.Dir(path), mod)
}

// EnsureFullPath ensures a directory exist from the given path.
func EnsureFullPath(path string, mod os.FileMode) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err = os.MkdirAll(path, mod); err != nil {
			return err
		}
	}

	return nil
}

// IsBoolSet checks if a bool prt is set.
func IsBoolSet(b *bool) bool {
	return b != nil && *b
}
