// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"testing"

	"github.com/stretchr/testify/assert"
	batchv1 "k8s.io/api/batch/v1"
)

func TestJob_diagnose(t *testing.T) {
	uu := map[string]struct {
		spec   batchv1.JobSpec
		status batchv1.JobStatus
		err    bool
	}{
		"complete": {
			status: batchv1.JobStatus{Succeeded: 1},
		},
		"complete-after-retry": {
			status: batchv1.JobStatus{Succeeded: 1, Failed: 1},
		},
		"failed": {
			status: batchv1.JobStatus{Failed: 1},
			err:    true,
		},
		"parallel-complete-after-retry": {
			spec:   batchv1.JobSpec{Parallelism: new(int32(3))},
			status: batchv1.JobStatus{Succeeded: 1, Failed: 1},
		},
		"parallel-failed": {
			spec:   batchv1.JobSpec{Parallelism: new(int32(3))},
			status: batchv1.JobStatus{Failed: 2},
			err:    true,
		},
		"completions-partial-failed": {
			spec:   batchv1.JobSpec{Completions: new(int32(3))},
			status: batchv1.JobStatus{Succeeded: 2, Failed: 1},
			err:    true,
		},
	}

	var j Job
	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			err := j.diagnose(toCompletion(&u.spec, &u.status), &u.status)
			assert.Equal(t, u.err, err != nil)
		})
	}
}
