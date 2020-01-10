package client

import (
	"os/user"
	"path"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
)

var toFileName = regexp.MustCompile(`[^(\w/\.)]`)

// IsClusterWide returns true if ns designates cluster scope, false otherwise.
func IsClusterWide(ns string) bool {
	return ns == NamespaceAll || ns == AllNamespaces || ns == ClusterScope
}

// CleanseNamespace ensures all ns maps to blank.
func CleanseNamespace(ns string) string {
	if IsAllNamespace(ns) {
		return AllNamespaces
	}

	return ns
}

// IsAllNamespace returns true if ns == all.
func IsAllNamespace(ns string) bool {
	return ns == NamespaceAll
}

// IsAllNamespaces returns true if all namespaces, false otherwise.
func IsAllNamespaces(ns string) bool {
	return ns == NamespaceAll || ns == AllNamespaces
}

// IsNamespaced returns true if a specific ns is given.
func IsNamespaced(ns string) bool {
	return !IsClusterScoped(ns)
}

// IsClusterScoped returns true if resource is not namespaced.
func IsClusterScoped(ns string) bool {
	return ns == ClusterScope
}

// Namespaced converts a resource path to namespace and resource name.
func Namespaced(p string) (string, string) {
	ns, n := path.Split(p)

	return strings.Trim(ns, "/"), n
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
