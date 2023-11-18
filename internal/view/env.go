// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

// Env represent K9s and K8s available environment variables.
type Env map[string]string

// EnvRX match $XXX, $!XXX, ${XXX} or ${!XXX} custom arg.
// |
// |                               (g2)(group 3)       (g5)( group 6   )
// |                            (    group 1    ) (       group 4        )
// |                           (                 group 0                  )
var envRX = regexp.MustCompile(`(\$(!?)([\w\-]+))|(\$\{(!?)([\w\-%/: ]+)})`)

// keyFromSubmatch extracts the name and inverse flag of a match.
func keyFromSubmatch(m []string) (key string, inverse bool) {
	// group 1 matches $XXX and $!XXX args.
	if m[1] != "" {
		return m[3], m[2] == "!"
	}
	// group 4 matches ${XXX} and ${!XXX} args.
	return m[6], m[5] == "!"
}

// Substitute replaces env variable keys from in a string with their corresponding values.
func (e Env) Substitute(arg string) (string, error) {
	matches := envRX.FindAllStringSubmatch(arg, -1)
	if len(matches) == 0 {
		return arg, nil
	}

	// To prevent the substitution starts with the shorter environment variable,
	// sort with the length of the found environment variables.
	sort.Slice(matches, func(i, j int) bool {
		return len(matches[i][0]) > len(matches[j][0])
	})

	for _, m := range matches {
		key, inverse := keyFromSubmatch(m)
		v, ok := e[strings.ToUpper(key)]
		if !ok {
			log.Warn().Msgf("no k9s environment matching key %q:%q", m[0], key)
			continue
		}
		if b, err := strconv.ParseBool(v); err == nil {
			if inverse {
				b = !b
			}
			v = fmt.Sprintf("%t", b)
		}
		arg = strings.Replace(arg, m[0], v, -1)
	}

	return arg, nil
}
