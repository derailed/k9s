// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/slogs"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
)

const maxJobNameSize = 42

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
func (c *CronJob) ListImages(_ context.Context, fqn string) ([]string, error) {
	cj, err := c.GetInstance(fqn)
	if err != nil {
		return nil, err
	}

	return render.ExtractImages(&cj.Spec.JobTemplate.Spec.Template.Spec), nil
}

// ContainerOverride represents overrides for a container.
type ContainerOverride struct {
	Command []string
	Args    []string
	Env     []v1.EnvVar
}

// Run a CronJob.
func (c *CronJob) Run(path string) error {
	return c.RunWithOverrides(path, nil)
}

// RunWithOverrides runs a CronJob with container overrides.
func (c *CronJob) RunWithOverrides(path string, overrides map[string]ContainerOverride) error {
	ns, n := client.Namespaced(path)
	auth, err := c.Client().CanI(ns, client.JobGVR, n, []string{client.GetVerb, client.CreateVerb})
	if err != nil {
		return err
	}
	if !auth {
		return fmt.Errorf("user is not authorized to run jobs")
	}

	o, err := c.getFactory().Get(c.gvr, path, true, labels.Everything())
	if err != nil {
		return err
	}
	var cj batchv1.CronJob
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &cj)
	if err != nil {
		return errors.New("expecting CronJob resource")
	}
	if err := applyContainerOverrides(&cj.Spec.JobTemplate.Spec.Template.Spec, overrides); err != nil {
		return err
	}

	jobName := cj.Name
	if len(cj.Name) >= maxJobNameSize {
		jobName = cj.Name[0:maxJobNameSize]
	}
	trueVal := true
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
					BlockOwnerDeletion: &trueVal,
					Controller:         &trueVal,
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

func applyContainerOverrides(spec *v1.PodSpec, overrides map[string]ContainerOverride) error {
	apply := func(containers []v1.Container, name string, override ContainerOverride) bool {
		for i := range containers {
			if containers[i].Name != name {
				continue
			}
			containers[i].Command = slices.Clone(override.Command)
			containers[i].Args = slices.Clone(override.Args)
			containers[i].Env = make([]v1.EnvVar, len(override.Env))
			for j := range override.Env {
				override.Env[j].DeepCopyInto(&containers[i].Env[j])
			}
			return true
		}
		return false
	}

	for name, override := range overrides {
		if !apply(spec.InitContainers, name, override) && !apply(spec.Containers, name, override) {
			return fmt.Errorf("container %q not found in CronJob", name)
		}
	}
	return nil
}

// ScanSA scans for serviceaccount refs.
func (c *CronJob) ScanSA(_ context.Context, fqn string, wait bool) (Refs, error) {
	ns, n := client.Namespaced(fqn)
	oo, err := c.getFactory().List(c.gvr, ns, wait, labels.Everything())
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
	o, err := c.getFactory().Get(c.gvr, fqn, true, labels.Everything())
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
	auth, err := c.Client().CanI(ns, c.gvr, n, []string{client.GetVerb, client.UpdateVerb})
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
		trueVal := true
		cj.Spec.Suspend = &trueVal
	}
	_, err = dial.BatchV1().CronJobs(ns).Update(ctx, cj, metav1.UpdateOptions{})

	return err
}

// Scan scans for cluster resource refs.
func (c *CronJob) Scan(_ context.Context, gvr *client.GVR, fqn string, wait bool) (Refs, error) {
	ns, n := client.Namespaced(fqn)
	oo, err := c.getFactory().List(c.gvr, ns, wait, labels.Everything())
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
		case client.CmGVR:
			if !hasConfigMap(&cj.Spec.JobTemplate.Spec.Template.Spec, n) {
				continue
			}
			refs = append(refs, Ref{
				GVR: c.GVR(),
				FQN: client.FQN(cj.Namespace, cj.Name),
			})
		case client.SecGVR:
			found, err := hasSecret(c.Factory, &cj.Spec.JobTemplate.Spec.Template.Spec, cj.Namespace, n, wait)
			if err != nil {
				slog.Warn("Failed to locate secret",
					slogs.FQN, fqn,
					slogs.Error, err,
				)
				continue
			}
			if !found {
				continue
			}
			refs = append(refs, Ref{
				GVR: c.GVR(),
				FQN: client.FQN(cj.Namespace, cj.Name),
			})
		case client.PcGVR:
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
