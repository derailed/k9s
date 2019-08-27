package config

const (
	// DefaultRefreshRate represents the refresh interval.
	DefaultRefreshRate = 2 // secs

	// DefaultLogLevel represents the default log level.
	DefaultLogLevel = "info"

	// DefaultCommand represents the default command to run.
	DefaultCommand = "po"
)

// Flags represents K9s configuration flags.
type Flags struct {
	RefreshRate *int
	LogLevel    *string
	Headless    *bool
	Command     *string
}

// NewFlags returns new configuration flags.
func NewFlags() *Flags {
	return &Flags{
		RefreshRate: intPtr(DefaultRefreshRate),
		LogLevel:    strPtr(DefaultLogLevel),
		Headless:    boolPtr(false),
		Command:     strPtr(DefaultCommand),
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
