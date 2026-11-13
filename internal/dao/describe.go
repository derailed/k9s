// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"log/slog"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/slogs"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/kubectl/pkg/describe"
)

// Describe describes a resource.
func Describe(c client.Connection, gvr *client.GVR, path string) (string, error) {
	mapper := RestMapper{Connection: c}
	m, err := mapper.ToRESTMapper()
	if err != nil {
		slog.Error("No REST mapper for resource",
			slogs.GVR, gvr,
			slogs.Error, err,
		)
		return "", err
	}

	gvk, err := m.KindFor(gvr.GVR())
	if err != nil {
		slog.Error("No GVK for resource %s",
			slogs.GVR, gvr,
			slogs.Error, err,
		)
		return "", err
	}

	ns, n := client.Namespaced(path)
	if client.IsClusterScoped(ns) {
		ns = client.BlankNamespace
	}
	mapping, err := mapper.ResourceFor(gvr.AsResourceName(), gvk.Kind)
	if err != nil {
		slog.Error("Unable to find mapper",
			slogs.GVR, gvr,
			slogs.ResName, n,
			slogs.Error, err,
		)
		return "", err
	}
	d, err := describe.Describer(c.Config().Flags(), mapping)
	if err != nil {
		slog.Error("Unable to find describer",
			slogs.GVR, gvr.AsResourceName(),
			slogs.Error, err,
		)
		return "", err
	}

	text, err := d.Describe(ns, n, describe.DescriberSettings{ShowEvents: false})
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), c.Config().CallTimeout())
	defer cancel()
	events, err := describeResourceEvents(ctx, c, gvr, ns, n, gvk.Kind)
	if err != nil {
		// Keep the resource description available when events cannot be read.
		slog.Warn("Unable to describe events", slogs.Error, err)
		return text + "Events: <unavailable>\n", nil
	}
	return text + events, nil
}

func describeResourceEvents(ctx context.Context, c client.Connection, gvr *client.GVR, ns, name, kind string) (string, error) {
	dynamic, err := c.DynDial()
	if err != nil {
		return "", err
	}
	obj, err := dynamic.Resource(gvr.GVR()).Namespace(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	kube, err := c.Dial()
	if err != nil {
		return "", err
	}
	uid := string(obj.GetUID())
	events := kube.CoreV1().Events(ns)
	selector := fields.Set{
		"involvedObject.name":      name,
		"involvedObject.namespace": ns,
		"involvedObject.kind":      kind,
		"involvedObject.uid":       uid,
	}.AsSelector()
	opts := metav1.ListOptions{FieldSelector: selector.String(), Limit: 500}
	var items []v1.Event
	for {
		list, err := events.List(ctx, opts)
		if err != nil {
			return "", err
		}
		items = append(items, list.Items...)
		if list.Continue == "" {
			break
		}
		opts.Continue = list.Continue
	}
	return formatDescribeEvents(items), nil
}
