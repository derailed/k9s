package views

import (
	"fmt"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
}

func TestCmdStackPushMax(t *testing.T) {
	s := newCmdStack()
	for i := 0; i < 20; i++ {
		s.push(fmt.Sprintf("cmd_%d", i))
	}
	top, ok := s.top()
	assert.True(t, ok)
	assert.Equal(t, "cmd_19", top)
}

func TestCmdStackPop(t *testing.T) {
	type expect struct {
		val string
		ok  bool
	}

	uu := []struct {
		cmds     []string
		popCount int
		e        expect
	}{
		{[]string{}, 2, expect{"", false}},
		{[]string{"a", "b", "c"}, 2, expect{"a", true}},
		{[]string{"a", "b", "c"}, 1, expect{"b", true}},
	}

	for _, u := range uu {
		s := newCmdStack()
		for _, v := range u.cmds {
			s.push(v)
		}
		for i := 0; i < u.popCount; i++ {
			s.pop()
		}
		top, ok := s.pop()
		assert.Equal(t, u.e.ok, ok)
		assert.Equal(t, u.e.val, top)
	}
}

func TestCmdStackEmpty(t *testing.T) {
	uu := []struct {
		cmds     []string
		popCount int
		e        bool
	}{
		{[]string{}, 0, true},
		{[]string{"a", "b", "c"}, 0, false},
		{[]string{"a", "b", "c"}, 3, true},
	}

	for _, u := range uu {
		s := newCmdStack()
		for _, v := range u.cmds {
			s.push(v)
		}
		for i := 0; i < u.popCount; i++ {
			s.pop()
		}
		assert.Equal(t, u.e, s.empty())
	}
}
