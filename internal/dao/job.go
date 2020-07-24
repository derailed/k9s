package dao

import (
	"context"
	"errors"
	"fmt"
	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/rs/zerolog/log"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	_ Accessor = (*Job)(nil)
	_ Nuker    = (*Job)(nil)
	_ Loggable = (*Job)(nil)
)

// Job represents a K8s job resource.
type Job struct {
	Resource
}

// List returns a collection of resources.
func (j *Job) List(ctx context.Context, ns string) ([]runtime.Object, error) {
	oo, err := j.Resource.List(ctx, ns)
	if err != nil {
		return nil, err
	}
	ctrl, _ := ctx.Value(internal.KeyPath).(string)
	_, n := client.Namespaced(ctrl)

	ll := make([]runtime.Object, 0, 10)
	for _, o := range oo {
		var j batchv1.Job
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &j)
		if err != nil {
			return nil, errors.New("expecting Job resource")
		}
		if n == "" {
			ll = append(ll, o)
			continue
		}

			for _, r := range j.ObjectMeta.OwnerReferences {
				if r.Name == n {
					ll = append(ll, o)
				}
			}
	}

	return ll, nil
}

// TailLogs tail logs for all pods represented by this Job.
func (j *Job) TailLogs(ctx context.Context, c LogChan, opts LogOptions) error {
	o, err := j.Factory.Get(j.gvr.String(), opts.Path, true, labels.Everything())
	if err != nil {
		return err
	}

	var job batchv1.Job
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &job)
	if err != nil {
		return errors.New("expecting a job resource")
	}

	if job.Spec.Selector == nil || len(job.Spec.Selector.MatchLabels) == 0 {
		return fmt.Errorf("No valid selector found on Job %s", opts.Path)
	}

	return podLogs(ctx, c, job.Spec.Selector.MatchLabels, opts)
}

// ScanSA scans for serviceaccount refs.
func (j *Job) ScanSA(ctx context.Context, fqn string, wait bool) (Refs, error) {
	ns, n := client.Namespaced(fqn)
	oo, err := j.Factory.List(j.GVR(), ns, wait, labels.Everything())
	if err != nil {
		return nil, err
	}

	refs := make(Refs, 0, len(oo))
	for _, o := range oo {
		var job batchv1.Job
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &job)
		if err != nil {
			return nil, errors.New("expecting Job resource")
		}
		if job.Spec.Template.Spec.ServiceAccountName == n {
			refs = append(refs, Ref{
				GVR: j.GVR(),
				FQN: client.FQN(job.Namespace, job.Name),
			})
		}
	}

	return refs, nil
}

// Scan scans for resource references.
func (j *Job) Scan(ctx context.Context, gvr, fqn string, wait bool) (Refs, error) {
	ns, n := client.Namespaced(fqn)
	oo, err := j.Factory.List(j.GVR(), ns, wait, labels.Everything())
	if err != nil {
		return nil, err
	}

	refs := make(Refs, 0, len(oo))
	for _, o := range oo {
		var job batchv1.Job
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &job)
		if err != nil {
			return nil, errors.New("expecting Job resource")
		}
		switch gvr {
		case "v1/configmaps":
			if !hasConfigMap(&job.Spec.Template.Spec, n) {
				continue
			}
			refs = append(refs, Ref{
				GVR: j.GVR(),
				FQN: client.FQN(job.Namespace, job.Name),
			})
		case "v1/secrets":
			found, err := hasSecret(j.Factory, &job.Spec.Template.Spec, job.Namespace, n, wait)
			if err != nil {
				log.Warn().Err(err).Msgf("locate secret %q", fqn)
				continue
			}
			if !found {
				continue
			}
			refs = append(refs, Ref{
				GVR: j.GVR(),
				FQN: client.FQN(job.Namespace, job.Name),
			})
		}
	}

	return refs, nil
}
