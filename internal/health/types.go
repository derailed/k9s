package health

// Level tracks health count categories.
type Level int

const (
	// Unknown represents no health level.
	Unknown Level = 1 << iota

	// Corpus tracks total health.
	Corpus

	// OK tracks healhy.
	OK

	// Warn tracks health warnings.
	Warn

	// Toast tracks unhealties.
	Toast
)

// Message represents a health message.
type Message struct {
	Level   Level
	Message string
	GVR     string
	FQN     string
}

// Messages tracks a collection of messages.
type Messages []Message

// Counts tracks health counts by category.
type Counts map[Level]int

// Vital tracks a resource vitals.
type Vital struct {
	Resource         string
	Total, OK, Toast int
}

// Vitals tracks a collection of resource health.
type Vitals []Vital
