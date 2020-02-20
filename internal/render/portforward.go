package render

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/gdamore/tcell"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

// PortForward renders a portforwards to screen.
type PortForward struct{}

// ColorerFunc colors a resource row.
func (PortForward) ColorerFunc() ColorerFunc {
	return func(ns string, re RowEvent) tcell.Color {
		return tcell.ColorSkyblue
	}
}

// Header returns a header row.
func (PortForward) Header(ns string) HeaderRow {
	return HeaderRow{
		Header{Name: "NAMESPACE"},
		Header{Name: "POD"},
		Header{Name: "CONTAINER"},
		Header{Name: "PORTS"},
		Header{Name: "URL"},
		Header{Name: "C"},
		Header{Name: "N"},
		Header{Name: "VALID", Wide: true},
		Header{Name: "AGE", Decorator: AgeDecorator},
	}
}

// Render renders a K8s resource to screen.
func (f PortForward) Render(o interface{}, gvr string, r *Row) error {
	pf, ok := o.(ForwardRes)
	if !ok {
		return fmt.Errorf("expecting a ForwardRes but got %T", o)
	}

	ports := strings.Split(pf.Ports()[0], ":")
	ns, n := client.Namespaced(pf.Path())

	r.ID = pf.Path()
	r.Fields = Fields{
		ns,
		trimContainer(n),
		pf.Container(),
		strings.Join(pf.Ports(), ","),
		UrlFor(pf.Config.Host, pf.Config.Path, ports[0]),
		asNum(pf.Config.C),
		asNum(pf.Config.N),
		"",
		pf.Age(),
	}

	return nil
}

// Helpers...

func trimContainer(n string) string {
	tokens := strings.Split(n, ":")
	if len(tokens) == 0 {
		return n
	}
	return tokens[0]
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

// BenchCfg represents a benchmark configuration.
type BenchCfg struct {
	C, N       int
	Host, Path string
}

// ForwardRes represents a benchmark resource.
type ForwardRes struct {
	Forwarder
	Config BenchCfg
}

// GetObjectKind returns a schema object.
func (f ForwardRes) GetObjectKind() schema.ObjectKind {
	return nil
}

// DeepCopyObject returns a container copy.
func (f ForwardRes) DeepCopyObject() runtime.Object {
	return f
}
