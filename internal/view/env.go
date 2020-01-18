package view

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/client"
)

// K9sEnv represent K9s available env variables.
type K9sEnv map[string]string

// EnvRX match $XXX custom arg.
var envRX = regexp.MustCompile(`\$([\w]+)(\d*)`)

func (e K9sEnv) envFor(ns, args string) (string, error) {
	envs := envRX.FindStringSubmatch(args)
	if len(envs) == 0 {
		return args, nil
	}

	q := envs[1]
	if envs[2] == "" {
		return e.subOut(args, q)
	}

	var index, err = strconv.Atoi(envs[2])
	if err != nil {
		return args, err
	}
	if client.IsNamespaced(ns) {
		index -= 1
	}
	if index >= 0 {
		q += strconv.Itoa(index)
	}

	return e.subOut(args, q)
}

func (e K9sEnv) subOut(args, q string) (string, error) {
	env, ok := e[strings.ToUpper(q)]
	if !ok {
		return "", fmt.Errorf("no env vars exists for argument %q using key %q", args, q)
	}

	return envRX.ReplaceAllString(args, env), nil
}
