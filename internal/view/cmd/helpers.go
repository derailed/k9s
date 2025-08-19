// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package cmd

import (
	"slices"
	"strings"

	"github.com/derailed/k9s/internal/client"
)

func ToLabels(s string) map[string]string {
	var (
		ll   = strings.Split(s, ",")
		lbls = make(map[string]string, len(ll))
	)
	for _, l := range ll {
		kv := strings.Split(l, "=")
		if len(kv) < 2 || kv[0] == "" || kv[1] == "" {
			continue
		}
		lbls[kv[0]] = kv[1]
	}
	if len(lbls) == 0 {
		return nil
	}

	return lbls
}

// ShouldAddSuggest checks if a suggestion match the given command.
func ShouldAddSuggest(command, suggest string) (string, bool) {
	if command != suggest && strings.HasPrefix(suggest, command) {
		return strings.TrimPrefix(suggest, command), true
	}

	return "", false
}

// SuggestSubCommand suggests namespaces or contexts based on current command.
func SuggestSubCommand(command string, namespaces client.NamespaceNames, contexts []string) []string {
	p := NewInterpreter(command)
	var suggests []string
	switch {
	case p.IsCowCmd(), p.IsHelpCmd(), p.IsAliasCmd(), p.IsBailCmd(), p.IsDirCmd():
		return nil

	case p.IsXrayCmd():
		_, ns, ok := p.XrayArgs()
		if !ok || ns == "" {
			return nil
		}
		suggests = completeNS(ns, namespaces)

	case p.IsContextCmd():
		n, ok := p.ContextArg()
		if !ok {
			return nil
		}
		suggests = completeCtx(command, n, contexts)

	case p.HasNS():
		if n, ok := p.HasContext(); ok {
			suggests = completeCtx(command, n, contexts)
		}
		if len(suggests) > 0 {
			break
		}

		ns, ok := p.NSArg()
		if !ok {
			return nil
		}
		suggests = completeNS(ns, namespaces)

	default:
		if n, ok := p.HasContext(); ok {
			suggests = completeCtx(command, n, contexts)
		}
	}
	slices.Sort(suggests)

	return suggests
}

func completeNS(s string, nn client.NamespaceNames) []string {
	s = strings.ToLower(s)
	var suggests []string
	if suggest, ok := ShouldAddSuggest(s, client.NamespaceAll); ok {
		suggests = append(suggests, suggest)
	}
	for ns := range nn {
		if suggest, ok := ShouldAddSuggest(s, ns); ok {
			suggests = append(suggests, suggest)
		}
	}

	return suggests
}

func completeCtx(command, s string, contexts []string) []string {
	var suggests []string
	for _, ctxName := range contexts {
		if suggest, ok := ShouldAddSuggest(s, ctxName); ok {
			if s == "" && !strings.HasSuffix(command, " ") {
				suggests = append(suggests, " "+suggest)
				continue
			}
			suggests = append(suggests, suggest)
		}
	}

	return suggests
}
