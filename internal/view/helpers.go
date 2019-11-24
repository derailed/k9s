package view

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func extractApp(ctx context.Context) (*App, error) {
	app, ok := ctx.Value(ui.KeyApp).(*App)
	if !ok {
		return nil, errors.New("No application found in context")
	}

	return app, nil
}

// In check if a string belongs to a set.
func in(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}

	return false
}

// AsKey maps a string representation of a key to a tcell key.
func asKey(key string) (tcell.Key, error) {
	for k, v := range tcell.KeyNames {
		if v == key {
			return k, nil
		}
	}

	return 0, fmt.Errorf("No matching key found %s", key)
}

// FwFQN returns a fully qualified ns/name:container id.
func fwFQN(po, co string) string {
	return po + ":" + co
}

func isTCPPort(p string) bool {
	return !strings.Contains(p, "UDP")
}

// Namespaced converts an fqn resource name to ns and name.
func namespaced(n string) (string, string) {
	ns, po := path.Split(n)
	return strings.Trim(ns, "/"), po
}

// ContainerID computes container ID based on ns/po/co.
func containerID(path, co string) string {
	ns, n := namespaced(path)
	po := strings.Split(n, "-")[0]

	return ns + "/" + po + ":" + co
}

// UrlFor computes fq url for a given benchmark configuration.
func urlFor(cfg config.BenchConfig, port string) string {
	host := "localhost"
	if cfg.HTTP.Host != "" {
		host = cfg.HTTP.Host
	}

	path := "/"
	if cfg.HTTP.Path != "" {
		path = cfg.HTTP.Path
	}

	return "http://" + host + ":" + port + path
}

func fqn(ns, n string) string {
	if ns == "" {
		return n
	}
	return ns + "/" + n
}

// AsNumb prints a number with thousand separator.
func asNum(n int) string {
	p := message.NewPrinter(language.English)
	return p.Sprintf("%d", n)
}
