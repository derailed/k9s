package model

import (
	"context"
	"errors"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/render"
	"k8s.io/apimachinery/pkg/runtime"
)

// Subject represents a subject model.
type Subject struct {
	Resource
}

// List returns a collection of subjects.
func (s *Subject) List(ctx context.Context) ([]runtime.Object, error) {
	kind, ok := ctx.Value(internal.KeySubjectKind).(string)
	if !ok {
		return nil, errors.New("expecting a SubjectKind")
	}

	crbs, err := s.listClusterRoleBindings(kind)
	if err != nil {
		return nil, err
	}

	rbs, err := s.listRoleBindings(kind)
	if err != nil {
		return nil, err
	}

	return append(crbs, rbs...), nil
}

func (s *Subject) listClusterRoleBindings(kind string) ([]runtime.Object, error) {
	crbs, err := fetchClusterRoleBindings(s.factory)
	if err != nil {
		return nil, err
	}

	oo := make([]runtime.Object, 0, len(crbs))
	for _, crb := range crbs {
		for _, su := range crb.Subjects {
			if su.Kind != kind || inSubjectRes(oo, su.Name) {
				continue
			}
			oo = append(oo, render.SubjectRef{
				Name:          su.Name,
				Kind:          "ClusterRoleBinding",
				FirstLocation: crb.Name,
			})
		}
	}

	return oo, nil
}

func (s *Subject) listRoleBindings(kind string) ([]runtime.Object, error) {
	rbs, err := fetchRoleBindings(s.factory)
	if err != nil {
		return nil, err
	}

	oo := make([]runtime.Object, 0, len(rbs))
	for _, rb := range rbs {
		for _, su := range rb.Subjects {
			if su.Kind != kind || inSubjectRes(oo, su.Name) {
				continue
			}
			oo = append(oo, render.SubjectRef{
				Name:          su.Name,
				Kind:          "RoleBinding",
				FirstLocation: rb.Name,
			})
		}
	}

	return oo, nil
}

// ----------------------------------------------------------------------------
// Helpers...

func inSubjectRes(oo []runtime.Object, match string) bool {
	for _, o := range oo {
		res, ok := o.(render.SubjectRef)
		if !ok {
			continue
		}
		if res.Name == match {
			return true
		}
	}
	return false
}
