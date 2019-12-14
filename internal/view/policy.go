package view

import (
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
)

const (
	group    = "Group"
	user     = "User"
	sa       = "ServiceAccount"
	allVerbs = "*"
)

// Policy presents a RBAC policy viewer.
type Policy struct {
	ResourceViewer
}

// NewPolicy returns a new viewer.
func NewPolicy(gvr client.GVR) *Policy {
	p := Policy{
		ResourceViewer: NewBrowser(gvr),
	}
	p.GetTable().SetColorerFn(render.Policy{}.ColorerFunc())
	p.SetBindKeysFn(p.bindKeys)
	p.GetTable().SetSortCol(1, len(render.Policy{}.Header(render.AllNamespaces)), false)

	return &p
}

func (p *Policy) Name() string {
	return "policy"
}

func (p *Policy) bindKeys(aa ui.KeyActions) {
	aa.Delete(ui.KeyShiftA, tcell.KeyCtrlSpace, ui.KeySpace)
	aa.Add(ui.KeyActions{
		ui.KeyShiftP: ui.NewKeyAction("Sort Namespace", p.GetTable().SortColCmd(0, true), false),
		ui.KeyShiftN: ui.NewKeyAction("Sort Name", p.GetTable().SortColCmd(1, true), false),
		ui.KeyShiftO: ui.NewKeyAction("Sort Group", p.GetTable().SortColCmd(2, true), false),
		ui.KeyShiftB: ui.NewKeyAction("Sort Binding", p.GetTable().SortColCmd(3, true), false),
	})
}

func mapSubject(subject string) string {
	switch subject {
	case "g":
		return group
	case "s":
		return sa
	default:
		return user
	}
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

func toVerbIcon(ok bool) string {
	if ok {
		return "[green::b] ✓ [::]"
	}
	return "[orangered::b] 𐄂 [::]"
}

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

	return append(r, render.Truncate(strings.Join(unknowns, ","), unknownLen))
}
