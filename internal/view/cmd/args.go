// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package cmd

import (
	"maps"
	"slices"
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
	arguments := make(args, len(aa))
	if len(aa) == 0 {
		return arguments
	}

	for i := 0; i < len(aa); i++ {
		a := strings.TrimSpace(aa[i])
		switch {
		case strings.Index(a, fuzzyFlag) == 0:
			if a == fuzzyFlag {
				i++
				if i < len(aa) {
					arguments[fuzzyKey] = strings.ToLower(strings.TrimSpace(aa[i]))
				}
			} else {
				arguments[fuzzyKey] = strings.ToLower(a[2:])
			}

		case strings.Index(a, filterFlag) == 0:
			if p.IsDirCmd() {
				if _, ok := arguments[topicKey]; !ok {
					arguments[topicKey] = a
				}
			} else {
				arguments[filterKey] = strings.ToLower(a[1:])
			}

		case isLabelArg(a):
			arguments[labelKey] = strings.ToLower(a)

		case strings.Index(a, contextFlag) == 0:
			arguments[contextKey] = a[1:]

		default:
			switch {
			case p.IsContextCmd():
				arguments[contextKey] = a

			case p.IsDirCmd():
				if _, ok := arguments[topicKey]; !ok {
					arguments[topicKey] = a
				}

			case p.IsXrayCmd():
				if _, ok := arguments[topicKey]; ok {
					arguments[nsKey] = strings.ToLower(a)
				} else {
					arguments[topicKey] = strings.ToLower(a)
				}

			default:
				arguments[nsKey] = strings.ToLower(a)
			}
		}
	}

	return arguments
}

func (a args) String() string {
	ss := make([]string, 0, len(a))
	kk := maps.Keys(a)
	for _, k := range slices.Sorted(kk) {
		v := a[k]
		switch k {
		case filterKey:
			v = filterFlag + v
		case contextKey:
			v = contextFlag + v
		}
		ss = append(ss, v)
	}

	return strings.Join(ss, " ")
}

func (a args) hasFilters() bool {
	_, fok := a[filterKey]
	_, zok := a[fuzzyKey]
	_, lok := a[labelKey]

	return fok || zok || lok
}

func isLabelArg(arg string) bool {
	for _, flag := range labelFlags {
		if strings.Contains(arg, flag) {
			return true
		}
	}

	return false
}
