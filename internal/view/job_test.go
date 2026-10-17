// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	dao.MetaAccess.RegisterMeta(client.JobGVR.String(), &metav1.APIResource{
		Name:         "jobs",
		SingularName: "job",
		Namespaced:   true,
		Kind:         "Jobs",
		Verbs:        []string{"get", "list", "watch", "delete", "update"},
		Categories:   []string{"k9s"},
	})
}

func TestJob(t *testing.T) {
	v := view.NewJob(client.JobGVR)

	require.NoError(t, v.Init(makeCtx(t)))
	assert.Equal(t, "Jobs", v.Name())

	var descriptions []string
	for _, h := range v.Hints() {
		descriptions = append(descriptions, h.Description)
	}
	assert.Contains(t, descriptions, "Suspend/Resume")
}
