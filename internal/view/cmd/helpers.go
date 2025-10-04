// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package cmd

import (
	"slices"
	"strings"

	"github.com/derailed/k9s/internal/client"
)

// ToLabels converts a string into a map of labels.
func ToLabels(s string) map[string]string {
	var (
		ll   = strings.Split(s, ",")
		lbls = make(map[string]string, len(ll))
	)
	for _, l := range ll {
		if k, v, ok := splitKv(l); ok {
			lbls[k] = v
		} else {
			continue
		}
	}
	if len(lbls) == 0 {
		return nil
	}

	return lbls
}

func splitKv(s string) (k, v string, ok bool) {
	switch {
	case strings.Contains(s, labelFlagNotEq):
		kv := strings.SplitN(s, labelFlagNotEq, 2)
		if len(kv) == 2 && kv[0] != "" && kv[1] != "" {
			return strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1]), true
		}
	case strings.Contains(s, labelFlagEqs):
		kv := strings.SplitN(s, labelFlagEqs, 2)
		if len(kv) == 2 && kv[0] != "" && kv[1] != "" {
			return strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1]), true
		}
	case strings.Contains(s, labelFlagEq):
		kv := strings.SplitN(s, labelFlagEq, 2)
		if len(kv) == 2 && kv[0] != "" && kv[1] != "" {
			return strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1]), true
		}
	}

	return "", "", false
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
