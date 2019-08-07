package ui

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	res "github.com/derailed/k9s/internal/resource"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	// DeltaSign signals a diff.
	DeltaSign = "Δ"
	// PlusSign signals inc.
	PlusSign = "↑"
	// MinusSign signal dec.
	MinusSign = "↓"
)

var percent = regexp.MustCompile(`\A(\d+)\%\z`)

// Deltas signals diffs between 2 strings.
func Deltas(o, n string) string {
	o, n = strings.TrimSpace(o), strings.TrimSpace(n)
	if o == "" || o == res.NAValue {
		return ""
	}

	if i, ok := numerical(o); ok {
		j, _ := numerical(n)
		switch {
		case i < j:
			return PlusSign
		case i > j:
			return MinusSign
		default:
			return ""
		}
	}

	if i, ok := percentage(o); ok {
		j, _ := percentage(n)
		switch {
		case i < j:
			return PlusSign
		case i > j:
			return MinusSign
		default:
			return ""
		}
	}

	if q1, err := resource.ParseQuantity(o); err == nil {
		q2, _ := resource.ParseQuantity(n)
		switch q1.Cmp(q2) {
		case -1:
			return PlusSign
		case 1:
			return MinusSign
		default:
			return ""
		}
	}

	if d1, err := time.ParseDuration(o); err == nil {
		d2, _ := time.ParseDuration(n)
		switch {
		case d2-d1 > 0:
			return PlusSign
		case d2-d1 < 0:
			return MinusSign
		default:
			return ""
		}
	}

	switch strings.Compare(o, n) {
	case 1, -1:
		return DeltaSign
	default:
		return ""
	}
}

func percentage(s string) (int, bool) {
	if res := percent.FindStringSubmatch(s); len(res) == 2 {
		n, _ := strconv.Atoi(res[1])
		return n, true
	}

	return 0, false
}

func numerical(s string) (int, bool) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}

	return n, true
}
