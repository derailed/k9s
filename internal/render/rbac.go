package render

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell"
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
type Rbac struct{}

// ColorerFunc colors a resource row.
func (Rbac) ColorerFunc() ColorerFunc {
	return func(ns string, re RowEvent) tcell.Color {
		return tcell.ColorMediumSpringGreen
	}
}

// Header returns a header row.
func (Rbac) Header(ns string) HeaderRow {
	h := HeaderRow{
		Header{Name: "NAME"},
		Header{Name: "API GROUP"},
	}

	return append(h, rbacVerbHeader()...)
}

// Render renders a K8s resource to screen.
func (Rbac) Render(o interface{}, gvr string, r *Row) error {
	p, ok := o.(PolicyRes)
	if !ok {
		return fmt.Errorf("expecting RuleRes but got %T", o)
	}

	r.ID = p.Resource
	r.Fields = append(r.Fields, cleanseResource(p.Resource), p.Group)
	r.Fields = append(r.Fields, asVerbs(p.Verbs)...)

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
	return "[orangered::b] 𐄂 [::]"
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

// Find locates a row by id. Retturns false is not found.
func (rr Rules) find(res string) (int, bool) {
	for i, r := range rr {
		if r.Resource == res {
			return i, true
		}
	}

	return 0, false
}
