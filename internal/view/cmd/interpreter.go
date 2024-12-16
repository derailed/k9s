// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package cmd

import (
	"strings"
)

// Interpreter tracks user prompt input.
type Interpreter struct {
	line string
	cmd  string
	args args
}

// NewInterpreter returns a new instance.
func NewInterpreter(s string) *Interpreter {
	c := Interpreter{
		line: s,
		args: make(args),
	}
	c.grok()

	return &c
}

func (c *Interpreter) grok() {
	ff := strings.Fields(c.line)
	if len(ff) == 0 {
		return
	}
	c.cmd = strings.ToLower(ff[0])
	c.args = newArgs(c, ff[1:])
}

// HasNS returns true if ns is present in prompt.
func (c *Interpreter) HasNS() bool {
	ns, ok := c.args[nsKey]

	return ok && ns != ""
}

// Cmd returns the command.
func (c *Interpreter) Cmd() string {
	return c.cmd
}

// IsBlank returns true if prompt is empty.
func (c *Interpreter) IsBlank() bool {
	return c.line == ""
}

// Amend merges prompts.
func (c *Interpreter) Amend(c1 *Interpreter) {
	c.cmd = c1.cmd
	if c.args == nil {
		c.args = make(args, len(c1.args))
	}
	for k, v := range c1.args {
		if v != "" {
			c.args[k] = v
		}
	}
}

// Reset resets with new command.
func (c *Interpreter) Reset(s string) *Interpreter {
	c.line = s
	c.grok()

	return c
}

// GetLine returns the prompt.
func (c *Interpreter) GetLine() string {
	return strings.TrimSpace(c.line)
}

// IsCowCmd returns true if cow cmd is detected.
func (c *Interpreter) IsCowCmd() bool {
	return c.cmd == cowCmd
}

// IsHelpCmd returns true if help cmd is detected.
func (c *Interpreter) IsHelpCmd() bool {
	_, ok := helpCmd[c.cmd]
	return ok
}

// IsBailCmd returns true if quit cmd is detected.
func (c *Interpreter) IsBailCmd() bool {
	_, ok := bailCmd[c.cmd]
	return ok
}

// IsAliasCmd returns true if alias cmd is detected.
func (c *Interpreter) IsAliasCmd() bool {
	_, ok := aliasCmd[c.cmd]
	return ok
}

// IsXrayCmd returns true if xray cmd is detected.
func (c *Interpreter) IsXrayCmd() bool {
	_, ok := xrayCmd[c.cmd]

	return ok
}

// IsContextCmd returns true if context cmd is detected.
func (c *Interpreter) IsContextCmd() bool {
	_, ok := contextCmd[c.cmd]

	return ok
}

// IsNamespaceCmd returns true if ns cmd is detected.
func (c *Interpreter) IsNamespaceCmd() bool {
	_, ok := namespaceCmd[c.cmd]

	return ok
}

// IsDirCmd returns true if dir cmd is detected.
func (c *Interpreter) IsDirCmd() bool {
	_, ok := dirCmd[c.cmd]
	return ok
}

// IsRBACCmd returns true if rbac cmd is detected.
func (c *Interpreter) IsRBACCmd() bool {
	return c.cmd == canCmd
}

// ContextArg returns context cmd arg.
func (c *Interpreter) ContextArg() (string, bool) {
	if !c.IsContextCmd() {
		return "", false
	}

	return c.args[contextKey], true
}

// ResetContextArg deletes context arg.
func (c *Interpreter) ResetContextArg() {
	delete(c.args, contextFlag)
}

// DirArg returns the directory is present.
func (c *Interpreter) DirArg() (string, bool) {
	if !c.IsDirCmd() {
		return "", false
	}
	d, ok := c.args[topicKey]

	return d, ok && d != ""
}

// CowArg returns the cow message.
func (c *Interpreter) CowArg() (string, bool) {
	if !c.IsCowCmd() {
		return "", false
	}
	m, ok := c.args[nsKey]

	return m, ok && m != ""
}

// RBACArgs returns the subject and topic is any.
func (c *Interpreter) RBACArgs() (string, string, bool) {
	if !c.IsRBACCmd() {
		return "", "", false
	}
	tt := rbacRX.FindStringSubmatch(c.line)
	if len(tt) < 3 {
		return "", "", false
	}

	return tt[1], tt[2], true
}

// XRayArgs return the gvr and ns if any.
func (c *Interpreter) XrayArgs() (string, string, bool) {
	if !c.IsXrayCmd() {
		return "", "", false
	}
	gvr, ok1 := c.args[topicKey]
	if !ok1 {
		return "", "", false
	}

	ns, ok2 := c.args[nsKey]
	switch {
	case ok1 && ok2:
		return gvr, ns, true
	case ok1 && !ok2:
		return gvr, "", true
	default:
		return "", "", false
	}
}

// FilterArg returns the current filter if any.
func (c *Interpreter) FilterArg() (string, bool) {
	f, ok := c.args[filterKey]

	return f, ok && f != ""
}

// FuzzyArg returns the fuzzy filter if any.
func (c *Interpreter) FuzzyArg() (string, bool) {
	f, ok := c.args[fuzzyKey]

	return f, ok && f != ""
}

// NSArg returns the current ns if any.
func (c *Interpreter) NSArg() (string, bool) {
	ns, ok := c.args[nsKey]

	return ns, ok && ns != ""
}

// HasContext returns the current context if any.
func (c *Interpreter) HasContext() (string, bool) {
	ctx, ok := c.args[contextKey]

	return ctx, ok && ctx != ""
}

// LabelsArg return the labels map if any.
func (c *Interpreter) LabelsArg() (map[string]string, bool) {
	ll, ok := c.args[labelKey]

	return ToLabels(ll), ok
}
