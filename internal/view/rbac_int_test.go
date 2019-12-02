package view

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/resource"
	"github.com/stretchr/testify/assert"
	rbacv1 "k8s.io/api/rbac/v1"
)

func TestHasVerb(t *testing.T) {
	uu := []struct {
		vv []string
		v  string
		e  bool
	}{
		{[]string{"*"}, "get", true},
		{[]string{"get", "list", "watch"}, "watch", true},
		{[]string{"get", "dope", "list"}, "watch", false},
		{[]string{"get"}, "get", true},
		{[]string{"post"}, "create", true},
		{[]string{"put"}, "update", true},
		{[]string{"list", "deletecollection"}, "deletecollection", true},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, hasVerb(u.vv, u.v))
	}
}

func TestAsVerbs(t *testing.T) {
	ok, nok := toVerbIcon(true), toVerbIcon(false)

	uu := []struct {
		vv []string
		e  render.Row
	}{
		{[]string{"*"}, render.Row{Fields: render.Fields{ok, ok, ok, ok, ok, ok, ok, ok, ""}}},
		{[]string{"get", "list", "patch"}, render.Row{Fields: render.Fields{ok, ok, nok, nok, nok, ok, nok, nok, ""}}},
		{[]string{"get", "list", "deletecollection", "post"}, render.Row{Fields: render.Fields{ok, ok, ok, nok, ok, nok, nok, nok, ""}}},
		{[]string{"get", "list", "blee"}, render.Row{Fields: render.Fields{ok, ok, nok, nok, nok, nok, nok, nok, "blee"}}},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, asVerbs(u.vv...))
	}
}

func TestParseRules(t *testing.T) {
	ok, nok := toVerbIcon(true), toVerbIcon(false)
	_ = nok

	uu := []struct {
		pp []rbacv1.PolicyRule
		e  render.Rows
	}{
		{
			[]rbacv1.PolicyRule{
				{APIGroups: []string{"*"}, Resources: []string{"*"}, Verbs: []string{"*"}},
			},
			render.Rows{
				render.Row{Fields: render.Fields{"*.*", "*", ok, ok, ok, ok, ok, ok, ok, ok, ""}},
			},
		},
		{
			[]rbacv1.PolicyRule{
				{APIGroups: []string{"*"}, Resources: []string{"*"}, Verbs: []string{"get"}},
			},
			render.Rows{
				render.Row{Fields: render.Fields{"*.*", "*", ok, nok, nok, nok, nok, nok, nok, nok, ""}},
			},
		},
		{
			[]rbacv1.PolicyRule{
				{APIGroups: []string{""}, Resources: []string{"*"}, Verbs: []string{"list"}},
			},
			render.Rows{
				render.Row{Fields: render.Fields{"*", "v1", nok, ok, nok, nok, nok, nok, nok, nok, ""}},
			},
		},
		{
			[]rbacv1.PolicyRule{
				{APIGroups: []string{""}, Resources: []string{"pods"}, Verbs: []string{"list"}, ResourceNames: []string{"fred"}},
			},
			render.Rows{
				render.Row{Fields: render.Fields{"pods", "v1", nok, ok, nok, nok, nok, nok, nok, nok, ""}},
				render.Row{Fields: render.Fields{"pods/fred", "v1", nok, ok, nok, nok, nok, nok, nok, nok, ""}},
			},
		},
		{
			[]rbacv1.PolicyRule{
				{APIGroups: []string{}, Resources: []string{}, Verbs: []string{"get"}, NonResourceURLs: []string{"/fred"}},
			},
			render.Rows{
				render.Row{Fields: render.Fields{"/fred", resource.NAValue, ok, nok, nok, nok, nok, nok, nok, nok, ""}},
			},
		},
		{
			[]rbacv1.PolicyRule{
				{APIGroups: []string{}, Resources: []string{}, Verbs: []string{"get"}, NonResourceURLs: []string{"fred"}},
			},
			render.Rows{
				render.Row{Fields: render.Fields{"/fred", resource.NAValue, ok, nok, nok, nok, nok, nok, nok, nok, ""}},
			},
		},
	}

	var v Rbac
	for _, u := range uu {
		evts := v.parseRules(u.pp)
		for k, v := range u.e {
			assert.Equal(t, v, evts[k].Fields)
		}
	}
}
