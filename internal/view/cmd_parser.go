// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"regexp"
	"strings"
)

const (
	cowCmd    = "cow"
	canCmd    = "can"
	filterCmd = "-f"
	nsCmd     = "-n"
	labelCmd  = "-l"
	slashCmd  = "/"
)

var (
	rbacRX = regexp.MustCompile(`^can\s+([u|g|s]):\s*([\w-:]+)\s*$`)

	contextCmd = map[string]struct{}{
		"ctx":      {},
		"context":  {},
		"contexts": {},
	}
	dirCmd = map[string]struct{}{
		"dir": {},
		"d":   {},
		"ls":  {},
	}
	bailCmd = map[string]struct{}{
		"q":    {},
		"q!":   {},
		"qa":   {},
		"Q":    {},
		"quit": {},
		"exit": {},
	}
	helpCmd = map[string]struct{}{
		"?":    {},
		"h":    {},
		"help": {},
	}
	aliasCmd = map[string]struct{}{
		"a":     {},
		"alias": {},
	}
	xrayCmd = map[string]struct{}{
		"x":    {},
		"xr":   {},
		"xray": {},
	}
)

type cmdParser struct {
	line string
	cmd  string
	args []string
}

func newCmdParser(s string) *cmdParser {
	c := cmdParser{line: strings.ToLower(s)}
	c.grok()

	return &c
}

func (c *cmdParser) isEmpty() bool {
	return c.line == ""
}

func (c *cmdParser) reset(s string) *cmdParser {
	c.line = strings.ToLower(s)
	c.grok()

	return c
}

func (c *cmdParser) getLine() string {
	return strings.TrimSpace(c.line)
}

func (c *cmdParser) grok() {
	ff := strings.Fields(c.line)
	if len(ff) == 0 {
		return
	}
	var args []string
	c.cmd, args = ff[0], ff[1:]
	if c.isEmpty() {
		return
	}
	for _, a := range args {
		switch {
		case strings.Index(a, filterCmd) == 0:
			if a == filterCmd {
				c.args = append(c.args, a)
			} else {
				c.args = append(c.args, filterCmd, a[2:])
			}
		case strings.Index(a, labelCmd) == 0:
			if a == labelCmd {
				c.args = append(c.args, a)
			} else {
				c.args = append(c.args, labelCmd, a[2:])
			}
		default:
			c.args = append(c.args, a)
		}
	}
}

func (c *cmdParser) getArg() (string, bool) {
	if c.isEmpty() || len(c.args) == 0 {
		return "", false
	}

	return c.args[0], true
}

func (c *cmdParser) filterCmd() (string, bool) {
	if c.isEmpty() || len(c.args) == 0 {
		return "", false
	}
	if c.args[0] == filterCmd && len(c.args) > 1 {
		return c.args[1], true
	}
	if strings.HasPrefix(c.args[0], filterCmd) {
		return strings.TrimPrefix(c.args[0], filterCmd), true
	}

	return "", false
}

func (c *cmdParser) hasSelector() bool {
	return c.args[0] == filterCmd || c.args[0] == labelCmd
}

func (c *cmdParser) nsCmd() (string, bool) {
	var ns string
	if c.isEmpty() || len(c.args) == 0 {
		return ns, false
	}

	if c.hasSelector() && len(c.args) == 2 {
		return ns, false
	}

	if c.args[0] == nsCmd && len(c.args) > 1 {
		ns = c.args[1]
	} else {
		ns = c.args[0]
	}

	return ns, len(ns) > 0
}

func (c *cmdParser) labelsCmd() (map[string]string, bool) {
	if c.isEmpty() || len(c.args) == 0 {
		return nil, false
	}
	if c.args[0] == labelCmd && len(c.args) > 1 {
		return toLabels(c.args[1]), true
	}
	if strings.HasPrefix(c.args[0], labelCmd) {
		return toLabels(strings.TrimPrefix(c.args[0], labelCmd)), true
	}

	return nil, false
}

func toLabels(s string) map[string]string {
	ll := strings.Split(s, ",")
	if len(ll) == 0 {
		return nil
	}
	lbls := make(map[string]string, len(ll))
	for _, l := range ll {
		kv := strings.Split(l, "=")
		if len(kv) == 0 || len(kv) < 2 {
			continue
		}
		lbls[kv[0]] = kv[1]
	}

	return lbls
}

func (c *cmdParser) isCowCmd() bool {
	return c.cmd == cowCmd
}

func (c *cmdParser) isHelpCmd() bool {
	_, ok := helpCmd[c.cmd]
	return ok
}

func (c *cmdParser) isBailoutCmd() bool {
	_, ok := bailCmd[c.cmd]
	return ok
}

func (c *cmdParser) isAliasCmd() bool {
	_, ok := aliasCmd[c.cmd]
	return ok
}

func (c *cmdParser) isXrayCmd() bool {
	_, ok := xrayCmd[c.cmd]
	return ok
}

func (c *cmdParser) isContextCmd() bool {
	_, ok := contextCmd[c.cmd]
	return ok
}

func (c *cmdParser) isDirCmd() bool {
	_, ok := dirCmd[c.cmd]
	return ok
}

func (c *cmdParser) isRbacCmd() bool {
	return c.cmd == canCmd
}

func (c *cmdParser) parseRbac() (string, string, bool) {
	if !c.isRbacCmd() {
		return "", "", false
	}
	tt := rbacRX.FindStringSubmatch(c.line)
	if len(tt) < 3 {
		return "", "", false
	}

	return tt[1], tt[2], true
}

func (c *cmdParser) parseXray() (string, string, bool) {
	if !c.isXrayCmd() {
		return "", "", false
	}
	switch len(c.args) {
	case 0:
		return "", "", false
	case 1:
		return c.args[0], "", true
	default:
		return c.args[0], c.args[1], true
	}
}
