package dao

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/watch"
	"github.com/rs/zerolog/log"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubectl/pkg/polymorphichelpers"
)

var (
	_ Accessor    = (*DaemonSet)(nil)
	_ Nuker       = (*DaemonSet)(nil)
	_ Loggable    = (*DaemonSet)(nil)
	_ Restartable = (*DaemonSet)(nil)
	_ Controller  = (*DaemonSet)(nil)
)

// DaemonSet represents a K8s daemonset.
type DaemonSet struct {
	Resource
}

// IsHappy check for happy deployments.
func (d *DaemonSet) IsHappy(ds appsv1.DaemonSet) bool {
	return ds.Status.DesiredNumberScheduled == ds.Status.CurrentNumberScheduled
}

// Restart a DaemonSet rollout.
func (d *DaemonSet) Restart(ctx context.Context, path string) error {
	ds, err := d.GetInstance(path)
	if err != nil {
		return err
	}

	auth, err := d.Client().CanI(ds.Namespace, "apps/v1/daemonsets", []string{client.PatchVerb})
	if err != nil {
		return err
	}
	if !auth {
		return fmt.Errorf("user is not authorized to restart a daemonset")
	}
	update, err := polymorphichelpers.ObjectRestarterFn(ds)
	if err != nil {
		return err
	}

	dial, err := d.Client().Dial()
	if err != nil {
		return err
	}
	_, err = dial.AppsV1().DaemonSets(ds.Namespace).Patch(
		ctx,
		ds.Name,
		types.StrategicMergePatchType,
		update,
		metav1.PatchOptions{},
	)
	return err
}

// TailLogs tail logs for all pods represented by this DaemonSet.
func (d *DaemonSet) TailLogs(ctx context.Context, c LogChan, opts LogOptions) error {
	ds, err := d.GetInstance(opts.Path)
	if err != nil {
		return err
	}

	if ds.Spec.Selector == nil || len(ds.Spec.Selector.MatchLabels) == 0 {
		return fmt.Errorf("no valid selector found on daemonset %q", opts.Path)
	}

	return podLogs(ctx, c, ds.Spec.Selector.MatchLabels, opts)
}

func podLogs(ctx context.Context, c LogChan, sel map[string]string, opts LogOptions) error {
	f, ok := ctx.Value(internal.KeyFactory).(*watch.Factory)
	if !ok {
		return errors.New("expecting a context factory")
	}
	ls, err := metav1.ParseToLabelSelector(toSelector(sel))
	if err != nil {
		return err
	}
	lsel, err := metav1.LabelSelectorAsSelector(ls)
	if err != nil {
		return err
	}

	ns, _ := client.Namespaced(opts.Path)
	oo, err := f.List("v1/pods", ns, true, lsel)
	if err != nil {
		return err
	}
	opts.MultiPods = true

	po := Pod{}
	po.Init(f, client.NewGVR("v1/pods"))
	for _, o := range oo {
		var pod v1.Pod
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &pod)
		if err != nil {
			return err
		}
		opts.Path = client.FQN(pod.Namespace, pod.Name)
		if err := po.TailLogs(ctx, c, opts); err != nil {
			return err
		}
	}
	return nil
}

// Pod returns a pod victim by name.
func (d *DaemonSet) Pod(fqn string) (string, error) {
	ds, err := d.GetInstance(fqn)
	if err != nil {
		return "", err
	}

	return podFromSelector(d.Factory, ds.Namespace, ds.Spec.Selector.MatchLabels)
}

// GetInstance returns a daemonset instance.
func (d *DaemonSet) GetInstance(fqn string) (*appsv1.DaemonSet, error) {
	o, err := d.Factory.Get(d.gvr.String(), fqn, true, labels.Everything())
	if err != nil {
		return nil, err
	}

	var ds appsv1.DaemonSet
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &ds)
	if err != nil {
		return nil, errors.New("expecting DaemonSet resource")
	}

	return &ds, nil
}

// ScanSA scans for serviceaccount refs.
func (d *DaemonSet) ScanSA(ctx context.Context, fqn string, wait bool) (Refs, error) {
	ns, n := client.Namespaced(fqn)
	oo, err := d.Factory.List(d.GVR(), ns, wait, labels.Everything())
	if err != nil {
		return nil, err
	}

	refs := make(Refs, 0, len(oo))
	for _, o := range oo {
		var ds appsv1.DaemonSet
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &ds)
		if err != nil {
			return nil, errors.New("expecting DaemonSet resource")
		}
		if ds.Spec.Template.Spec.ServiceAccountName == n {
			refs = append(refs, Ref{
				GVR: d.GVR(),
				FQN: client.FQN(ds.Namespace, ds.Name),
			})
		}
	}

	return refs, nil
}

// Scan scans for cluster refs.
func (d *DaemonSet) Scan(ctx context.Context, gvr, fqn string, wait bool) (Refs, error) {
	ns, n := client.Namespaced(fqn)
	oo, err := d.Factory.List(d.GVR(), ns, wait, labels.Everything())
	if err != nil {
		return nil, err
	}

	refs := make(Refs, 0, len(oo))
	for _, o := range oo {
		var ds appsv1.DaemonSet
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &ds)
		if err != nil {
			return nil, errors.New("expecting StatefulSet resource")
		}
		switch gvr {
		case "v1/configmaps":
			if !hasConfigMap(&ds.Spec.Template.Spec, n) {
				continue
			}
			refs = append(refs, Ref{
				GVR: d.GVR(),
				FQN: client.FQN(ds.Namespace, ds.Name),
			})
		case "v1/secrets":
			found, err := hasSecret(d.Factory, &ds.Spec.Template.Spec, ds.Namespace, n, wait)
			if err != nil {
				log.Warn().Err(err).Msgf("locate secret %q", fqn)
				continue
			}
			if !found {
				continue
			}
			refs = append(refs, Ref{
				GVR: d.GVR(),
				FQN: client.FQN(ds.Namespace, ds.Name),
			})
		case "v1/persistentvolumeclaims":
			if !hasPVC(&ds.Spec.Template.Spec, n) {
				continue
			}
			refs = append(refs, Ref{
				GVR: d.GVR(),
				FQN: client.FQN(ds.Namespace, ds.Name),
			})
		}
	}

	return refs, nil
}

// ----------------------------------------------------------------------------
// Helpers...

func toSelector(m map[string]string) string {
	s := make([]string, 0, len(m))
	for k, v := range m {
		s = append(s, k+"="+v)
	}

	return strings.Join(s, ",")
}
