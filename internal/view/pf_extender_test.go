// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"errors"
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/watch"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
)

func TestEnsurePodPortFwdAllowed(t *testing.T) {
	uu := map[string]struct {
		podExists   bool
		podPhase    corev1.PodPhase
		expectError bool
	}{
		"pod-not-exist": {
			expectError: true,
		},
		"pod-pending": {
			podExists:   true,
			podPhase:    corev1.PodPending,
			expectError: true,
		},
		"pod-running": {
			podExists:   true,
			podPhase:    corev1.PodRunning,
			expectError: false,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			f := testFactory{}
			if u.podExists {
				f.expectedGet = &unstructured.Unstructured{
					Object: map[string]any{
						"status": map[string]any{
							"phase": u.podPhase,
						},
					},
				}
			}

			err := ensurePodPortFwdAllowed(f, "ns/name")
			if u.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

type testFactory struct {
	expectedGet runtime.Object
}

var _ dao.Factory = testFactory{}

func (testFactory) Client() client.Connection {
	return nil
}
func (t testFactory) Get(*client.GVR, string, bool, labels.Selector) (runtime.Object, error) {
	if t.expectedGet != nil {
		return t.expectedGet, nil
	}

	return nil, errors.New("not found")
}
func (testFactory) List(*client.GVR, string, bool, labels.Selector) ([]runtime.Object, error) {
	return nil, nil
}
func (testFactory) ForResource(string, *client.GVR) (informers.GenericInformer, error) {
	return nil, nil
}
func (testFactory) CanForResource(string, *client.GVR, []string) (informers.GenericInformer, error) {
	return nil, nil
}
func (testFactory) Forwarders() watch.Forwarders {
	return nil
}
func (testFactory) WaitForCacheSync()      {}
func (testFactory) DeleteForwarder(string) {}
