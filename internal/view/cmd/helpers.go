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
	if ctx, quote, ok := quotedContextPrefix(command); ok {
		suggests := completeCtx(command+" ", ctx, contexts, true)
		qq := string(quote)
		if quote != 0 {
			for i, s := range suggests {
				suggests[i] = s + qq
			}
			if ctx != "" && isExactContext(ctx, contexts) && !slices.Contains(suggests, qq) {
				suggests = append(suggests, qq)
			}
		}
		slices.Sort(suggests)
		return suggests
	}

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
		suggests = completeCtx(command, n, contexts, true)
		if len(suggests) == 0 {
			suggests = completeCtxToken(n, contexts)
		}

	case p.HasNS():
		if n, ok := p.HasContext(); ok {
			suggests = completeCtx(command, n, contexts, false)
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
			suggests = completeCtx(command, n, contexts, false)
		}
	}
	slices.Sort(suggests)

	return suggests
}

func isExactContext(ctx string, contexts []string) bool {
	for _, ctxName := range contexts {
		if ctxName == ctx {
			return true
		}
	}
	return false
}

func quotedContextPrefix(command string) (string, byte, bool) {
	at := strings.LastIndex(command, "@")
	if at == -1 {
		return "", 0, false
	}
	if at > 0 && !isWhitespace(command[at-1]) {
		return "", 0, false
	}
	tail := command[at+1:]
	if tail == "" {
		return "", 0, false
	}
	quote := tail[0]
	if quote != '"' && quote != '\'' {
		return "", 0, false
	}
	rest := tail[1:]
	if strings.IndexByte(rest, quote) != -1 {
		return "", 0, false
	}

	return rest, quote, true
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

func completeCtx(command, s string, contexts []string, allowSpaces bool) []string {
	var suggests []string
	seen := make(map[string]struct{}, len(contexts))
	disallowedMatch := false
	for _, ctxName := range contexts {
		if !allowSpaces && strings.IndexAny(ctxName, " \t") != -1 {
			if strings.HasPrefix(ctxName, s) {
				disallowedMatch = true
			}
			continue
		}
		if _, ok := seen[ctxName]; ok {
			continue
		}
		seen[ctxName] = struct{}{}
		if suggest, ok := ShouldAddSuggest(s, ctxName); ok {
			if s == "" && !strings.HasSuffix(command, " ") {
				suggests = append(suggests, " "+suggest)
				continue
			}
			suggests = append(suggests, suggest)
		}
	}

	if len(suggests) == 0 && disallowedMatch {
		return []string{""}
	}

	return suggests
}

func completeCtxToken(needle string, contexts []string) []string {
	idx := strings.LastIndexAny(needle, " \t")
	if idx == -1 {
		return nil
	}
	prefix := strings.TrimSpace(needle[:idx])
	tail := strings.TrimSpace(needle[idx+1:])
	if prefix == "" || tail == "" {
		return nil
	}
	var suggests []string
	seen := make(map[string]struct{}, len(contexts))
	for _, ctxName := range contexts {
		if !strings.HasPrefix(ctxName, prefix) {
			continue
		}
		if len(ctxName) <= len(prefix) || !isWhitespace(ctxName[len(prefix)]) {
			continue
		}
		for _, token := range strings.Fields(ctxName[len(prefix):]) {
			if strings.HasPrefix(token, tail) {
				remain := strings.TrimPrefix(token, tail)
				if remain == "" {
					continue
				}
				if _, ok := seen[remain]; ok {
					continue
				}
				seen[remain] = struct{}{}
				suggests = append(suggests, remain)
			}
		}
	}

	return suggests
}
