package views

import (
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/config"
	res "github.com/derailed/k9s/internal/resource"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	deltaSign = "Δ"
	plusSign  = "↑"
	minusSign = "↓"
)

func isTCPPort(p string) bool {
	return !strings.Contains(p, "UDP")
}

// StripPort removes the named port id if present.
func stripPort(p string) string {
	tokens := strings.Split(p, ":")
	if len(tokens) == 2 {
		return strings.Replace(tokens[1], "╱UDP", "", 1)
	}

	return p
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
func urlFor(cfg config.BenchConfig, co, port string) string {
	host := "localhost"
	if cfg.Host != "" {
		host = cfg.Host
	}

	path := "/"
	if cfg.Path != "" {
		path = cfg.Path
	}

	return "http://" + host + ":" + port + path
}

func fqn(ns, n string) string {
	if ns == "" {
		return n
	}
	return ns + "/" + n
}

func deltas(o, n string) string {
	o, n = strings.TrimSpace(o), strings.TrimSpace(n)
	if o == "" || o == res.NAValue {
		return ""
	}

	if i, ok := numerical(o); ok {
		j, _ := numerical(n)
		switch {
		case i < j:
			return plusSign
		case i > j:
			return minusSign
		default:
			return ""
		}
	}

	if i, ok := percentage(o); ok {
		j, _ := percentage(n)
		switch {
		case i < j:
			return plusSign
		case i > j:
			return minusSign
		default:
			return ""
		}
	}

	if q1, err := resource.ParseQuantity(o); err == nil {
		q2, _ := resource.ParseQuantity(n)
		switch q1.Cmp(q2) {
		case -1:
			return plusSign
		case 1:
			return minusSign
		default:
			return ""
		}
	}

	if d1, err := time.ParseDuration(o); err == nil {
		d2, _ := time.ParseDuration(n)
		switch {
		case d2-d1 > 0:
			return plusSign
		case d2-d1 < 0:
			return minusSign
		default:
			return ""
		}
	}

	switch strings.Compare(o, n) {
	case 1, -1:
		return deltaSign
	default:
		return ""
	}
}

var percent = regexp.MustCompile(`\A(\d+)\%\z`)

func percentage(s string) (int, bool) {
	if res := percent.FindStringSubmatch(s); len(res) == 2 {
		n, _ := strconv.Atoi(res[1])
		return n, true
	}

	return 0, false
}

func numerical(s string) (int, bool) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}

	return n, true
}

// AsNumb prints a number with thousand separator.
func asNum(n int) string {
	p := message.NewPrinter(language.English)
	return p.Sprintf("%d", n)
}
