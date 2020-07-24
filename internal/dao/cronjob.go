package dao

import (
	"context"
	"errors"
	"fmt"

	"github.com/derailed/k9s/internal/client"
	"github.com/rs/zerolog/log"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
)

const maxJobNameSize = 42

var (
	_ Accessor = (*CronJob)(nil)
	_ Runnable = (*CronJob)(nil)
)

// CronJob represents a cronjob K8s resource.
type CronJob struct {
	Generic
}

// Run a CronJob.
func (c *CronJob) Run(path string) error {
	ns, _ := client.Namespaced(path)
	auth, err := c.Client().CanI(ns, "batch/v1beta1/cronjobs", []string{client.GetVerb, client.CreateVerb})
	if err != nil {
		return err
	}
	if !auth {
		return fmt.Errorf("user is not authorize to run cronjobs")
	}

	o, err := c.Factory.Get("batch/v1beta1/cronjobs", path, true, labels.Everything())
	if err != nil {
		return err
	}
	var cj batchv1beta1.CronJob
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &cj)
	if err != nil {
		return errors.New("expecting CronJob resource")
	}
	var jobName = cj.Name
	if len(cj.Name) >= maxJobNameSize {
		jobName = cj.Name[0:maxJobNameSize]
	}
	true := true
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName + "-manual-" + rand.String(3),
			Namespace: ns,
			Labels:    cj.Spec.JobTemplate.Labels,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         "batch/v1beta",
					Kind:               "CronJob",
					BlockOwnerDeletion: &true,
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
	oo, err := c.Factory.List(c.GVR(), ns, wait, labels.Everything())
	if err != nil {
		return nil, err
	}

	refs := make(Refs, 0, len(oo))
	for _, o := range oo {
		var cj batchv1beta1.CronJob
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &cj)
		if err != nil {
			return nil, errors.New("expecting CronJob resource")
		}
		if cj.Spec.JobTemplate.Spec.Template.Spec.ServiceAccountName == n {
			refs = append(refs, Ref{
				GVR: c.GVR(),
				FQN: client.FQN(cj.Namespace, cj.Name),
			})
		}
	}

	return refs, nil
}

// Scan scans for cluster resource refs.
func (c *CronJob) Scan(ctx context.Context, gvr, fqn string, wait bool) (Refs, error) {
	ns, n := client.Namespaced(fqn)
	oo, err := c.Factory.List(c.GVR(), ns, wait, labels.Everything())
	if err != nil {
		return nil, err
	}

	refs := make(Refs, 0, len(oo))
	for _, o := range oo {
		var cj batchv1beta1.CronJob
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &cj)
		if err != nil {
			return nil, errors.New("expecting CronJob resource")
		}
		switch gvr {
		case "v1/configmaps":
			if !hasConfigMap(&cj.Spec.JobTemplate.Spec.Template.Spec, n) {
				continue
			}
			refs = append(refs, Ref{
				GVR: c.GVR(),
				FQN: client.FQN(cj.Namespace, cj.Name),
			})
		case "v1/secrets":
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
		}
	}

	return refs, nil
}
