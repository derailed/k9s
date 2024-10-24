// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// RoleBinding renders a K8s RoleBinding to screen.
type RoleBinding struct {
	Base
}

// Header returns a header rbw.
func (RoleBinding) Header(ns string) model1.Header {
	var h model1.Header
	if client.IsAllNamespaces(ns) {
		h = append(h, model1.HeaderColumn{Name: "NAMESPACE"})
	}

	return append(h,
		model1.HeaderColumn{Name: "NAME"},
		model1.HeaderColumn{Name: "ROLE"},
		model1.HeaderColumn{Name: "KIND"},
		model1.HeaderColumn{Name: "SUBJECTS"},
		model1.HeaderColumn{Name: "LABELS", Wide: true},
		model1.HeaderColumn{Name: "VALID", Wide: true},
		model1.HeaderColumn{Name: "AGE", Time: true},
	)
}

// Render renders a K8s resource to screen.
func (r RoleBinding) Render(o interface{}, ns string, row *model1.Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected RoleBinding, but got %T", o)
	}
	var rb rbacv1.RoleBinding
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &rb)
	if err != nil {
		return err
	}

	kind, ss := renderSubjects(rb.Subjects)

	row.ID = client.MetaFQN(rb.ObjectMeta)
	row.Fields = make(model1.Fields, 0, len(r.Header(ns)))
	if client.IsAllNamespaces(ns) {
		row.Fields = append(row.Fields, rb.Namespace)
	}
	row.Fields = append(row.Fields,
		rb.Name,
		rb.RoleRef.Name,
		kind,
		ss,
		mapToStr(rb.Labels),
		"",
		ToAge(rb.GetCreationTimestamp()),
	)

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

func renderSubjects(ss []rbacv1.Subject) (kind string, subjects string) {
	if len(ss) == 0 {
		return NAValue, ""
	}

	tt := make([]string, 0, len(ss))
	for _, s := range ss {
		kind = toSubjectAlias(s.Kind)
		tt = append(tt, s.Name)
	}
	return kind, strings.Join(tt, ",")
}

func toSubjectAlias(s string) string {
	if len(s) == 0 {
		return s
	}

	switch s {
	case rbacv1.UserKind:
		return "User"
	case rbacv1.GroupKind:
		return "Group"
	case rbacv1.ServiceAccountKind:
		return "SvcAcct"
	default:
		return strings.ToUpper(s)
	}
}
