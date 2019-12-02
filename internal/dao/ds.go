package dao

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal"
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

type DaemonSet struct {
	Resource
}

var _ Accessor = &DaemonSet{}
var _ Loggable = &DaemonSet{}
var _ Restartable = &DaemonSet{}

// Restart a DaemonSet rollout.
func (d *DaemonSet) Restart(ns, n string) error {
	o, err := d.Get(ns, string(d.gvr), n, labels.Everything())
	if err != nil {
		return err
	}

	var ds appsv1.DaemonSet
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &ds)
	if err != nil {
		return err
	}

	update, err := polymorphichelpers.ObjectRestarterFn(&ds)
	if err != nil {
		return err
	}

	_, err = d.Client().DialOrDie().AppsV1().DaemonSets(ns).Patch(ds.Name, types.StrategicMergePatchType, update)
	return err
}

// Logs tail logs for all pods represented by this DaemonSet.
func (d *DaemonSet) TailLogs(ctx context.Context, c chan<- string, opts LogOptions) error {
	log.Debug().Msgf("Tailing DaemonSet %q -- %q", opts.Namespace, opts.Name)
	o, err := d.Get(opts.Namespace, "apps/v1/daemonsets", opts.Name, labels.Everything())
	if err != nil {
		return err
	}

	var ds appsv1.DaemonSet
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &ds)
	if err != nil {
		return errors.New("expecting daemonset resource")
	}

	if ds.Spec.Selector == nil || len(ds.Spec.Selector.MatchLabels) == 0 {
		return fmt.Errorf("No valid selector found on daemonset %s", opts.FQN())
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

	oo, err := f.List(opts.Namespace, "v1/pods", lsel)
	if err != nil {
		return err
	}

	if len(oo) > 1 {
		opts.MultiPods = true
	}

	po := Pod{}
	for _, o := range oo {
		var pod v1.Pod
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &pod)
		if err != nil {
			return err
		}
		if pod.Status.Phase == v1.PodRunning {
			opts.Namespace, opts.Name = pod.Namespace, pod.Name
			if err := po.TailLogs(ctx, c, opts); err != nil {
				return err
			}
		}
	}
	return nil
}

// Helpers...

func toSelector(m map[string]string) string {
	s := make([]string, 0, len(m))
	for k, v := range m {
		s = append(s, k+"="+v)
	}

	return strings.Join(s, ",")
}
