// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/slogs"
	"github.com/derailed/tcell/v2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func rbacVerbHeader() model1.Header {
	return model1.Header{
		model1.HeaderColumn{Name: "GET   "},
		model1.HeaderColumn{Name: "LIST  "},
		model1.HeaderColumn{Name: "WATCH "},
		model1.HeaderColumn{Name: "CREATE"},
		model1.HeaderColumn{Name: "PATCH "},
		model1.HeaderColumn{Name: "UPDATE"},
		model1.HeaderColumn{Name: "DELETE"},
		model1.HeaderColumn{Name: "DEL-LIST "},
		model1.HeaderColumn{Name: "EXTRAS", Attrs: model1.Attrs{Wide: true}},
	}
}

// Policy renders a rbac policy to screen.
type Policy struct {
	Base
}

// ColorerFunc colors a resource row.
func (Policy) ColorerFunc() model1.ColorerFunc {
	return func(string, model1.Header, *model1.RowEvent) tcell.Color {
		return tcell.ColorMediumSpringGreen
	}
}

// Header returns a header row.
func (Policy) Header(string) model1.Header {
	h := model1.Header{
		model1.HeaderColumn{Name: "NAMESPACE"},
		model1.HeaderColumn{Name: "NAME"},
		model1.HeaderColumn{Name: "API-GROUP"},
		model1.HeaderColumn{Name: "BINDING"},
	}
	h = append(h, rbacVerbHeader()...)
	h = append(h, model1.HeaderColumn{Name: "VALID", Attrs: model1.Attrs{Wide: true}})

	return h
}

// Render renders a K8s resource to screen.
func (Policy) Render(o any, _ string, r *model1.Row) error {
	p, ok := o.(*PolicyRes)
	if !ok {
		return fmt.Errorf("expecting PolicyRes but got %T", o)
	}

	r.ID = client.FQN(p.Namespace, p.Resource)
	r.Fields = append(r.Fields,
		p.Namespace,
		cleanseResource(p.Resource),
		p.Group,
		p.Binding,
	)
	r.Fields = append(r.Fields, asVerbs(p.Verbs)...)
	r.Fields = append(r.Fields, "")

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

func cleanseResource(r string) string {
	if r == "" || r[0] == '/' {
		return r
	}
	tt := strings.Split(r, "/")
	switch len(tt) {
	case 2, 3:
		return strings.TrimPrefix(r, tt[0]+"/")
	default:
		return r
	}
}

// PolicyRes represents a rbac policy rule.
type PolicyRes struct {
	Namespace, Binding string
	Resource, Group    string
	ResourceName       string
	NonResourceURL     string
	Verbs              []string
}

// NewPolicyRes returns a new policy.
func NewPolicyRes(ns, binding, res, grp string, vv []string) *PolicyRes {
	return &PolicyRes{
		Namespace: ns,
		Binding:   binding,
		Resource:  res,
		Group:     grp,
		Verbs:     vv,
	}
}

// GR returns the group/resource path.
func (p *PolicyRes) GR() string {
	return p.Group + "/" + p.Resource
}

// Merge merges two policies.
func (p *PolicyRes) Merge(p1 *PolicyRes) (*PolicyRes, error) {
	if p.GR() != p1.GR() {
		return nil, fmt.Errorf("policy mismatch %s vs %s", p.GR(), p1.GR())
	}

	for _, v := range p1.Verbs {
		if !p.hasVerb(v) {
			p.Verbs = append(p.Verbs, v)
		}
	}

	return p, nil
}

func (p *PolicyRes) hasVerb(v1 string) bool {
	for _, v := range p.Verbs {
		if v == v1 {
			return true
		}
	}

	return false
}

// GetObjectKind returns a schema object.
func (*PolicyRes) GetObjectKind() schema.ObjectKind {
	return nil
}

// DeepCopyObject returns a container copy.
func (p *PolicyRes) DeepCopyObject() runtime.Object {
	return p
}

// Policies represents a collection of RBAC policies.
type Policies []*PolicyRes

// Upsert adds a new policy.
func (pp Policies) Upsert(p *PolicyRes) Policies {
	idx, ok := pp.find(p.GR())
	if !ok {
		return append(pp, p)
	}
	p, err := pp[idx].Merge(p)
	if err != nil {
		slog.Error("Policy upsert failed", slogs.Error, err)
		return pp
	}
	pp[idx] = p

	return pp
}

// Find locates a row by id. Returns false is not found.
func (pp Policies) find(gr string) (int, bool) {
	for i, p := range pp {
		if p.GR() == gr {
			return i, true
		}
	}

	return 0, false
}
