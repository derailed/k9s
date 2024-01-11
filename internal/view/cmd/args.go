// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package cmd

import (
	"strings"
)

const (
	nsKey      = "ns"
	topicKey   = "topic"
	filterKey  = "filter"
	fuzzyKey   = "fuzzy"
	labelKey   = "labels"
	contextKey = "context"
)

type args map[string]string

func newArgs(p *Interpreter, aa []string) args {
	args := make(args, len(aa))
	if len(aa) == 0 {
		return args
	}

	for i := 0; i < len(aa); i++ {
		a := strings.TrimSpace(aa[i])
		switch {
		case strings.Index(a, contextFlag) == 0:
			args[contextKey] = a[1:]

		case strings.Index(a, fuzzyFlag) == 0:
			if a == fuzzyFlag {
				if i++; i < len(aa) {
					args[fuzzyKey] = strings.ToLower(strings.TrimSpace(aa[i]))
				}
			} else {
				args[fuzzyKey] = strings.ToLower(a[2:])
			}

		case strings.Index(a, filterFlag) == 0:
			args[filterKey] = strings.ToLower(a[1:])

		case strings.Contains(a, labelFlag):
			if ll := ToLabels(a); len(ll) != 0 {
				args[labelKey] = strings.ToLower(a)
			}

		default:
			switch {
			case p.IsContextCmd():
				args[contextKey] = a
			case p.IsDirCmd():
				if _, ok := args[topicKey]; !ok {
					args[topicKey] = a
				}
			case p.IsXrayCmd():
				if _, ok := args[topicKey]; ok {
					args[nsKey] = strings.ToLower(a)
				} else {
					args[topicKey] = strings.ToLower(a)
				}
			default:
				args[nsKey] = strings.ToLower(a)
			}
		}
	}

	return args
}

func (a args) hasFilters() bool {
	_, fok := a[filterKey]
	_, zok := a[fuzzyKey]
	_, lok := a[labelKey]

	return fok || zok || lok
}
