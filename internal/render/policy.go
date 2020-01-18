package render

import (
	"fmt"

	"github.com/derailed/k9s/internal/client"
	"github.com/gdamore/tcell"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func rbacVerbHeader() HeaderRow {
	return HeaderRow{
		Header{Name: "GET   "},
		Header{Name: "LIST  "},
		Header{Name: "WATCH "},
		Header{Name: "CREATE"},
		Header{Name: "PATCH "},
		Header{Name: "UPDATE"},
		Header{Name: "DELETE"},
		Header{Name: "DLIST "},
		Header{Name: "EXTRAS"},
	}
}

// Policy renders a rbac policy to screen.
type Policy struct{}

// ColorerFunc colors a resource row.
func (Policy) ColorerFunc() ColorerFunc {
	return func(ns string, re RowEvent) tcell.Color {
		return tcell.ColorMediumSpringGreen
	}
}

// Header returns a header row.
func (Policy) Header(ns string) HeaderRow {
	h := HeaderRow{
		Header{Name: "NAMESPACE"},
		Header{Name: "NAME"},
		Header{Name: "API GROUP"},
		Header{Name: "BINDING"},
	}
	return append(h, rbacVerbHeader()...)
}

// Render renders a K8s resource to screen.
func (Policy) Render(o interface{}, gvr string, r *Row) error {
	p, ok := o.(PolicyRes)
	if !ok {
		return fmt.Errorf("expecting PolicyRes but got %T", o)
	}

	r.ID = client.FQN(p.Namespace, p.Resource)
	r.Fields = append(r.Fields, p.Namespace, cleanseResource(p.Resource), p.Group, p.Binding)
	r.Fields = append(r.Fields, asVerbs(p.Verbs)...)

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

func cleanseResource(r string) string {
	if r[0] == '/' {
		return r
	}
	_, n := client.Namespaced(r)
	return n
}

// PolicyRes represents a rback policy rule.
type PolicyRes struct {
	Namespace, Binding string
	Resource, Group    string
	ResourceName       string
	NonResourceURL     string
	Verbs              []string
}

// NewPolicyRes returns a new policy.
func NewPolicyRes(ns, binding, res, grp string, vv []string) PolicyRes {
	return PolicyRes{
		Namespace: ns,
		Binding:   binding,
		Resource:  res,
		Group:     grp,
		Verbs:     vv,
	}
}

// GetObjectKind returns a schema object.
func (p PolicyRes) GetObjectKind() schema.ObjectKind {
	return nil
}

// DeepCopyObject returns a container copy.
func (p PolicyRes) DeepCopyObject() runtime.Object {
	return p
}

// Policies represents a collection of RBAC policies.
type Policies []PolicyRes

// Upsert adds a new policy.
func (pp Policies) Upsert(p PolicyRes) Policies {
	idx, ok := pp.find(p.Resource)
	if !ok {
		return append(pp, p)
	}
	pp[idx] = p

	return pp
}

// Find locates a row by id. Retturns false is not found.
func (pp Policies) find(res string) (int, bool) {
	for i, p := range pp {
		if p.Resource == res {
			return i, true
		}
	}

	return 0, false
}
