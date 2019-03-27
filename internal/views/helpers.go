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
	if c == "n/a" {
		return n
	}

	if i, ok := numerical(c); ok {
		if j, ok := numerical(n); ok {
			switch {
			case i < j:
				return plus(n)
			case i > j:
				return minus(n)
			default:
				return n
			}
		}
		return n
	}

	if isAlpha(c) {
		if strings.Contains(c, "(") {
			return n
		}
		switch strings.Compare(c, n) {
		case -1:
			return plus(n)
		case 1:
			return minus(n)
		default:
			return n
		}
	}

	if len(c) == 0 {
		return n
	}

	switch strings.Compare(c, n) {
	case 1, -1:
		return delta(n)
	default:
		return n
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

func delta(s string) string {
	return suffix(s, "𝜟")
}

func plus(s string) string {
	return suffix(s, "+")
}

func minus(s string) string {
	return suffix(s, "-")
}

func suffix(s, su string) string {
	return s + "(" + su + ")"
}
