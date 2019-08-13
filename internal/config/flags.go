package config

const (
	// DefaultRefreshRate represents the refresh interval.
	DefaultRefreshRate = 2 // secs
	// DefaultLogLevel represents the default log level.
	DefaultLogLevel = "info"
)

// Flags represents K9s configuration flags.
type Flags struct {
	RefreshRate *int
	LogLevel    *string
	Headless    *bool
}

// NewFlags returns new configuration flags.
func NewFlags() *Flags {
	return &Flags{
		RefreshRate: intPtr(DefaultRefreshRate),
		LogLevel:    strPtr(DefaultLogLevel),
		Headless:    boolPtr(false),
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
