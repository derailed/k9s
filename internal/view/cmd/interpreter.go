// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package cmd

import (
	"strings"
)

type Interpreter struct {
	line string
	cmd  string
	args args
}

func NewInterpreter(s string) *Interpreter {
	c := Interpreter{
		line: strings.ToLower(s),
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
	c.cmd = ff[0]
	c.args = newArgs(c, ff[1:])
}

func (c *Interpreter) HasNS() bool {
	ns, ok := c.args[nsKey]

	return ok && ns != ""
}

func (c *Interpreter) Cmd() string {
	return c.cmd
}

func (c *Interpreter) IsBlank() bool {
	return c.line == ""
}

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

func (c *Interpreter) Reset(s string) *Interpreter {
	c.line = strings.ToLower(s)
	c.grok()

	return c
}

func (c *Interpreter) GetLine() string {
	return strings.TrimSpace(c.line)
}

func (c *Interpreter) IsCowCmd() bool {
	return c.cmd == cowCmd
}

func (c *Interpreter) IsHelpCmd() bool {
	_, ok := helpCmd[c.cmd]
	return ok
}

func (c *Interpreter) IsBailCmd() bool {
	_, ok := bailCmd[c.cmd]
	return ok
}

func (c *Interpreter) IsAliasCmd() bool {
	_, ok := aliasCmd[c.cmd]
	return ok
}

func (c *Interpreter) IsXrayCmd() bool {
	_, ok := xrayCmd[c.cmd]

	return ok
}

func (c *Interpreter) IsContextCmd() bool {
	_, ok := contextCmd[c.cmd]

	return ok
}

func (c *Interpreter) IsDirCmd() bool {
	_, ok := dirCmd[c.cmd]
	return ok
}

func (c *Interpreter) IsRBACCmd() bool {
	return c.cmd == canCmd
}

func (c *Interpreter) ContextArg() (string, bool) {
	if !c.IsContextCmd() {
		return "", false
	}

	return c.args[contextKey], true
}

func (c *Interpreter) ResetContextArg() {
	delete(c.args, contextFlag)
}

func (c *Interpreter) DirArg() (string, bool) {
	if !c.IsDirCmd() || c.args[topicKey] == "" {
		return "", false
	}

	return c.args[topicKey], true
}

func (c *Interpreter) CowArg() (string, bool) {
	if !c.IsCowCmd() || c.args[nsKey] == "" {
		return "", false
	}

	return c.args[nsKey], true
}

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

func (c *Interpreter) FilterArg() (string, bool) {
	f, ok := c.args[filterKey]

	return f, ok
}

func (c *Interpreter) NSArg() (string, bool) {
	ns, ok := c.args[nsKey]

	return ns, ok
}

func (c *Interpreter) HasContext() (string, bool) {
	ctx, ok := c.args[contextKey]
	if !ok || ctx == "" {
		return "", false
	}

	return ctx, ok
}

func (c *Interpreter) LabelsArg() (map[string]string, bool) {
	ll, ok := c.args[labelKey]
	if !ok {
		return nil, false
	}

	return toLabels(ll), true
}
