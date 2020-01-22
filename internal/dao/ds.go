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
)

// DaemonSet represents a K8s daemonset.
type DaemonSet struct {
	Resource
}

// Restart a DaemonSet rollout.
func (d *DaemonSet) Restart(path string) error {
	o, err := d.Factory.Get(d.gvr.String(), path, true, labels.Everything())
	if err != nil {
		return err
	}
	var ds appsv1.DaemonSet
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &ds)
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
	update, err := polymorphichelpers.ObjectRestarterFn(&ds)
	if err != nil {
		return err
	}
	_, err = d.Client().DialOrDie().AppsV1().DaemonSets(ds.Namespace).Patch(ds.Name, types.StrategicMergePatchType, update)

	return err
}

// TailLogs tail logs for all pods represented by this DaemonSet.
func (d *DaemonSet) TailLogs(ctx context.Context, c chan<- string, opts LogOptions) error {
	o, err := d.Factory.Get(d.gvr.String(), opts.Path, true, labels.Everything())
	if err != nil {
		return err
	}
	var ds appsv1.DaemonSet
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &ds)
	if err != nil {
		return errors.New("expecting daemonset resource")
	}

	if ds.Spec.Selector == nil || len(ds.Spec.Selector.MatchLabels) == 0 {
		return fmt.Errorf("no valid selector found on daemonset %q", opts.Path)
	}

	return podLogs(ctx, c, ds.Spec.Selector.MatchLabels, opts)
}

func podLogs(ctx context.Context, c chan<- string, sel map[string]string, opts LogOptions) error {
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

	if len(oo) > 1 {
		opts.MultiPods = true
	}

	po := Pod{}
	po.Init(f, client.NewGVR("v1/pods"))
	for _, o := range oo {
		var pod v1.Pod
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &pod)
		if err != nil {
			return err
		}
		log.Debug().Msgf("TAILING logs on pod %q", pod.Name)
		opts.Path = client.FQN(pod.Namespace, pod.Name)
		if err := po.TailLogs(ctx, c, opts); err != nil {
			return err
		}
	}
	return nil
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
