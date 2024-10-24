// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/tcell/v2"
	"github.com/rs/zerolog/log"
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
		model1.HeaderColumn{Name: "EXTRAS", Wide: true},
	}
}

// Policy renders a rbac policy to screen.
type Policy struct {
	Base
}

// ColorerFunc colors a resource row.
func (Policy) ColorerFunc() model1.ColorerFunc {
	return func(ns string, _ model1.Header, re *model1.RowEvent) tcell.Color {
		return tcell.ColorMediumSpringGreen
	}
}

// Header returns a header row.
func (Policy) Header(ns string) model1.Header {
	h := model1.Header{
		model1.HeaderColumn{Name: "NAMESPACE"},
		model1.HeaderColumn{Name: "NAME"},
		model1.HeaderColumn{Name: "API-GROUP"},
		model1.HeaderColumn{Name: "BINDING"},
	}
	h = append(h, rbacVerbHeader()...)
	h = append(h, model1.HeaderColumn{Name: "VALID", Wide: true})

	return h
}

// Render renders a K8s resource to screen.
func (Policy) Render(o interface{}, gvr string, r *model1.Row) error {
	p, ok := o.(PolicyRes)
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
	if r == "" {
		return r
	}
	if r[0] == '/' {
		return r
	}
	_, n := client.Namespaced(r)
	return n
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
func NewPolicyRes(ns, binding, res, grp string, vv []string) PolicyRes {
	return PolicyRes{
		Namespace: ns,
		Binding:   binding,
		Resource:  res,
		Group:     grp,
		Verbs:     vv,
	}
}

// GR returns the group/resource path.
func (p PolicyRes) GR() string {
	return p.Group + "/" + p.Resource
}

func (p PolicyRes) Merge(p1 PolicyRes) (PolicyRes, error) {
	if p.GR() != p1.GR() {
		return PolicyRes{}, fmt.Errorf("policy mismatch %s vs %s", p.GR(), p1.GR())
	}

	for _, v := range p1.Verbs {
		if !p.hasVerb(v) {
			p.Verbs = append(p.Verbs, v)
		}
	}

	return p, nil
}

func (p PolicyRes) hasVerb(v1 string) bool {
	for _, v := range p.Verbs {
		if v == v1 {
			return true
		}
	}

	return false
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
	idx, ok := pp.find(p.GR())
	if !ok {
		return append(pp, p)
	}
	p, err := pp[idx].Merge(p)
	if err != nil {
		log.Error().Err(err).Msg("policy upsert failed")
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
