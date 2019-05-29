package views

import "github.com/rs/zerolog/log"

const maxStackSize = 10

type cmdStack struct {
	index int
	stack []string
}

func newCmdStack() *cmdStack {
	return &cmdStack{stack: make([]string, 0, maxStackSize)}
}

func (s *cmdStack) push(cmd string) {
	if len(s.stack) == maxStackSize {
		s.stack = s.stack[1 : len(s.stack)-1]
	}
	s.stack = append(s.stack, cmd)
}

func (s *cmdStack) pop() (string, bool) {
	if s.empty() {
		return "", false
	}
	log.Info().Msgf("Before Pop %v", s.stack)
	top := s.stack[len(s.stack)-1]
	s.stack = s.stack[:len(s.stack)-1]
	log.Info().Msgf("After Pop %v", s.stack)
	return top, true
}

func (s *cmdStack) top() (string, bool) {
	if s.empty() {
		return "", false
	}
	log.Info().Msgf("Top %v -- %s", s.stack, s.stack[len(s.stack)-1])

	return s.stack[len(s.stack)-1], true
}

func (s *cmdStack) empty() bool {
	return len(s.stack) == 0
}

func (s *cmdStack) last() bool {
	return len(s.stack) == 1
}
