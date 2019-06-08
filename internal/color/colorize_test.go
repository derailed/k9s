package color

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColorize(t *testing.T) {
	uu := map[string]struct {
		s string
		c Paint
		e string
	}{
		"white":   {"blee", White, "\x1b[37mblee\x1b[0m"},
		"black":   {"blee", Black, "\x1b[30mblee\x1b[0m"},
		"default": {"blee", 0, "\x1b[37mblee\x1b[0m"},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, Colorize(u.s, u.c))
		})
	}
}
