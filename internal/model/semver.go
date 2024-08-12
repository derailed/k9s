// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model

import (
	"fmt"
	"regexp"
	"strconv"
)

var versionRX = regexp.MustCompile(`\Av(\d+)\.(\d+)\.(\d+)\z`)

// SemVer represents a semantic version.
type SemVer struct {
	Major, Minor, Patch int
}

// NewSemVer returns a new semantic version.
func NewSemVer(version string) *SemVer {
	var v SemVer
	v.Major, v.Minor, v.Patch = v.parse(NormalizeVersion(version))

	return &v
}

// String returns version as a string.
func (v *SemVer) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (*SemVer) parse(version string) (major, minor, patch int) {
	mm := versionRX.FindStringSubmatch(version)
	if len(mm) < 4 {
		return
	}
	major, _ = strconv.Atoi(mm[1])
	minor, _ = strconv.Atoi(mm[2])
	patch, _ = strconv.Atoi(mm[3])

	return
}

// NormalizeVersion ensures the version starts with a v.
func NormalizeVersion(version string) string {
	if version == "" {
		return version
	}
	if version[0] == 'v' {
		return version
	}
	return "v" + version
}

// IsCurrent asserts if at latest release.
func (v *SemVer) IsCurrent(latest *SemVer) bool {
	return v.Major >= latest.Major && v.Minor >= latest.Minor && v.Patch >= latest.Patch
}
