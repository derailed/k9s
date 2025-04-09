// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package client

import (
	"log/slog"
	"os"
	"os/user"
	"path"
	"regexp"
	"strings"

	"github.com/derailed/k9s/internal/slogs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var toFileName = regexp.MustCompile(`[^(\w/\.)]`)

// IsClusterWide returns true if ns designates cluster scope, false otherwise.
func IsClusterWide(ns string) bool {
	return ns == NamespaceAll || ns == BlankNamespace || ns == ClusterScope
}

// CleanseNamespace ensures all ns maps to blank.
func CleanseNamespace(ns string) string {
	if IsAllNamespace(ns) {
		return BlankNamespace
	}

	return ns
}

// IsAllNamespace returns true if ns == all.
func IsAllNamespace(ns string) bool {
	return ns == NamespaceAll
}

// IsAllNamespaces returns true if all namespaces, false otherwise.
func IsAllNamespaces(ns string) bool {
	return ns == NamespaceAll || ns == BlankNamespace
}

// IsNamespaced returns true if a specific ns is given.
func IsNamespaced(ns string) bool {
	return !IsAllNamespaces(ns) && !IsClusterScoped(ns)
}

// IsClusterScoped returns true if resource is not namespaced.
func IsClusterScoped(ns string) bool {
	return ns == ClusterScope
}

// Namespaced converts a resource path to namespace and resource name.
func Namespaced(p string) (ns, name string) {
	ns, name = path.Split(p)

	return strings.Trim(ns, "/"), name
}

// CoFQN returns a fully qualified container name.
func CoFQN(m *metav1.ObjectMeta, co string) string {
	return MetaFQN(m) + ":" + co
}

// FQN returns a fully qualified resource name.
func FQN(ns, n string) string {
	if ns == "" {
		return n
	}
	return ns + "/" + n
}

// MetaFQN returns a fully qualified resource name.
func MetaFQN(m *metav1.ObjectMeta) string {
	if m.Namespace == "" {
		return FQN(ClusterScope, m.Name)
	}

	return FQN(m.Namespace, m.Name)
}

func mustHomeDir() string {
	usr, err := user.Current()
	if err != nil {
		slog.Error("Die getting user home directory", slogs.Error, err)
		os.Exit(1)
	}
	return usr.HomeDir
}

func toHostDir(host string) string {
	h := strings.Replace(
		strings.Replace(host, "https://", "", 1),
		"http://", "", 1,
	)
	return toFileName.ReplaceAllString(h, "_")
}
