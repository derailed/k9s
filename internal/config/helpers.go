package config

import (
	"errors"
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
	if err == nil {
		return usr.Username
	}

	unknownUserIdError := user.UnknownUserIdError(1)
	if errors.As(err, &unknownUserIdError) {
		// error is raised due to unknown user id (which might happen on linux systems enrolled to a domain
		// controller, see https://github.com/derailed/k9s/issues/1895 for more information), but $USER might
		// contain the username
		username := os.Getenv("USER")
		if len(username) != 0 {
			return username
		}
	}

	// always return the error of user.Current, even if the $USER lookup failed -> the root cause is the same
	log.Fatal().Err(err).Msg("Die on retrieving user info")
	return "" // won't reach until here
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
