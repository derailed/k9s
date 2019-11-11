package filters

import (
	"testing"
)

func TestFilter(t *testing.T) {
	uu := map[string]struct {
		filter        string
		parsedFilter  string
		regex         bool
		fuzzy         bool
		labelSelector bool
	}{
		"basic filter": {
			filter:       "my_app",
			parsedFilter: "my_app",
			regex:        false,
		},

		"labelSelector/cool": {
			filter:        "l: app=my_app,env=prod",
			parsedFilter:  "app=my_app,env=prod",
			labelSelector: true,
		},
		"labelSelector/noSpace": {
			filter:        "l:app=my_app,env=prod",
			parsedFilter:  "app=my_app,env=prod",
			labelSelector: true,
		},
		"labelSelector/legacyLabelSelector": {
			filter:        "-l app=my_app,env=prod",
			parsedFilter:  "app=my_app,env=prod",
			labelSelector: true,
		},

		"fuzzy/cool": {
			filter:       "f:my_app",
			parsedFilter: "my_app",
			fuzzy:        true,
		},
		"fuzzy/noSpace": {
			filter:       "f: my_app",
			parsedFilter: "my_app",
			fuzzy:        true,
		},
		"fuzzy/legacyLabelSelector": {
			filter:       "-fmy_app",
			parsedFilter: "my_app",
			fuzzy:        true,
		},
	}

	for name, u := range uu {
		t.Run(name, func(t *testing.T) {
			parsedFilter, isLabelSelector := LabelSelector(u.filter)
			parsedFilter, isFuzzyFilter := FuzzyFilter(parsedFilter)

			if isFuzzyFilter != u.fuzzy {
				t.Fatalf("Wrong fuzzy modifier. Expected '%v' got '%v'", u.fuzzy, isFuzzyFilter)
			}

			if isLabelSelector != u.labelSelector {
				t.Fatalf("Wrong label selector modifier. Expected '%v' got '%v'", u.fuzzy, isFuzzyFilter)
			}

			if parsedFilter != u.parsedFilter {
				t.Fatalf("Expected parsed filter to be '%s' got '%s'", u.parsedFilter, parsedFilter)
			}
		})
	}
}
