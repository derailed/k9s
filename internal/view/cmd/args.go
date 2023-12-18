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
			i++
			args[filterKey] = strings.TrimSpace(aa[i])
			continue

		case strings.Index(a, filterFlag) == 0:
			args[filterKey] = a[1:]

		case strings.Contains(a, labelFlag):
			if ll := toLabels(a); len(ll) != 0 {
				args[labelKey] = a
			}

		default:
			a := strings.TrimSpace(aa[i])
			switch {
			case p.IsContextCmd():
				args[contextKey] = a
			case p.IsDirCmd():
				if _, ok := args[topicKey]; !ok {
					args[topicKey] = a
				}
			case p.IsXrayCmd():
				if _, ok := args[topicKey]; ok {
					args[nsKey] = a
				} else {
					args[topicKey] = a
				}
			default:
				args[nsKey] = a
			}
		}
	}

	return args
}

func (a args) hasFilters() bool {
	_, fok := a[filterKey]
	_, lok := a[labelKey]

	return fok || lok
}
