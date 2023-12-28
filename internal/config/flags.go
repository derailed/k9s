// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

const (
	// DefaultRefreshRate represents the refresh interval.
	DefaultRefreshRate = 2 // secs

	// DefaultLogLevel represents the default log level.
	DefaultLogLevel = "info"

	// DefaultCommand represents the default command to run.
	DefaultCommand = ""
)

// Flags represents K9s configuration flags.
type Flags struct {
	RefreshRate   *int
	LogLevel      *string
	LogFile       *string
	Headless      *bool
	Logoless      *bool
	Command       *string
	AllNamespaces *bool
	ReadOnly      *bool
	Write         *bool
	Crumbsless    *bool
	ScreenDumpDir *string
}

// NewFlags returns new configuration flags.
func NewFlags() *Flags {
	return &Flags{
		RefreshRate:   intPtr(DefaultRefreshRate),
		LogLevel:      strPtr(DefaultLogLevel),
		LogFile:       strPtr(AppLogFile),
		Headless:      boolPtr(false),
		Logoless:      boolPtr(false),
		Command:       strPtr(DefaultCommand),
		AllNamespaces: boolPtr(false),
		ReadOnly:      boolPtr(false),
		Write:         boolPtr(false),
		Crumbsless:    boolPtr(false),
		ScreenDumpDir: strPtr(AppDumpsDir),
	}
}

func boolPtr(b bool) *bool {
	return &b
}

func intPtr(i int) *int {
	return &i
}

func strPtr(s string) *string {
	return &s
}
