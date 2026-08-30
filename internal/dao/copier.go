// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/slogs"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/rand"
)

// CopyOpts tracks resource copy options.
type CopyOpts struct {
	// Namespace tracks the target namespace.
	Namespace string

	// Name tracks the target resource name. Defaults to the source name when blank.
	Name string

	// Overwrite updates the target resource when it already exists.
	Overwrite bool
}

// Copier is for resources that can be duplicated into another namespace.
type Copier interface {
	// Copy duplicates a resource into a target namespace and returns the new fqn.
	Copy(ctx context.Context, path string, opts *CopyOpts) (string, error)

	// CopyPrep returns a sanitized copy of a resource, ready to be created.
	CopyPrep(ctx context.Context, path string, opts *CopyOpts) (*unstructured.Unstructured, error)

	// CopyCreate persists a prepared resource copy and returns the new fqn.
	CopyCreate(ctx context.Context, u *unstructured.Unstructured, overwrite bool) (string, error)
}

const (
	// copySuffixLen tracks the number of random chars tacked onto a duplicate name.
	copySuffixLen = 5

	// maxNameLen tracks the max length of a generated name. Resource names are
	// routinely reused as label values which cap out at 63 chars.
	maxNameLen = 63
)

var _ Copier = (*Generic)(nil)

// serverOwnedFields tracks fields the api-server owns and that must be dropped
// prior to recreating a resource elsewhere.
var serverOwnedFields = [][]string{
	{"metadata", "creationTimestamp"},
	{"metadata", "deletionGracePeriodSeconds"},
	{"metadata", "deletionTimestamp"},
	{"metadata", "finalizers"},
	{"metadata", "generateName"},
	{"metadata", "generation"},
	{"metadata", "managedFields"},
	{"metadata", "ownerReferences"},
	{"metadata", "resourceVersion"},
	{"metadata", "selfLink"},
	{"metadata", "uid"},
	{"status"},
}

// staleAnnotations tracks annotations that no longer apply to a copy.
var staleAnnotations = []string{
	"kubectl.kubernetes.io/last-applied-configuration",
	"deployment.kubernetes.io/revision",
}

// boundFields tracks per resource fields that are bound to the source namespace
// or cluster and thus cannot be carried over to a copy.
var boundFields = map[*client.GVR][][]string{
	client.SvcGVR: {
		{"spec", "clusterIP"},
		{"spec", "clusterIPs"},
		{"spec", "healthCheckNodePort"},
		{"spec", "ipFamilies"},
		{"spec", "ipFamilyPolicy"},
	},
	client.PodGVR: {
		{"spec", "nodeName"},
	},
	client.PvcGVR: {
		{"spec", "volumeName"},
	},
	client.SaGVR: {
		{"secrets"},
	},
	client.JobGVR: {
		{"spec", "selector"},
		{"spec", "manualSelector"},
	},
}

// jobTrackingLabels tracks labels the job controller stamps from the source uid.
var jobTrackingLabels = []string{
	"controller-uid",
	"job-name",
	"batch.kubernetes.io/controller-uid",
	"batch.kubernetes.io/job-name",
}

// Copy duplicates a resource into a target namespace.
func (g *Generic) Copy(ctx context.Context, path string, opts *CopyOpts) (string, error) {
	u, err := g.CopyPrep(ctx, path, opts)
	if err != nil {
		return "", err
	}

	return g.CopyCreate(ctx, u, opts.Overwrite)
}

// CopyPrep fetches a resource and strips out everything binding it to its
// current namespace so it can be recreated elsewhere.
func (g *Generic) CopyPrep(ctx context.Context, path string, opts *CopyOpts) (*unstructured.Unstructured, error) {
	if opts == nil || opts.Namespace == "" {
		return nil, errors.New("no target namespace specified")
	}

	ns, n := client.Namespaced(path)
	if client.IsClusterScoped(ns) {
		return nil, fmt.Errorf("%s is cluster scoped and cannot be copied to a namespace", g.gvr)
	}

	target := opts.Name
	if target == "" {
		target = n
	}
	// Copying a resource onto itself only makes sense as a duplicate, so mint a
	// unique name unless the user explicitly asked to overwrite.
	if ns == opts.Namespace && target == n {
		if opts.Overwrite {
			return nil, fmt.Errorf("source and target are identical: %s", client.FQN(ns, n))
		}
		target = suffixName(n)
	}

	if err := g.canCopy(opts.Namespace, target, opts.Overwrite); err != nil {
		return nil, err
	}

	o, err := g.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	u, ok := o.DeepCopyObject().(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("expecting unstructured but got %T", o)
	}
	sanitizeForCopy(u, g.gvr)
	u.SetNamespace(opts.Namespace)
	u.SetName(target)

	return u, nil
}

// CopyCreate persists a prepared copy. The object is sanitized once more since
// it may have been hand edited on the way in.
func (g *Generic) CopyCreate(ctx context.Context, u *unstructured.Unstructured, overwrite bool) (string, error) {
	if u == nil {
		return "", errors.New("no resource to copy")
	}

	u = u.DeepCopy()
	sanitizeForCopy(u, g.gvr)
	ns, n := u.GetNamespace(), u.GetName()
	if ns == "" {
		return "", errors.New("no target namespace specified")
	}
	if n == "" {
		return "", errors.New("no target name specified")
	}
	if err := g.canCopy(ns, n, overwrite); err != nil {
		return "", err
	}

	dial, err := g.dynClient()
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(ctx, g.Client().Config().CallTimeout())
	defer cancel()

	fqn := client.FQN(ns, n)
	_, err = dial.Namespace(ns).Create(ctx, u, metav1.CreateOptions{})
	if err == nil {
		return fqn, nil
	}
	if !apierrors.IsAlreadyExists(err) {
		return "", err
	}

	if !overwrite {
		return "", fmt.Errorf("%s already exists in namespace %q. Enable overwrite to replace it", n, ns)
	}
	current, err := dial.Namespace(ns).Get(ctx, n, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	u.SetResourceVersion(current.GetResourceVersion())
	if _, err := dial.Namespace(ns).Update(ctx, u, metav1.UpdateOptions{}); err != nil {
		return "", err
	}

	return fqn, nil
}

// canCopy asserts the user is allowed to write the resource in the target ns.
func (g *Generic) canCopy(ns, n string, overwrite bool) error {
	verbs := []string{client.CreateVerb}
	if overwrite {
		verbs = append(verbs, client.UpdateVerb)
	}
	auth, err := g.Client().CanI(ns, g.gvr, n, verbs)
	if err != nil {
		return err
	}
	if !auth {
		return fmt.Errorf("user is not authorized to create %s in namespace %q", g.gvr, ns)
	}

	return nil
}

// suffixName mints a unique name for a copy living alongside its source.
func suffixName(n string) string {
	if l := maxNameLen - copySuffixLen - 1; len(n) > l {
		n = n[:l]
	}

	return n + "-" + rand.String(copySuffixLen)
}

// sanitizeForCopy strips out all the bits that tie a resource to its current
// namespace so it can be recreated elsewhere.
func sanitizeForCopy(u *unstructured.Unstructured, gvr *client.GVR) {
	for _, f := range serverOwnedFields {
		unstructured.RemoveNestedField(u.Object, f...)
	}

	if aa := u.GetAnnotations(); len(aa) > 0 {
		for _, a := range staleAnnotations {
			delete(aa, a)
		}
		if len(aa) == 0 {
			unstructured.RemoveNestedField(u.Object, "metadata", "annotations")
		} else {
			u.SetAnnotations(aa)
		}
	}

	for _, f := range boundFields[gvr] {
		unstructured.RemoveNestedField(u.Object, f...)
	}

	switch gvr {
	case client.SvcGVR:
		sanitizeSvcPorts(u)
	case client.JobGVR:
		sanitizeJobLabels(u)
	}
}

// sanitizeSvcPorts drops node ports so the api-server allocates fresh ones.
func sanitizeSvcPorts(u *unstructured.Unstructured) {
	pp, ok, err := unstructured.NestedSlice(u.Object, "spec", "ports")
	if !ok || err != nil {
		return
	}
	for i := range pp {
		p, ok := pp[i].(map[string]any)
		if !ok {
			continue
		}
		delete(p, "nodePort")
	}
	if err := unstructured.SetNestedSlice(u.Object, pp, "spec", "ports"); err != nil {
		slog.Warn("Unable to reset service node ports", slogs.Error, err)
	}
}

// sanitizeJobLabels drops the tracking labels stamped from the source job uid.
func sanitizeJobLabels(u *unstructured.Unstructured) {
	for _, f := range [][]string{
		{"metadata", "labels"},
		{"spec", "template", "metadata", "labels"},
	} {
		ll, ok, err := unstructured.NestedStringMap(u.Object, f...)
		if !ok || err != nil {
			continue
		}
		for _, l := range jobTrackingLabels {
			delete(ll, l)
		}
		if len(ll) == 0 {
			unstructured.RemoveNestedField(u.Object, f...)
			continue
		}
		if err := unstructured.SetNestedStringMap(u.Object, ll, f...); err != nil {
			slog.Warn("Unable to reset job tracking labels", slogs.Error, err)
		}
	}
}
