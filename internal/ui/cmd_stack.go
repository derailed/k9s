package ui

const maxStackSize = 10

// CmdStack tracks users command breadcrumbs.
type CmdStack struct {
	index int
	stack []string
}

// NewCmdStack returns a new cmd stack.
func NewCmdStack() *CmdStack {
	return &CmdStack{stack: make([]string, 0, maxStackSize)}
}

// Items returns current stack content.
func (s *CmdStack) Items() []string {
	return s.stack
}

// Push adds a new item,
func (s *CmdStack) Push(cmd string) {
	if len(s.stack) == maxStackSize {
		s.stack = s.stack[1 : len(s.stack)-1]
	}
	s.stack = append(s.stack, cmd)
}

// Pop delete an item.
func (s *CmdStack) Pop() (string, bool) {
	if s.Empty() {
		return "", false
	}

	top := s.stack[len(s.stack)-1]
	s.stack = s.stack[:len(s.stack)-1]

	return top, true
}

// Top return top element.
func (s *CmdStack) Top() (string, bool) {
	if s.Empty() {
		return "", false
	}

	return s.stack[len(s.stack)-1], true
}

// Empty check if stack is empty.
func (s *CmdStack) Empty() bool {
	return len(s.stack) == 0
}

// Last returns the last command.
func (s *CmdStack) Last() bool {
	return len(s.stack) == 1
}
