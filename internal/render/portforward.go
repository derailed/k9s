package render

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/gdamore/tcell/v2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Forwarder represents a port forwarder.
type Forwarder interface {
	// ID returns the PF FQN.
	ID() string

	// Container returns a container name.
	Container() string

	// Ports returns container exposed ports.
	Port() string

	// Active returns forwarder current state.
	Active() bool

	// Age returns forwarder age.
	Age() string
}

// PortForward renders a portforwards to screen.
type PortForward struct {
	Base
}

// ColorerFunc colors a resource row.
func (PortForward) ColorerFunc() ColorerFunc {
	return func(ns string, _ Header, re RowEvent) tcell.Color {
		return tcell.ColorSkyblue
	}
}

// Header returns a header row.
func (PortForward) Header(ns string) Header {
	return Header{
		HeaderColumn{Name: "NAMESPACE"},
		HeaderColumn{Name: "NAME"},
		HeaderColumn{Name: "CONTAINER"},
		HeaderColumn{Name: "PORTS"},
		HeaderColumn{Name: "URL"},
		HeaderColumn{Name: "C"},
		HeaderColumn{Name: "N"},
		HeaderColumn{Name: "VALID", Wide: true},
		HeaderColumn{Name: "AGE", Time: true},
	}
}

// Render renders a K8s resource to screen.
func (f PortForward) Render(o interface{}, gvr string, r *Row) error {
	pf, ok := o.(ForwardRes)
	if !ok {
		return fmt.Errorf("expecting a ForwardRes but got %T", o)
	}

	ports := strings.Split(pf.Port(), ":")
	r.ID = pf.ID()
	ns, n := client.Namespaced(r.ID)

	r.Fields = Fields{
		ns,
		trimContainer(n),
		pf.Container(),
		pf.Port(),
		UrlFor(pf.Config.Host, pf.Config.Path, ports[0]),
		AsThousands(int64(pf.Config.C)),
		AsThousands(int64(pf.Config.N)),
		"",
		pf.Age(),
	}

	return nil
}

// Helpers...

func trimContainer(n string) string {
	tokens := strings.Split(n, "|")
	if len(tokens) == 0 {
		return n
	}
	_, name := client.Namespaced(tokens[0])

	return name
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
