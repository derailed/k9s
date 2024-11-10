// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"errors"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ Accessor = (*Reference)(nil)

// Reference represents cluster resource references.
type Reference struct {
	NonResource
}

// List collects all references.
func (r *Reference) List(ctx context.Context, ns string) ([]runtime.Object, error) {
	gvr, ok := ctx.Value(internal.KeyGVR).(client.GVR)
	if !ok {
		return nil, errors.New("no context for gvr found")
	}
	switch gvr {
	case SaGVR:
		return r.ScanSA(ctx)
	default:
		return r.Scan(ctx)
	}
}

// Get fetch a given reference.
func (r *Reference) Get(ctx context.Context, path string) (runtime.Object, error) {
	panic("NYI")
}

// Scan scan cluster resources for references.
func (r *Reference) Scan(ctx context.Context) ([]runtime.Object, error) {
	refs, err := ScanForRefs(ctx, r.Factory)
	if err != nil {
		return nil, err
	}

	fqn, ok := ctx.Value(internal.KeyPath).(string)
	if !ok {
		return nil, errors.New("expecting context Path")
	}
	ns, _ := client.Namespaced(fqn)
	oo := make([]runtime.Object, 0, len(refs))
	for _, ref := range refs {
		_, n := client.Namespaced(ref.FQN)
		oo = append(oo, render.ReferenceRes{
			Namespace: ns,
			Name:      n,
			GVR:       ref.GVR,
		})
	}

	return oo, nil
}

// ScanSA scans for serviceaccount refs.
func (r *Reference) ScanSA(ctx context.Context) ([]runtime.Object, error) {
	refs, err := ScanForSARefs(ctx, r.Factory)
	if err != nil {
		return nil, err
	}

	fqn, ok := ctx.Value(internal.KeyPath).(string)
	if !ok {
		return nil, errors.New("expecting context Path")
	}
	ns, _ := client.Namespaced(fqn)
	oo := make([]runtime.Object, 0, len(refs))
	for _, ref := range refs {
		_, n := client.Namespaced(ref.FQN)
		oo = append(oo, render.ReferenceRes{
			Namespace: ns,
			Name:      n,
			GVR:       ref.GVR,
		})
	}

	return oo, nil
}
