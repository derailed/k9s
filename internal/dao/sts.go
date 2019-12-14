package dao

import (
	"context"
	"errors"
	"fmt"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubectl/pkg/polymorphichelpers"
)

type StatefulSet struct {
	Generic
}

var _ Accessor = &StatefulSet{}
var _ Loggable = &StatefulSet{}
var _ Restartable = &StatefulSet{}
var _ Scalable = &StatefulSet{}

// Scale a StatefulSet.
func (s *StatefulSet) Scale(path string, replicas int32) error {
	ns, n := k8s.Namespaced(path)
	scale, err := s.Client().DialOrDie().AppsV1().StatefulSets(ns).GetScale(n, metav1.GetOptions{})
	if err != nil {
		return err
	}
	scale.Spec.Replicas = replicas
	_, err = s.Client().DialOrDie().AppsV1().StatefulSets(ns).UpdateScale(n, scale)

	return err
}

// Restart a StatefulSet rollout.
func (s *StatefulSet) Restart(path string) error {
	o, err := s.Get(string(s.gvr), path, labels.Everything())
	if err != nil {
		return err
	}

	var ds appsv1.StatefulSet
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &ds)
	if err != nil {
		return err
	}

	update, err := polymorphichelpers.ObjectRestarterFn(&ds)
	if err != nil {
		return err
	}

	_, err = s.Client().DialOrDie().AppsV1().StatefulSets(ds.Namespace).Patch(ds.Name, types.StrategicMergePatchType, update)
	return err
}

// Logs tail logs for all pods represented by this StatefulSet.
func (s *StatefulSet) TailLogs(ctx context.Context, c chan<- string, opts LogOptions) error {
	log.Debug().Msgf("Tailing StatefulSet %q", opts.Path)
	o, err := s.Get(string(s.gvr), opts.Path, labels.Everything())
	if err != nil {
		return err
	}

	var dp appsv1.StatefulSet
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &dp)
	if err != nil {
		return errors.New("expecting StatefulSet resource")
	}

	if dp.Spec.Selector == nil || len(dp.Spec.Selector.MatchLabels) == 0 {
		return fmt.Errorf("No valid selector found on StatefulSet %s", opts.Path)
	}

	return podLogs(ctx, c, dp.Spec.Selector.MatchLabels, opts)
}
