// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/xray"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const initTreeRefreshRate = 500 * time.Millisecond

// TreeListener represents a tree model listener.
type TreeListener interface {
	// TreeChanged notifies the model data changed.
	TreeChanged(*xray.TreeNode)

	// TreeLoadFailed notifies the load failed.
	TreeLoadFailed(error)
}

// Tree represents a tree model.
type Tree struct {
	gvr         client.GVR
	namespace   string
	root        *xray.TreeNode
	listeners   []TreeListener
	inUpdate    int32
	refreshRate time.Duration
	query       string
}

// NewTree returns a new model.
func NewTree(gvr client.GVR) *Tree {
	return &Tree{
		gvr:         gvr,
		refreshRate: 2 * time.Second,
	}
}

// ClearFilter clears out active filter.
func (t *Tree) ClearFilter() {
	t.query = ""
}

// SetFilter sets the current filter.
func (t *Tree) SetFilter(q string) {
	t.query = q
}

// AddListener adds a listener.
func (t *Tree) AddListener(l TreeListener) {
	t.listeners = append(t.listeners, l)
}

// RemoveListener delete a listener.
func (t *Tree) RemoveListener(l TreeListener) {
	victim := -1
	for i, lis := range t.listeners {
		if lis == l {
			victim = i
			break
		}
	}

	if victim >= 0 {
		t.listeners = append(t.listeners[:victim], t.listeners[victim+1:]...)
	}
}

// Watch initiates model updates.
func (t *Tree) Watch(ctx context.Context) {
	t.Refresh(ctx)
	go t.updater(ctx)
}

// Refresh update the model now.
func (t *Tree) Refresh(ctx context.Context) {
	t.refresh(ctx)
}

// GetNamespace returns the model namespace.
func (t *Tree) GetNamespace() string {
	return t.namespace
}

// SetNamespace sets up model namespace.
func (t *Tree) SetNamespace(ns string) {
	t.namespace = ns
	if t.root == nil {
		return
	}
	t.root.Clear()
}

// SetRefreshRate sets model refresh duration.
func (t *Tree) SetRefreshRate(d time.Duration) {
	t.refreshRate = d
}

// ClusterWide checks if resource is scope for all namespaces.
func (t *Tree) ClusterWide() bool {
	return client.IsClusterWide(t.namespace)
}

// InNamespace checks if current namespace matches desired namespace.
func (t *Tree) InNamespace(ns string) bool {
	return t.namespace == ns
}

// Empty return true if no model data.
func (t *Tree) Empty() bool {
	return t.root.IsLeaf()
}

// Peek returns model data.
func (t *Tree) Peek() *xray.TreeNode {
	return t.root
}

// Describe describes a given resource.
func (t *Tree) Describe(ctx context.Context, gvr, path string) (string, error) {
	meta, err := t.getMeta(ctx, gvr)
	if err != nil {
		return "", err
	}

	desc, ok := meta.DAO.(dao.Describer)
	if !ok {
		return "", fmt.Errorf("no describer for %q", meta.DAO.GVR())
	}

	return desc.Describe(path)
}

// ToYAML returns a resource yaml.
func (t *Tree) ToYAML(ctx context.Context, gvr, path string) (string, error) {
	meta, err := t.getMeta(ctx, gvr)
	if err != nil {
		return "", err
	}

	desc, ok := meta.DAO.(dao.Describer)
	if !ok {
		return "", fmt.Errorf("no describer for %q", meta.DAO.GVR())
	}

	return desc.ToYAML(path, false)
}

func (t *Tree) updater(ctx context.Context) {
	defer log.Debug().Msgf("Tree-model canceled -- %q", t.gvr)

	rate := initTreeRefreshRate
	for {
		select {
		case <-ctx.Done():
			t.root = nil
			return
		case <-time.After(rate):
			rate = t.refreshRate
			t.refresh(ctx)
		}
	}
}

func (t *Tree) refresh(ctx context.Context) {
	if !atomic.CompareAndSwapInt32(&t.inUpdate, 0, 1) {
		log.Debug().Msgf("Dropping update...")
		return
	}
	defer atomic.StoreInt32(&t.inUpdate, 0)

	if err := t.reconcile(ctx); err != nil {
		log.Error().Err(err).Msg("Reconcile failed")
		t.fireTreeLoadFailed(err)
		return
	}
}

func (t *Tree) list(ctx context.Context, a dao.Accessor) ([]runtime.Object, error) {
	factory, ok := ctx.Value(internal.KeyFactory).(dao.Factory)
	if !ok {
		return nil, fmt.Errorf("expected Factory in context but got %T", ctx.Value(internal.KeyFactory))
	}
	a.Init(factory, t.gvr)

	return a.List(ctx, client.CleanseNamespace(t.namespace))
}

func (t *Tree) reconcile(ctx context.Context) error {
	meta := t.resourceMeta()
	oo, err := t.list(ctx, meta.DAO)
	if err != nil {
		return err
	}

	ns := client.CleanseNamespace(t.namespace)
	res := t.gvr.R()
	root := xray.NewTreeNode(res, res)
	ctx = context.WithValue(ctx, xray.KeyParent, root)
	if _, ok := meta.TreeRenderer.(*xray.Generic); ok {
		table, ok := oo[0].(*metav1.Table)
		if !ok {
			return fmt.Errorf("expecting a Table but got %T", oo[0])
		}
		if err := genericTreeHydrate(ctx, ns, table, meta.TreeRenderer); err != nil {
			return err
		}
	} else if err := treeHydrate(ctx, ns, oo, meta.TreeRenderer); err != nil {
		return err
	}

	root.Sort()
	if t.query != "" {
		t.root = root.Filter(t.query, rxMatch)
	}
	if t.root == nil || t.root.Diff(root) {
		t.root = root
		t.fireTreeChanged(t.root)
	}

	return nil
}

func (t *Tree) resourceMeta() ResourceMeta {
	meta, ok := Registry[t.gvr.String()]
	if !ok {
		meta = ResourceMeta{
			DAO:      &dao.Table{},
			Renderer: &render.Generic{},
		}
	}
	if meta.DAO == nil {
		meta.DAO = &dao.Resource{}
	}

	return meta
}

func (t *Tree) fireTreeChanged(root *xray.TreeNode) {
	for _, l := range t.listeners {
		l.TreeChanged(root)
	}
}

func (t *Tree) fireTreeLoadFailed(err error) {
	for _, l := range t.listeners {
		l.TreeLoadFailed(err)
	}
}

func (t *Tree) getMeta(ctx context.Context, gvr string) (ResourceMeta, error) {
	meta := t.resourceMeta()
	factory, ok := ctx.Value(internal.KeyFactory).(dao.Factory)
	if !ok {
		return ResourceMeta{}, fmt.Errorf("expected Factory in context but got %T", ctx.Value(internal.KeyFactory))
	}
	meta.DAO.Init(factory, client.NewGVR(gvr))

	return meta, nil
}

// ----------------------------------------------------------------------------
// Helpers...

func rxMatch(q, path string) bool {
	rx := regexp.MustCompile(`(?i)` + q)

	tokens := strings.Split(path, "::")
	for _, t := range tokens {
		if rx.MatchString(t) {
			return true
		}
	}
	return false
}

func treeHydrate(ctx context.Context, ns string, oo []runtime.Object, re TreeRenderer) error {
	if re == nil {
		return fmt.Errorf("no tree renderer defined for this resource")
	}
	for _, o := range oo {
		if err := re.Render(ctx, ns, o); err != nil {
			return err
		}
	}

	return nil
}

func genericTreeHydrate(ctx context.Context, ns string, table *metav1.Table, re TreeRenderer) error {
	tre, ok := re.(*xray.Generic)
	if !ok {
		return fmt.Errorf("expecting xray.Generic renderer but got %T", re)
	}

	tre.SetTable(ns, table)
	// BOZO!! Need table row sorter!!
	for _, row := range table.Rows {
		if err := tre.Render(ctx, ns, row); err != nil {
			return err
		}
	}

	return nil
}
