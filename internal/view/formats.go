// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

// Shared sparkline-legend format strings used by both pulse.go and charts.go.
const (
	cpuFmt = " %s [%s::b]%s[white::-]([%s::]%sm[white::]/[%s::]%sm[-::])"
	memFmt = " %s [%s::b]%s[white::-]([%s::]%sMi[white::]/[%s::]%sMi[-::])"
)
