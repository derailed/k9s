// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/tcell/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Forwarder represents a port forwarder.
type Forwarder interface {
	// ID returns the PF FQN.
	ID() string

	// Container returns a container name.
	Container() string

	// Port returns container exposed port.
	Port() string

	// Address returns the host address.
	Address() string

	// Active returns forwarder current state.
	Active() bool

	// Age returns forwarder age.
	Age() time.Time
}

// PortForward renders a portforwards to screen.
type PortForward struct {
	Base
}

// ColorerFunc colors a resource row.
func (PortForward) ColorerFunc() model1.ColorerFunc {
	return func(ns string, _ model1.Header, re *model1.RowEvent) tcell.Color {
		return tcell.ColorSkyblue
	}
}

// Header returns a header row.
func (PortForward) Header(ns string) model1.Header {
	return model1.Header{
		model1.HeaderColumn{Name: "NAMESPACE"},
		model1.HeaderColumn{Name: "NAME"},
		model1.HeaderColumn{Name: "CONTAINER"},
		model1.HeaderColumn{Name: "PORTS"},
		model1.HeaderColumn{Name: "URL"},
		model1.HeaderColumn{Name: "C"},
		model1.HeaderColumn{Name: "N"},
		model1.HeaderColumn{Name: "VALID", Wide: true},
		model1.HeaderColumn{Name: "AGE", Time: true},
	}
}

// Render renders a K8s resource to screen.
func (f PortForward) Render(o interface{}, gvr string, r *model1.Row) error {
	pf, ok := o.(ForwardRes)
	if !ok {
		return fmt.Errorf("expecting a ForwardRes but got %T", o)
	}

	ports := strings.Split(pf.Port(), ":")
	r.ID = pf.ID()
	ns, n := client.Namespaced(r.ID)

	r.Fields = model1.Fields{
		ns,
		trimContainer(n),
		pf.Container(),
		pf.Port(),
		UrlFor(pf.Config.Host, pf.Config.Path, ports[0], pf.Address()),
		AsThousands(int64(pf.Config.C)),
		AsThousands(int64(pf.Config.N)),
		"",
		ToAge(metav1.Time{Time: pf.Age()}),
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
func UrlFor(host, path, port, address string) string {
	if host == "" {
		host = address
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
