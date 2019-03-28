package resource

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestJobToCompletion(t *testing.T) {
	t0 := testTime()
	t1, t2 := metav1.Time{t0}, metav1.Time{t0.Add(10 * time.Second)}
	var c, p int32 = 10, 20

	uu := []struct {
		j batchv1.JobSpec
		s batchv1.JobStatus
		e string
	}{
		{
			batchv1.JobSpec{
				Completions: &c,
				Parallelism: &p,
			},
			batchv1.JobStatus{
				Succeeded:      1,
				Active:         1,
				Failed:         0,
				StartTime:      &t1,
				CompletionTime: &t2,
			},
			"1/10",
		},
		{
			batchv1.JobSpec{
				Parallelism: &p,
			},
			batchv1.JobStatus{
				Succeeded:      1,
				Active:         1,
				Failed:         0,
				StartTime:      &t1,
				CompletionTime: &t2,
			},
			"1/1 of 20",
		},
		{
			batchv1.JobSpec{
				Completions: &c,
			},
			batchv1.JobStatus{
				Succeeded:      1,
				Active:         1,
				Failed:         0,
				StartTime:      &t1,
				CompletionTime: &t2,
			},
			"1/10",
		},
		{
			batchv1.JobSpec{},
			batchv1.JobStatus{
				Succeeded:      1,
				Active:         1,
				Failed:         0,
				StartTime:      &t1,
				CompletionTime: &t2,
			},
			"1/1",
		},
	}

	var j *Job
	for _, u := range uu {
		assert.Equal(t, u.e, j.toCompletion(u.j, u.s))
	}
}

func TestJobToDuration(t *testing.T) {
	t0 := testTime().UTC()
	t1, t2 := metav1.Time{t0}, metav1.Time{t0.Add(10 * time.Second)}

	uu := []struct {
		s batchv1.JobStatus
		e string
	}{
		{
			batchv1.JobStatus{
				StartTime:      &t1,
				CompletionTime: &t2,
			},
			"10s",
		},
		{
			batchv1.JobStatus{
				StartTime: &metav1.Time{time.Now().Add(-10 * time.Second)},
			},
			"10s",
		},
		{
			batchv1.JobStatus{
				CompletionTime: &t2,
			},
			MissingValue,
		},
	}

	var j *Job
	for _, u := range uu {
		assert.Equal(t, u.e, j.toDuration(u.s))
	}
}

func TestJobToContainers(t *testing.T) {
	uu := []struct {
		s    v1.PodSpec
		c, i string
	}{
		{
			v1.PodSpec{
				InitContainers: []v1.Container{
					{Name: "i1", Image: "fred"},
				},
				Containers: []v1.Container{
					{Name: "c1", Image: "blee"},
				},
			},
			"i1,c1", "fred,blee",
		},
		{
			v1.PodSpec{
				InitContainers: []v1.Container{
					{Name: "i1", Image: "fred"},
				},
				Containers: []v1.Container{
					{Name: "c1", Image: "blee"},
					{Name: "c2", Image: "duh"},
				},
			},
			"i1,c1,(+1)...", "fred,blee,(+1)...",
		},
	}

	var j *Job
	for _, u := range uu {
		c, i := j.toContainers(u.s)
		assert.Equal(t, u.c, c)
		assert.Equal(t, u.i, i)
	}
}
