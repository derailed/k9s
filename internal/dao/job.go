// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"errors"
	"fmt"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/rs/zerolog/log"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	_ Accessor    = (*Job)(nil)
	_ Nuker       = (*Job)(nil)
	_ Loggable    = (*Job)(nil)
	_ ImageLister = (*Deployment)(nil)
)

// Job represents a K8s job resource.
type Job struct {
	Resource
}

// ListImages lists container images.
func (j *Job) ListImages(ctx context.Context, fqn string) ([]string, error) {
	job, err := j.GetInstance(fqn)
	if err != nil {
		return nil, err
	}

	return render.ExtractImages(&job.Spec.Template.Spec), nil
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
func (j *Job) TailLogs(ctx context.Context, opts *LogOptions) ([]LogChan, error) {
	o, err := j.getFactory().Get(j.gvrStr(), opts.Path, true, labels.Everything())
	if err != nil {
		return nil, err
	}

	var job batchv1.Job
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &job)
	if err != nil {
		return nil, errors.New("expecting a job resource")
	}

	if job.Spec.Selector == nil || len(job.Spec.Selector.MatchLabels) == 0 {
		return nil, fmt.Errorf("no valid selector found for job: %s", opts.Path)
	}

	return podLogs(ctx, job.Spec.Selector.MatchLabels, opts)
}

func (j *Job) GetInstance(fqn string) (*batchv1.Job, error) {
	o, err := j.getFactory().Get(j.gvrStr(), fqn, true, labels.Everything())
	if err != nil {
		return nil, err
	}

	var job batchv1.Job
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &job)
	if err != nil {
		return nil, errors.New("expecting a job resource")
	}

	return &job, nil
}

// ScanSA scans for serviceaccount refs.
func (j *Job) ScanSA(ctx context.Context, fqn string, wait bool) (Refs, error) {
	ns, n := client.Namespaced(fqn)
	oo, err := j.getFactory().List(j.GVR(), ns, wait, labels.Everything())
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
		if serviceAccountMatches(job.Spec.Template.Spec.ServiceAccountName, n) {
			refs = append(refs, Ref{
				GVR: j.GVR(),
				FQN: client.FQN(job.Namespace, job.Name),
			})
		}
	}

	return refs, nil
}

// Scan scans for resource references.
func (j *Job) Scan(ctx context.Context, gvr client.GVR, fqn string, wait bool) (Refs, error) {
	ns, n := client.Namespaced(fqn)
	oo, err := j.getFactory().List(j.GVR(), ns, wait, labels.Everything())
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
		case CmGVR:
			if !hasConfigMap(&job.Spec.Template.Spec, n) {
				continue
			}
			refs = append(refs, Ref{
				GVR: j.GVR(),
				FQN: client.FQN(job.Namespace, job.Name),
			})
		case SecGVR:
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
		case PcGVR:
			if !hasPC(&job.Spec.Template.Spec, n) {
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
