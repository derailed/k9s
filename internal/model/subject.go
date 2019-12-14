package model

import (
	"context"
	"errors"
	"fmt"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/render"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Subject represents a subject model.
type Subject struct {
	Resource

	subjectKind string
}

// List returns a collection of subjects.
func (s *Subject) List(ctx context.Context) ([]runtime.Object, error) {
	var ok bool
	s.subjectKind, ok = ctx.Value(internal.KeySubject).(string)
	if !ok {
		return nil, errors.New("expecting a subject")
	}

	crbs, err := s.factory.List(render.ClusterScope, "rbac.authorization.k8s.io/v1/clusterrolebindings", labels.Everything())
	if err != nil {
		return nil, err
	}

	rbs, err := s.factory.List(render.ClusterScope, "rbac.authorization.k8s.io/v1/rolebindings", labels.Everything())
	if err != nil {
		return nil, err
	}

	return append(crbs, rbs...), nil
}

// Hydrate returns a pod as container rows.
func (s *Subject) Hydrate(oo []runtime.Object, rr render.Rows, re Renderer) error {
	for i, o := range oo {
		res, ok := o.(*unstructured.Unstructured)
		if !ok {
			return fmt.Errorf("expecting unstructured but got %T", o)
		}

		if err := re.Render(res, render.AllNamespaces, &rr[i]); err != nil {
			return err
		}
	}

	return nil
}

func (s *Subject) fetchClusterRoleBindings() ([]runtime.Object, error) {
	oo, err := s.factory.List(render.ClusterScope, "rbac.authorization.k8s.io/v1/clusterrolebindings", labels.Everything())
	if err != nil {
		return nil, err
	}

	rows := make([]runtime.Object, 0, len(oo))
	for _, o := range oo {
		var crb rbacv1.ClusterRoleBinding
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &crb)
		if err != nil {
			return nil, err
		}
		for _, subject := range crb.Subjects {
			if subject.Kind != s.subjectKind {
				continue
			}
			rows = append(rows, SubjectRes{
				id:     subject.Name,
				fields: render.Fields{subject.Name, "ClusterRoleBinding", crb.Name},
			})
		}
	}

	return rows, nil
}

func (s *Subject) fetchRoleBindings() ([]runtime.Object, error) {
	oo, err := s.factory.List(render.ClusterScope, "rbac.authorization.k8s.io/v1/rolebindings", labels.Everything())
	if err != nil {
		return nil, err
	}

	rows := make([]runtime.Object, 0, len(oo))
	for _, o := range oo {
		var rb rbacv1.RoleBinding
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &rb)
		if err != nil {
			return nil, err
		}
		for _, subject := range rb.Subjects {
			if subject.Kind == s.subjectKind {
				rows = append(rows, SubjectRes{
					id:     subject.Name,
					fields: render.Fields{subject.Name, "RoleBinding", rb.Name},
				})
			}
		}
	}

	return rows, nil
}

// ----------------------------------------------------------------------------

// SubjectRes represents a subject resource.
type SubjectRes struct {
	id     string
	fields render.Fields
}

func (s SubjectRes) GetID() string            { return s.id }
func (s SubjectRes) GetFields() render.Fields { return s.fields }

// GetObjectKind returns a schema object.
func (s SubjectRes) GetObjectKind() schema.ObjectKind {
	return nil
}

// DeepCopyObject returns a container copy.
func (s SubjectRes) DeepCopyObject() runtime.Object {
	return s
}
