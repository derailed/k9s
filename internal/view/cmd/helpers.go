// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package cmd

import (
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
func ShouldAddSuggest(suggestionMode, command, suggest string) float64 {
	var searchCondition bool
	var weight float64 = 0

	switch suggestionMode {
	case "LONGEST_SUBSTRING":
		searchCondition = strings.Contains(suggest, command)
		weight = float64(len(command)) / float64(len(suggest))
	case "LONGEST_PREFIX":
		searchCondition = strings.HasPrefix(suggest, command)
		weight = float64(len(command)) / float64(len(suggest))
	default:
		searchCondition = strings.HasPrefix(suggest, command)
	}

	if command != suggest && searchCondition {
		return weight
	}

	return -1
}

// SuggestSubCommand suggests namespaces or contexts based on current command.
func SuggestSubCommand(suggestionMode, command string, namespaces client.NamespaceNames, contexts []string) map[string]float64 {
	p := NewInterpreter(command)
	var suggests map[string]float64
	switch {
	case p.IsCowCmd(), p.IsHelpCmd(), p.IsAliasCmd(), p.IsBailCmd(), p.IsDirCmd():
		return nil

	case p.IsXrayCmd():
		_, ns, ok := p.XrayArgs()
		if !ok || ns == "" {
			return nil
		}
		suggests = completeNS(suggestionMode, ns, namespaces)

	case p.IsContextCmd():
		n, ok := p.ContextArg()
		if !ok {
			return nil
		}
		suggests = completeCtx(suggestionMode, command, n, contexts)

	case p.HasNS():
		if n, ok := p.HasContext(); ok {
			suggests = completeCtx(suggestionMode, command, n, contexts)
		}
		if len(suggests) > 0 {
			break
		}

		ns, ok := p.NSArg()
		if !ok {
			return nil
		}
		suggests = completeNS(suggestionMode, ns, namespaces)

	default:
		if n, ok := p.HasContext(); ok {
			suggests = completeCtx(suggestionMode, command, n, contexts)
		}
	}

	return suggests
}

func completeNS(suggestionMode, s string, nn client.NamespaceNames) map[string]float64 {
	s = strings.ToLower(s)
	suggests := make(map[string]float64)
	if weight := ShouldAddSuggest(suggestionMode, s, client.NamespaceAll); weight != -1 {
		suggests[client.NamespaceAll] = weight
	}
	for ns := range nn {
		if weight := ShouldAddSuggest(suggestionMode, s, ns); weight != -1 {
			suggests[ns] = weight
		}
	}

	if len(suggests) == 0 {
		return nil
	}

	return suggests
}

func completeCtx(suggestionMode, command, s string, contexts []string) map[string]float64 {
	suggests := make(map[string]float64)
	for _, ctxName := range contexts {
		if weight := ShouldAddSuggest(suggestionMode, s, ctxName); weight != -1 {
			if s == "" && !strings.HasSuffix(command, " ") {
				suggests[" "+ctxName] = weight
				continue
			}
			suggests[ctxName] = weight
		}
	}

	if len(suggests) == 0 {
		return nil
	}

	return suggests
}
