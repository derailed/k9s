package view

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Env represent K9s and K8s available environment variables.
type Env map[string]string

// EnvRX match $XXX custom arg.
var envRX = regexp.MustCompile(`\$(\!?[\w|\d|\-|]+)`)

// Substitute replaces env variable keys from in a string with their corresponding values.
func (e Env) Substitute(arg string) (string, error) {
	kk := envRX.FindAllString(arg, -1)
	if len(kk) == 0 {
		return arg, nil
	}

	for _, k := range kk {
		key, inverse := k[1:], false
		if key[0] == '!' {
			key, inverse = key[1:], true
		}
		v, ok := e[strings.ToUpper(key)]
		if !ok {
			return "", fmt.Errorf("no environment matching key %q:%q", k, key)
		}
		if b, err := strconv.ParseBool(v); err == nil {
			if inverse {
				b = !b
			}
			v = fmt.Sprintf("%t", b)
		}
		arg = strings.Replace(arg, k, v, -1)
	}

	return arg, nil
}
