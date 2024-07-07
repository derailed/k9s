// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"errors"
	"fmt"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/rs/zerolog/log"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
)

const (
	maxJobNameSize = 42
	jobGVR         = "batch/v1/jobs"
)

var (
	_ Accessor    = (*CronJob)(nil)
	_ Runnable    = (*CronJob)(nil)
	_ ImageLister = (*CronJob)(nil)
)

// CronJob represents a cronjob K8s resource.
type CronJob struct {
	Generic
}

// ListImages lists container images.
func (c *CronJob) ListImages(ctx context.Context, fqn string) ([]string, error) {
	cj, err := c.GetInstance(fqn)
	if err != nil {
		return nil, err
	}

	return render.ExtractImages(&cj.Spec.JobTemplate.Spec.Template.Spec), nil
}

// Run a CronJob.
func (c *CronJob) Run(path string) error {
	ns, n := client.Namespaced(path)
	auth, err := c.Client().CanI(ns, jobGVR, n, []string{client.GetVerb, client.CreateVerb})
	if err != nil {
		return err
	}
	if !auth {
		return fmt.Errorf("user is not authorized to run jobs")
	}

	o, err := c.getFactory().Get(c.GVR(), path, true, labels.Everything())
	if err != nil {
		return err
	}
	var cj batchv1.CronJob
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &cj)
	if err != nil {
		return errors.New("expecting CronJob resource")
	}
	jobName := cj.Name
	if len(cj.Name) >= maxJobNameSize {
		jobName = cj.Name[0:maxJobNameSize]
	}
	true := true
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:        jobName + "-manual-" + rand.String(3),
			Namespace:   ns,
			Labels:      cj.Spec.JobTemplate.Labels,
			Annotations: cj.Spec.JobTemplate.Annotations,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         c.gvr.GV().String(),
					Kind:               "CronJob",
					BlockOwnerDeletion: &true,
					Controller:         &true,
					Name:               cj.Name,
					UID:                cj.UID,
				},
			},
		},
		Spec: cj.Spec.JobTemplate.Spec,
	}
	dial, err := c.Client().Dial()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), c.Client().Config().CallTimeout())
	defer cancel()
	_, err = dial.BatchV1().Jobs(ns).Create(ctx, job, metav1.CreateOptions{})

	return err
}

// ScanSA scans for serviceaccount refs.
func (c *CronJob) ScanSA(ctx context.Context, fqn string, wait bool) (Refs, error) {
	ns, n := client.Namespaced(fqn)
	oo, err := c.getFactory().List(c.GVR(), ns, wait, labels.Everything())
	if err != nil {
		return nil, err
	}

	refs := make(Refs, 0, len(oo))
	for _, o := range oo {
		var cj batchv1.CronJob
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &cj)
		if err != nil {
			return nil, errors.New("expecting CronJob resource")
		}
		if serviceAccountMatches(cj.Spec.JobTemplate.Spec.Template.Spec.ServiceAccountName, n) {
			refs = append(refs, Ref{
				GVR: c.GVR(),
				FQN: client.FQN(cj.Namespace, cj.Name),
			})
		}
	}

	return refs, nil
}

// GetInstance fetch a matching cronjob.
func (c *CronJob) GetInstance(fqn string) (*batchv1.CronJob, error) {
	o, err := c.getFactory().Get(c.GVR(), fqn, true, labels.Everything())
	if err != nil {
		return nil, err
	}

	var cj batchv1.CronJob
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &cj)
	if err != nil {
		return nil, errors.New("expecting cronjob resource")
	}

	return &cj, nil
}

// ToggleSuspend toggles suspend/resume on a CronJob.
func (c *CronJob) ToggleSuspend(ctx context.Context, path string) error {
	ns, n := client.Namespaced(path)
	auth, err := c.Client().CanI(ns, c.GVR(), n, []string{client.GetVerb, client.UpdateVerb})
	if err != nil {
		return err
	}
	if !auth {
		return fmt.Errorf("user is not authorized to (un)suspend cronjobs")
	}

	dial, err := c.Client().Dial()
	if err != nil {
		return err
	}
	cj, err := dial.BatchV1().CronJobs(ns).Get(ctx, n, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if cj.Spec.Suspend != nil {
		current := !*cj.Spec.Suspend
		cj.Spec.Suspend = &current
	} else {
		true := true
		cj.Spec.Suspend = &true
	}
	_, err = dial.BatchV1().CronJobs(ns).Update(ctx, cj, metav1.UpdateOptions{})

	return err
}

// Scan scans for cluster resource refs.
func (c *CronJob) Scan(ctx context.Context, gvr client.GVR, fqn string, wait bool) (Refs, error) {
	ns, n := client.Namespaced(fqn)
	oo, err := c.getFactory().List(c.GVR(), ns, wait, labels.Everything())
	if err != nil {
		return nil, err
	}

	refs := make(Refs, 0, len(oo))
	for _, o := range oo {
		var cj batchv1.CronJob
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &cj)
		if err != nil {
			return nil, errors.New("expecting CronJob resource")
		}
		switch gvr {
		case CmGVR:
			if !hasConfigMap(&cj.Spec.JobTemplate.Spec.Template.Spec, n) {
				continue
			}
			refs = append(refs, Ref{
				GVR: c.GVR(),
				FQN: client.FQN(cj.Namespace, cj.Name),
			})
		case SecGVR:
			found, err := hasSecret(c.Factory, &cj.Spec.JobTemplate.Spec.Template.Spec, cj.Namespace, n, wait)
			if err != nil {
				log.Warn().Err(err).Msgf("locate secret %q", fqn)
				continue
			}
			if !found {
				continue
			}
			refs = append(refs, Ref{
				GVR: c.GVR(),
				FQN: client.FQN(cj.Namespace, cj.Name),
			})
		case PcGVR:
			if !hasPC(&cj.Spec.JobTemplate.Spec.Template.Spec, n) {
				continue
			}
			refs = append(refs, Ref{
				GVR: c.GVR(),
				FQN: client.FQN(cj.Namespace, cj.Name),
			})
		}
	}

	return refs, nil
}
