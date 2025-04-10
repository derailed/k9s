// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package cmd

import (
	"regexp"

	"k8s.io/apimachinery/pkg/util/sets"
)

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
	rbacRX = regexp.MustCompile(`^can\s+([ugs]):\s*([\w-:]+)\s*$`)

	contextCmd = sets.New(
		"ctx",
		"context",
		"contexts",
	)
	namespaceCmd = sets.New(
		"ns",
		"namespace",
		"namespaces",
	)
	dirCmd = sets.New(
		"dir",
		"dirs",
		"d",
		"ls",
	)
	bailCmd = sets.New(
		"q",
		"q!",
		"qa",
		"Q",
		"quit",
		"exit",
	)
	helpCmd = sets.New(
		"?",
		"h",
		"help",
	)
	aliasCmd = sets.New(
		"a",
		"alias",
		"aliases",
	)
	xrayCmd = sets.New(
		"x",
		"xr",
		"xray",
	)
)
