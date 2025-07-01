// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package health

// Level tracks health count categories.
type Level int

const (
	// Unknown represents no health level.
	Unknown Level = 1 << iota

	// Corpus tracks total health.
	Corpus

	// S1 tracks series 1.
	S1

	// S2 tracks series 2.
	S2

	// S3 tracks series 3.
	S3
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
type Counts map[Level]int64

// Vital tracks a resource vitals.
type Vital struct {
	Resource         string
	Total, OK, Toast int
}

// Vitals tracks a collection of resource health.
type Vitals []Vital
