// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/duration"
)

// Job renders a K8s Job to screen.
type Job struct {
	Base
}

// Header returns a header row.
func (Job) Header(ns string) model1.Header {
	return model1.Header{
		model1.HeaderColumn{Name: "NAMESPACE"},
		model1.HeaderColumn{Name: "NAME"},
		model1.HeaderColumn{Name: "VS", VS: true},
		model1.HeaderColumn{Name: "COMPLETIONS"},
		model1.HeaderColumn{Name: "DURATION"},
		model1.HeaderColumn{Name: "SELECTOR", Wide: true},
		model1.HeaderColumn{Name: "CONTAINERS", Wide: true},
		model1.HeaderColumn{Name: "IMAGES", Wide: true},
		model1.HeaderColumn{Name: "VALID", Wide: true},
		model1.HeaderColumn{Name: "AGE", Time: true},
	}
}

// Render renders a K8s resource to screen.
func (j Job) Render(o interface{}, ns string, r *model1.Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected Job, but got %T", o)
	}
	var job batchv1.Job
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &job)
	if err != nil {
		return err
	}
	ready := toCompletion(job.Spec, job.Status)

	cc, ii := toContainers(job.Spec.Template.Spec)

	r.ID = client.MetaFQN(job.ObjectMeta)
	r.Fields = model1.Fields{
		job.Namespace,
		job.Name,
		computeVulScore(job.ObjectMeta, &job.Spec.Template.Spec),
		ready,
		toDuration(job.Status),
		jobSelector(job.Spec),
		cc,
		ii,
		AsStatus(j.diagnose(ready, job.Status)),
		ToAge(job.GetCreationTimestamp()),
	}

	return nil
}

func (Job) diagnose(ready string, status batchv1.JobStatus) error {
	tokens := strings.Split(ready, "/")
	if tokens[0] != tokens[1] && status.Failed > 0 {
		return fmt.Errorf("%d pods failed", status.Failed)
	}
	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

const maxShow = 2

func toContainers(p v1.PodSpec) (string, string) {
	cc, ii := parseContainers(p.InitContainers)
	cn, ci := parseContainers(p.Containers)

	cc, ii = append(cc, cn...), append(ii, ci...)

	// Limit to 2 of each...
	if len(cc) > maxShow {
		cc = append(cc[:2], "(+"+strconv.Itoa(len(cc)-maxShow)+")...")
	}
	if len(ii) > maxShow {
		ii = append(ii[:2], "(+"+strconv.Itoa(len(ii)-maxShow)+")...")
	}

	return strings.Join(cc, ","), strings.Join(ii, ",")
}

func parseContainers(cos []v1.Container) (nn, ii []string) {
	for _, co := range cos {
		nn = append(nn, co.Name)
		ii = append(ii, co.Image)
	}

	return nn, ii
}

func toCompletion(spec batchv1.JobSpec, status batchv1.JobStatus) (s string) {
	if spec.Completions != nil {
		return strconv.Itoa(int(status.Succeeded)) + "/" + strconv.Itoa(int(*spec.Completions))
	}

	if spec.Parallelism == nil {
		return strconv.Itoa(int(status.Succeeded)) + "/1"
	}

	p := *spec.Parallelism
	if p > 1 {
		return strconv.Itoa(int(status.Succeeded)) + "/1 of " + strconv.Itoa(int(p))
	}

	return strconv.Itoa(int(status.Succeeded)) + "/1"
}

func toDuration(status batchv1.JobStatus) string {
	if status.StartTime == nil {
		return MissingValue
	}

	var d time.Duration
	switch {
	case status.CompletionTime == nil:
		d = time.Since(status.StartTime.Time)
	default:
		d = status.CompletionTime.Sub(status.StartTime.Time)
	}

	return duration.HumanDuration(d)
}
