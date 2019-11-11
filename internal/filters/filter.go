package filters

import "strings"

const (
	labelPrefixLegacy = "-l"
	labelPrefix       = "l:"

	fuzzyPrefixLegacy = "-f"
	fuzzyPrefix       = "f:"
)

type Filter struct {
	Query           string
	IsLabelSelector bool
	IsFuzzy         bool
	Sticky          bool
}

// LabelSelector extract the query without the label selector prefix.
// If the query is not sticky, returns the same query and false.
func LabelSelector(filter string) (string, bool) {
	// maintain compatibility with the current `-l` label selector prefix.
	filter, ok := HasPrefixAndTrim(filter, labelPrefixLegacy)
	if ok {
		return filter, ok
	}

	return HasPrefixAndTrim(filter, labelPrefix)
}

// FuzzyFilter extract the query without the fuzzy filter prefix.
// If the query is not sticky, returns the same query and false.
func FuzzyFilter(filter string) (string, bool) {
	// maintain compatibility with the current `-f` fuzzy filter prefix.
	filter, ok := HasPrefixAndTrim(filter, fuzzyPrefix)
	if ok {
		return filter, ok
	}

	return HasPrefixAndTrim(filter, fuzzyPrefixLegacy)
}

func HasPrefixAndTrim(filter, prefix string) (string, bool) {
	if strings.HasPrefix(filter, prefix) {
		return strings.TrimSpace(strings.TrimPrefix(filter, prefix)), true
	}

	return filter, false
}
