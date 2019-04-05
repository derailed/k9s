package views

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
)

func toPerc(f float64) string {
	return fmt.Sprintf("%.0f%%", f)
}

func deltas(o, n string) string {
	o, n = strings.TrimSpace(o), strings.TrimSpace(n)
	if o == "" || o == "n/a" {
		return ""
	}

	if i, ok := numerical(o); ok {
		j, _ := numerical(n)
		switch {
		case i < j:
			return plus()
		case i > j:
			return minus()
		default:
			return ""
		}
	}

	if q1, err := resource.ParseQuantity(o); err == nil {
		q2, _ := resource.ParseQuantity(n)
		switch q1.Cmp(q2) {
		case -1:
			return plus()
		case 1:
			return minus()
		default:
			return ""
		}
	}

	if d1, err := time.ParseDuration(o); err == nil {
		d2, _ := time.ParseDuration(n)
		switch {
		case d2-d1 > 0:
			return plus()
		case d2-d1 < 0:
			return minus()
		default:
			return ""
		}
	}

	switch strings.Compare(o, n) {
	case 1, -1:
		return delta()
	default:
		return ""
	}
}

func numerical(s string) (int, bool) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}

	return n, true
}

func delta() string {
	return "𝜟"
}

func plus() string {
	return "⬆"
}

func minus() string {
	return "⬇︎"
}
