package k8s

import (
	"math"
	"path"
	"strings"
)

const megaByte = 1024 * 1024

// ToMB converts bytes to megabytes.
func ToMB(v int64) float64 {
	return float64(v) / megaByte
}

func toPerc(v1, v2 float64) float64 {
	if v2 == 0 {
		return 0
	}
	return math.Round((v1 / v2) * 100)
}

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
