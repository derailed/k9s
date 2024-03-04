// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/model1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const allVerbs = "*"

var (
	k8sVerbs = []string{
		"get",
		"list",
		"watch",
		"create",
		"patch",
		"update",
		"delete",
		"deletecollection",
	}

	httpTok8sVerbs = map[string]string{
		"post": "create",
		"put":  "update",
	}
)

// Rbac renders a rbac to screen.
type Rbac struct {
	Base
}

// ColorerFunc colors a resource row.
func (Rbac) ColorerFunc() model1.ColorerFunc {
	return model1.DefaultColorer
}

// Header returns a header row.
func (Rbac) Header(ns string) model1.Header {
	h := make(model1.Header, 0, 10)
	h = append(h,
		model1.HeaderColumn{Name: "NAME"},
		model1.HeaderColumn{Name: "API-GROUP"},
	)
	h = append(h, rbacVerbHeader()...)

	return append(h, model1.HeaderColumn{Name: "VALID", Wide: true})
}

// Render renders a K8s resource to screen.
func (r Rbac) Render(o interface{}, ns string, ro *model1.Row) error {
	p, ok := o.(PolicyRes)
	if !ok {
		return fmt.Errorf("expecting RuleRes but got %T", o)
	}

	ro.ID = p.Resource
	ro.Fields = make(model1.Fields, 0, len(r.Header(ns)))
	ro.Fields = append(ro.Fields,
		cleanseResource(p.Resource),
		p.Group,
	)
	ro.Fields = append(ro.Fields, asVerbs(p.Verbs)...)
	ro.Fields = append(ro.Fields, "")

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

func asVerbs(verbs []string) []string {
	const (
		verbLen    = 4
		unknownLen = 30
	)

	r := make([]string, 0, len(k8sVerbs)+1)
	for _, v := range k8sVerbs {
		r = append(r, toVerbIcon(hasVerb(verbs, v)))
	}

	var unknowns []string
	for _, v := range verbs {
		if hv, ok := httpTok8sVerbs[v]; ok {
			v = hv
		}
		if !hasVerb(k8sVerbs, v) && v != allVerbs {
			unknowns = append(unknowns, v)
		}
	}

	return append(r, Truncate(strings.Join(unknowns, ","), unknownLen))
}

func toVerbIcon(ok bool) string {
	if ok {
		return "[green::b] ✓ [::]"
	}
	return "[orangered::b] × [::]"
}

func hasVerb(verbs []string, verb string) bool {
	if len(verbs) == 1 && verbs[0] == allVerbs {
		return true
	}

	for _, v := range verbs {
		if hv, ok := httpTok8sVerbs[v]; ok {
			if hv == verb {
				return true
			}
		}
		if v == verb {
			return true
		}
	}

	return false
}

// RuleRes represents an rbac rule.
type RuleRes struct {
	Resource, Group string
	ResourceName    string
	NonResourceURL  string
	Verbs           []string
}

// NewRuleRes returns a new rule.
func NewRuleRes(res, grp string, vv []string) RuleRes {
	return RuleRes{
		Resource: res,
		Group:    grp,
		Verbs:    vv,
	}
}

// GetObjectKind returns a schema object.
func (r RuleRes) GetObjectKind() schema.ObjectKind {
	return nil
}

// DeepCopyObject returns a container copy.
func (r RuleRes) DeepCopyObject() runtime.Object {
	return r
}

// Rules represents a collection of rules.
type Rules []RuleRes

// Upsert adds a new rule.
func (rr Rules) Upsert(r RuleRes) Rules {
	idx, ok := rr.find(r.Resource)
	if !ok {
		return append(rr, r)
	}
	rr[idx] = r

	return rr
}

// Find locates a row by id. Returns false is not found.
func (rr Rules) find(res string) (int, bool) {
	for i, r := range rr {
		if r.Resource == res {
			return i, true
		}
	}

	return 0, false
}
