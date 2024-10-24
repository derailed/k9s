// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package internal

import (
	"regexp"
	"strings"

	"github.com/derailed/k9s/internal/view/cmd"
)

var (
	inverseRx = regexp.MustCompile(`\A\!`)
	fuzzyRx   = regexp.MustCompile(`\A-f\s?([\w-]+)\b`)
	labelRx   = regexp.MustCompile(`\A\-l`)
)

// Helpers...

// IsInverseSelector checks if inverse char has been provided.
func IsInverseSelector(s string) bool {
	if s == "" {
		return false
	}
	return inverseRx.MatchString(s)
}

// IsLabelSelector checks if query is a label query.
func IsLabelSelector(s string) bool {
	if labelRx.MatchString(s) {
		return true
	}

	return !strings.Contains(s, " ") && cmd.ToLabels(s) != nil
}

// IsFuzzySelector checks if query is fuzzy.
func IsFuzzySelector(s string) (string, bool) {
	mm := fuzzyRx.FindStringSubmatch(s)
	if len(mm) != 2 {
		return "", false
	}

	return mm[1], true
}
