package views

import (
	"fmt"
	"regexp"
	"strings"
)

// K9sEnv represent K9s available env variables.
type K9sEnv map[string]string

// EnvRX match $XXX custom arg.
var envRX = regexp.MustCompile(`\A\$([\w|-]+)`)

func (e K9sEnv) envFor(n string) (string, error) {
	envs := envRX.FindStringSubmatch(n)
	if len(envs) == 0 {
		return n, nil
	}
	env, ok := e[strings.ToUpper(envs[1])]
	if !ok {
		return "", fmt.Errorf("No matching for %s", n)
	}

	return env, nil
}
