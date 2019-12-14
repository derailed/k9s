package client

import (
	"os/user"
	"path"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
)

var toFileName = regexp.MustCompile(`[^(\w/\.)]`)

// Namespaced converts a resource path to namespace and resource name.
func Namespaced(n string) (string, string) {
	ns, po := path.Split(n)

	return strings.Trim(ns, "/"), po
}

// FQN returns a fully qualified resource name.
func FQN(ns, n string) string {
	if ns == "" {
		return n
	}
	return ns + "/" + n
}

func mustHomeDir() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal().Err(err).Msg("Die getting user home directory")
	}
	return usr.HomeDir
}

func toHostDir(host string) string {
	h := strings.Replace(strings.Replace(host, "https://", "", 1), "http://", "", 1)
	return toFileName.ReplaceAllString(h, "_")
}
