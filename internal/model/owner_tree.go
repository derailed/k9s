// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/slogs"
	"github.com/derailed/k9s/internal/xray"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

// Ensure OwnerTree satisfies the TreeModel interface.
var _ TreeModel = (*OwnerTree)(nil)

// OwnerTree represents a tree model that builds its hierarchy from
// OwnerReferences, traversing all resource types in the namespace.
type OwnerTree struct {
	gvr         *client.GVR
	namespace   string
	root        *xray.TreeNode
	listeners   []TreeListener
	inUpdate    int32
	refreshRate time.Duration
	query       string
}

// NewOwnerTree returns a new OwnerTree model for the given root GVR.
func NewOwnerTree(gvr *client.GVR) *OwnerTree {
	return &OwnerTree{
		gvr:         gvr,
		refreshRate: 2 * time.Second,
	}
}

// SetRefreshRate sets model refresh duration.
func (t *OwnerTree) SetRefreshRate(d time.Duration) {
	t.refreshRate = d
}

// SetNamespace sets up model namespace.
func (t *OwnerTree) SetNamespace(ns string) {
	t.namespace = ns
	if t.root == nil {
		return
	}
	t.root.Clear()
}

// GetNamespace returns the model namespace.
func (t *OwnerTree) GetNamespace() string {
	return t.namespace
}

// AddListener adds a listener.
func (t *OwnerTree) AddListener(l TreeListener) {
	t.listeners = append(t.listeners, l)
}

// Watch initiates model updates.
func (t *OwnerTree) Watch(ctx context.Context) {
	t.refresh(ctx)
	go t.updater(ctx)
}

// Peek returns model data.
func (t *OwnerTree) Peek() *xray.TreeNode {
	return t.root
}

// ClearFilter clears out active filter.
func (t *OwnerTree) ClearFilter() {
	t.query = ""
}

// SetFilter sets the current filter.
func (t *OwnerTree) SetFilter(q string) {
	t.query = q
}

// ToYAML returns the YAML representation of the given resource.
func (*OwnerTree) ToYAML(ctx context.Context, gvr *client.GVR, path string) (string, error) {
	factory, ok := ctx.Value(internal.KeyFactory).(dao.Factory)
	if !ok {
		return "", fmt.Errorf("expected Factory in context but got %T", ctx.Value(internal.KeyFactory))
	}
	g := new(dao.Generic)
	g.Init(factory, gvr)

	return g.ToYAML(path, false)
}

// Describe returns a textual description of the given resource.
func (*OwnerTree) Describe(ctx context.Context, gvr *client.GVR, path string) (string, error) {
	factory, ok := ctx.Value(internal.KeyFactory).(dao.Factory)
	if !ok {
		return "", fmt.Errorf("expected Factory in context but got %T", ctx.Value(internal.KeyFactory))
	}
	g := new(dao.Generic)
	g.Init(factory, gvr)

	return g.Describe(path)
}

// ----------------------------------------------------------------------------
// Internals

func (t *OwnerTree) updater(ctx context.Context) {
	defer slog.Debug("OwnerTree model canceled", slogs.GVR, t.gvr)

	for {
		select {
		case <-ctx.Done():
			t.root = nil
			return
		case <-time.After(t.refreshRate):
			t.refresh(ctx)
		}
	}
}

func (t *OwnerTree) refresh(ctx context.Context) {
	if !atomic.CompareAndSwapInt32(&t.inUpdate, 0, 1) {
		slog.Debug("OwnerTree dropping update...")
		return
	}
	defer atomic.StoreInt32(&t.inUpdate, 0)

	if err := t.reconcile(ctx); err != nil {
		slog.Error("OwnerTree reconcile failed", slogs.Error, err)
		t.fireTreeLoadFailed(err)
	}
}

// reconcile builds the owner-reference tree for the model's root GVR.
//
// Algorithm:
//  1. List all resources in the namespace across every known GVR.
//  2. Build an ownerUID→children reverse index.
//  3. List root resources (of the model's GVR).
//  4. Recursively attach owned resources as children in the tree.
func (t *OwnerTree) reconcile(ctx context.Context) error {
	reconcileStart := time.Now()

	factory, ok := ctx.Value(internal.KeyFactory).(dao.Factory)
	if !ok {
		return fmt.Errorf("expected Factory in context but got %T", ctx.Value(internal.KeyFactory))
	}

	ns := client.CleanseNamespace(t.namespace)

	// Step 1 & 2: build the reverse ownerRef index.
	indexStart := time.Now()
	ownerIndex := buildOwnerIndex(factory, ns)
	slog.Debug("OwnerTree: buildOwnerIndex", slogs.Duration, time.Since(indexStart), slogs.Entries, len(ownerIndex))
	// Step 3: list root resources from the informer cache (no direct API call).
	listStart := time.Now()
	rootObjs, err := factory.List(t.gvr, ns, false, labels.Everything())
	slog.Debug("OwnerTree: factory.List root GVR", slogs.GVR, t.gvr, slogs.Duration, time.Since(listStart), slogs.Count, len(rootObjs))
	if err != nil {
		return fmt.Errorf("listing root GVR %s: %w", t.gvr, err)
	}

	// Step 4: build the tree.
	buildStart := time.Now()
	root := xray.NewTreeNode(t.gvr, t.gvr.R())
	for _, o := range rootObjs {
		u, ok := o.(*unstructured.Unstructured)
		if !ok {
			continue
		}
		nodeFQN := resourceFQN(u)
		slog.Info("OwnerTree: root resource", slogs.FQN, nodeFQN, slogs.UID, u.GetUID(), slogs.ChildrenInIndex, len(ownerIndex[u.GetUID()]))
		node := xray.NewTreeNode(t.gvr, nodeFQN)
		setOwnerNodeStatus(node, u)
		buildOwnerChildren(node, u.GetUID(), ownerIndex, map[types.UID]struct{}{u.GetUID(): {}})
		root.Add(node)
	}
	slog.Debug("OwnerTree: tree build", "duration", time.Since(buildStart))

	sortStart := time.Now()
	root.Sort()
	slog.Debug("OwnerTree: sort", "duration", time.Since(sortStart))

	resultRoot := root
	if t.query != "" {
		filterStart := time.Now()
		rx, err := regexp.Compile(`(?i)` + t.query)
		if err != nil {
			slog.Warn("OwnerTree: invalid filter regex", "query", t.query, "error", err)
		} else if filtered := root.Filter(t.query, func(_, path string) bool {
			for _, tok := range strings.Split(path, "::") {
				if rx.MatchString(tok) {
					return true
				}
			}
			return false
		}); filtered != nil {
			resultRoot = filtered
		}
		slog.Debug("OwnerTree: filter", "query", t.query, "duration", time.Since(filterStart))
	}

	if t.root == nil || t.root.Diff(resultRoot) {
		t.root = resultRoot
		t.fireTreeChanged(t.root)
	}

	slog.Debug("OwnerTree: reconcile total", slogs.GVR, t.gvr, "ns", ns, "duration", time.Since(reconcileStart))

	return nil
}

// buildOwnerIndex builds a map of ownerUID → owned resources using the
// informer lister cache (in-memory, no network calls). Resources whose
// informer is not yet synced are silently skipped — they will appear on
// the next reconcile cycle once the watch stream has caught up.
func buildOwnerIndex(factory dao.Factory, ns string) map[types.UID][]*unstructured.Unstructured {
	totalStart := time.Now()

	// Collect all namespaced, listable GVRs first.
	gvrScanStart := time.Now()
	var gvrs []*client.GVR
	for _, gvr := range dao.MetaAccess.AllGVRs() {
		meta, err := dao.MetaAccess.MetaFor(gvr)
		if err != nil || dao.IsK9sMeta(meta) || !meta.Namespaced {
			continue
		}
		if !client.Can(meta.Verbs, "view") {
			continue
		}
		gvrs = append(gvrs, gvr)
	}
	slog.Debug("OwnerTree: GVR scan", "duration", time.Since(gvrScanStart), "gvrCount", len(gvrs))

	// Read from the informer lister cache — zero network calls.
	listerStart := time.Now()
	index := make(map[types.UID][]*unstructured.Unstructured)
	totalObjs := 0
	skipped := 0
	for _, gvr := range gvrs {
		inf, err := factory.ForResource(ns, gvr)
		if err != nil || inf == nil {
			slog.Debug("OwnerTree: no informer", slogs.GVR, gvr, slogs.Error, err)
			skipped++
			continue
		}
		if !inf.Informer().HasSynced() {
			slog.Debug("OwnerTree: informer not yet synced, skipping", slogs.GVR, gvr)
			skipped++
			continue
		}
		var objs []runtime.Object
		if client.IsClusterScoped(ns) {
			objs, err = inf.Lister().List(labels.Everything())
		} else {
			objs, err = inf.Lister().ByNamespace(ns).List(labels.Everything())
		}
		if err != nil {
			slog.Debug("OwnerTree: lister error", slogs.GVR, gvr, slogs.Error, err)
			continue
		}
		totalObjs += len(objs)
		for _, o := range objs {
			u, ok := o.(*unstructured.Unstructured)
			if !ok {
				continue
			}
			for _, ref := range u.GetOwnerReferences() {
				index[ref.UID] = append(index[ref.UID], u)
			}
		}
	}
	slog.Debug("OwnerTree: index built from cache",
		"duration", time.Since(listerStart),
		"gvrCount", len(gvrs),
		"skipped", skipped,
		"objects", totalObjs,
		"entries", len(index),
	)
	slog.Debug("OwnerTree: buildOwnerIndex total", slogs.Duration, time.Since(totalStart))
	return index
}

// buildOwnerChildren recursively adds children to a tree node using the
// ownerIndex reverse map. visited tracks UIDs already on the current path
// to prevent infinite recursion on circular OwnerReferences.
func buildOwnerChildren(parent *xray.TreeNode, uid types.UID, ownerIndex map[types.UID][]*unstructured.Unstructured, visited map[types.UID]struct{}) {
	for _, child := range ownerIndex[uid] {
		if _, seen := visited[child.GetUID()]; seen {
			slog.Warn("OwnerTree: cycle detected, skipping", slogs.UID, child.GetUID())
			continue
		}
		childGVR := gvrFromUnstructured(child)
		childNode := xray.NewTreeNode(childGVR, resourceFQN(child))
		childNode.Extras[xray.KindKey] = child.GetKind()
		setOwnerNodeStatus(childNode, child)
		visited[child.GetUID()] = struct{}{}
		buildOwnerChildren(childNode, child.GetUID(), ownerIndex, visited)
		delete(visited, child.GetUID())
		parent.Add(childNode)
	}
}

// gvrFromUnstructured derives the *client.GVR for a given unstructured object.
// It first tries MetaAccess (for registered resources), and falls back to
// constructing a best-effort GVR from the object's APIVersion and Kind.
func gvrFromUnstructured(u *unstructured.Unstructured) *client.GVR {
	apiVersion := u.GetAPIVersion()
	kind := u.GetKind()

	gv, err := schema.ParseGroupVersion(apiVersion)
	if err == nil {
		if gvr, _, found := dao.MetaAccess.GVK2GVR(gv, kind); found {
			return gvr
		}
	}

	// Fallback: build a best-effort GVR string: group/version/resources.
	resource := strings.ToLower(kind) + "s"
	if err != nil {
		return client.NewGVR(resource)
	}
	if gv.Group == "" {
		return client.NewGVR(gv.Version + "/" + resource)
	}
	return client.NewGVR(gv.Group + "/" + gv.Version + "/" + resource)
}

// resourceFQN returns a FQN (namespace/name or just name for cluster-scoped)
// for an unstructured object.
func resourceFQN(u *unstructured.Unstructured) string {
	ns := u.GetNamespace()
	if ns == "" {
		return u.GetName()
	}
	return client.FQN(ns, u.GetName())
}

// setOwnerNodeStatus inspects the status conditions of an unstructured
// resource and sets the tree node's StatusKey and InfoKey accordingly.
func setOwnerNodeStatus(node *xray.TreeNode, u *unstructured.Unstructured) {
	conditions, found, _ := unstructured.NestedSlice(u.Object, "status", "conditions")
	if !found || len(conditions) == 0 {
		// No conditions — try a simple phase field.
		if phase, ok, _ := unstructured.NestedString(u.Object, "status", "phase"); ok && phase != "" {
			node.Extras[xray.InfoKey] = phase
			switch strings.ToLower(phase) {
			case "failed", "error", "terminating":
				node.Extras[xray.StatusKey] = xray.ToastStatus
			}
		}
		return
	}

	for _, c := range conditions {
		cond, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		status, _, _ := unstructured.NestedString(cond, "status")
		reason, _, _ := unstructured.NestedString(cond, "reason")

		if status != "True" {
			node.Extras[xray.StatusKey] = xray.ToastStatus
			if reason != "" {
				node.Extras[xray.InfoKey] = reason
			}
			return
		} else {
			node.Extras[xray.StatusKey] = xray.CompletedStatus
		}
			
		if reason != "" {
			node.Extras[xray.InfoKey] = reason
		}
	}
}

func (t *OwnerTree) fireTreeChanged(root *xray.TreeNode) {
	for _, l := range t.listeners {
		l.TreeChanged(root)
	}
}

func (t *OwnerTree) fireTreeLoadFailed(err error) {
	for _, l := range t.listeners {
		l.TreeLoadFailed(err)
	}
}
