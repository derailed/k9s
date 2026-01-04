// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

const (
	// DefaultRefreshRate represents the refresh interval.
	DefaultRefreshRate float32 = 2.0 // secs

	// DefaultLogLevel represents the default log level.
	DefaultLogLevel = "info"

	// DefaultCommand represents the default command to run.
	DefaultCommand = ""
)

// Flags represents K9s configuration flags.
type Flags struct {
	RefreshRate   *float32
	LogLevel      *string
	LogFile       *string
	Headless      *bool
	Logoless      *bool
	Command       *string
	AllNamespaces *bool
	ReadOnly      *bool
	Write         *bool
	Crumbsless    *bool
	Splashless    *bool
	Invert        *bool
	ScreenDumpDir *string
}

// NewFlags returns new configuration flags.
func NewFlags() *Flags {
	return &Flags{
		RefreshRate:   float32Ptr(DefaultRefreshRate),
		LogLevel:      strPtr(DefaultLogLevel),
		LogFile:       strPtr(AppLogFile),
		Headless:      boolPtr(false),
		Logoless:      boolPtr(false),
		Command:       strPtr(DefaultCommand),
		AllNamespaces: boolPtr(false),
		ReadOnly:      boolPtr(false),
		Write:         boolPtr(false),
		Crumbsless:    boolPtr(false),
		Splashless:    boolPtr(false),
		Invert:        boolPtr(false),
		ScreenDumpDir: strPtr(AppDumpsDir),
	}
}

func boolPtr(b bool) *bool {
	return &b
}

func float32Ptr(f float32) *float32 {
	return &f
}

func strPtr(s string) *string {
	return &s
}
