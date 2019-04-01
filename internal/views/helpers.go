package views

import (
	"fmt"
	"strconv"
	"strings"
)

func toPerc(f float64) string {
	return fmt.Sprintf("%.0f%%", f)
}

func deltas(c, n string) string {
	c, n = strings.TrimSpace(c), strings.TrimSpace(n)
	if c == "n/a" {
		return ""
	}

	if i, ok := numerical(c); ok {
		if j, ok := numerical(n); ok {
			switch {
			case i < j:
				return plus()
			case i > j:
				return minus()
			default:
				return ""
			}
		}
		return ""
	}

	if isAlpha(c) {
		if strings.Contains(c, "(") {
			return ""
		}
		switch strings.Compare(c, n) {
		case -1:
			return plus()
		case 1:
			return minus()
		default:
			return ""
		}
	}

	if len(c) == 0 {
		return ""
	}

	switch strings.Compare(c, n) {
	case 1, -1:
		return delta()
	default:
		return ""
	}
}

func isAlpha(s string) bool {
	if len(s) == 0 {
		return false
	}

	if _, err := strconv.Atoi(string(s[0])); err != nil {
		return false
	}
	return true
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
