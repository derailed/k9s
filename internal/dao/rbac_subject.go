// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"errors"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/render"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	_ Accessor = (*Subject)(nil)
	_ Nuker    = (*Subject)(nil)
)

// Subject represents a subject model.
type Subject struct {
	Resource
}

// List returns a collection of subjects.
func (s *Subject) List(ctx context.Context, ns string) ([]runtime.Object, error) {
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

	for _, rb := range rbs {
		crbs = crbs.Upsert(rb)
	}

	oo := make([]runtime.Object, len(crbs))
	for i, o := range crbs {
		oo[i] = o
	}
	return oo, nil
}

func (s *Subject) listClusterRoleBindings(kind string) (render.Subjects, error) {
	crbs, err := fetchClusterRoleBindings(s.Factory)
	if err != nil {
		return nil, err
	}

	oo := make(render.Subjects, 0, len(crbs))
	for _, crb := range crbs {
		for _, su := range crb.Subjects {
			if su.Kind != kind {
				continue
			}
			oo = oo.Upsert(render.SubjectRes{
				Name:          su.Name,
				Kind:          "ClusterRoleBinding",
				FirstLocation: crb.Name,
			})
		}
	}

	return oo, nil
}

func (s *Subject) listRoleBindings(kind string) (render.Subjects, error) {
	rbs, err := fetchRoleBindings(s.Factory)
	if err != nil {
		return nil, err
	}

	oo := make(render.Subjects, 0, len(rbs))
	for _, rb := range rbs {
		for _, su := range rb.Subjects {
			if su.Kind != kind {
				continue
			}
			oo = oo.Upsert(render.SubjectRes{
				Name:          su.Name,
				Kind:          "RoleBinding",
				FirstLocation: rb.Name,
			})
		}
	}

	return oo, nil
}
