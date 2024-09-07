// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package cmd

import "regexp"

const (
	cowCmd      = "cow"
	canCmd      = "can"
	nsFlag      = "-n"
	filterFlag  = "/"
	labelFlag   = "="
	fuzzyFlag   = "-f"
	contextFlag = "@"
)

var (
	rbacRX = regexp.MustCompile(`^can\s+([u|g|s]):\s*([\w-:]+)\s*$`)

	contextCmd = map[string]struct{}{
		"ctx":      {},
		"context":  {},
		"contexts": {},
	}
	namespaceCmd = map[string]struct{}{
		"ns":         {},
		"namespace":  {},
		"namespaces": {},
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
