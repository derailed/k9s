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

var defaultROBHeader = model1.Header{
	model1.HeaderColumn{Name: "NAMESPACE"},
	model1.HeaderColumn{Name: "NAME"},
	model1.HeaderColumn{Name: "ROLE"},
	model1.HeaderColumn{Name: "KIND"},
	model1.HeaderColumn{Name: "SUBJECTS"},
	model1.HeaderColumn{Name: "LABELS", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "VALID", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "AGE", Attrs: model1.Attrs{Time: true}},
}

// RoleBinding renders a K8s RoleBinding to screen.
type RoleBinding struct {
	Base
}

// Header returns a header row.
func (r RoleBinding) Header(_ string) model1.Header {
	return r.doHeader(defaultROBHeader)
}

// Render renders a K8s resource to screen.
func (r RoleBinding) Render(o any, _ string, row *model1.Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected Unstructured, but got %T", o)
	}
	if err := r.defaultRow(raw, row); err != nil {
		return err
	}
	if r.specs.isEmpty() {
		return nil
	}
	cols, err := r.specs.realize(raw, defaultROBHeader, row)
	if err != nil {
		return err
	}
	cols.hydrateRow(row)

	return nil
}

func (RoleBinding) defaultRow(raw *unstructured.Unstructured, row *model1.Row) error {
	var rb rbacv1.RoleBinding
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &rb)
	if err != nil {
		return err
	}

	kind, ss := renderSubjects(rb.Subjects)

	row.ID = client.MetaFQN(&rb.ObjectMeta)
	row.Fields = model1.Fields{
		rb.Namespace,
		rb.Name,
		rb.RoleRef.Name,
		kind,
		ss,
		mapToStr(rb.Labels),
		"",
		ToAge(rb.GetCreationTimestamp()),
	}

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

func renderSubjects(ss []rbacv1.Subject) (kind, subjects string) {
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
	if s == "" {
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
