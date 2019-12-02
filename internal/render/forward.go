package render

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell"
)

// Forwarder represents a port forwarder.
type Forwarder interface {
	// Path returns a resource FQN.
	Path() string

	// Container returns a container name.
	Container() string

	// Ports returns container exposed ports.
	Ports() []string

	// Active returns forwarder current state.
	Active() bool

	// Age returns forwarder age.
	Age() string
}

// Forward renders a portforwards to screen.
type Forward struct{}

// ColorerFunc colors a resource row.
func (Forward) ColorerFunc() ColorerFunc {
	return func(ns string, re RowEvent) tcell.Color {
		return tcell.ColorSkyblue
	}
}

// Header returns a header row.
func (Forward) Header(ns string) HeaderRow {
	return HeaderRow{
		Header{Name: "NAMESPACE"},
		Header{Name: "NAME"},
		Header{Name: "CONTAINER"},
		Header{Name: "PORTS"},
		Header{Name: "URL"},
		Header{Name: "C"},
		Header{Name: "N"},
		Header{Name: "AGE", Decorator: ageDecorator},
	}
}

// Render renders a K8s resource to screen.
func (f Forward) Render(o interface{}, gvr string, r *Row) error {
	pf, ok := o.(PortForwarder)
	if !ok {
		return fmt.Errorf("expecting a portforward but got %T", o)
	}

	ports := strings.Split(pf.Ports()[0], ":")
	ns, na := Namespaced(pf.Path())

	r.ID = pf.Path()
	r.Fields = Fields{
		ns,
		na,
		pf.Container(),
		strings.Join(pf.Ports(), ","),
		UrlFor(pf.Host(), pf.HttpPath(), ports[0]),
		asNum(pf.C()),
		asNum(pf.N()),
		pf.Age(),
	}

	return nil
}

// Helpers...

type PortForwarder interface {
	Forwarder
	BenchConfigurator
}

type BenchConfigurators map[string]BenchConfigurator

type BenchConfigurator interface {
	// C returns the number of concurent connections.
	C() int

	// N returns the number of requests.
	N() int

	// Host returns the forward host address.
	Host() string

	// Path returns the http path.
	HttpPath() string
}

// UrlFor computes fq url for a given benchmark configuration.
func UrlFor(host, path, port string) string {
	if host == "" {
		host = "localhost"
	}
	if path == "" {
		path = "/"
	}

	return "http://" + host + ":" + port + path
}
