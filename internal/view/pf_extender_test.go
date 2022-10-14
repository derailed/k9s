package view

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/watch"
)

func TestEnsurePodPortFwdAllowed(t *testing.T) {
	testCases := []struct {
		name        string
		podExists   bool
		podPhase    corev1.PodPhase
		expectError bool
	}{
		{
			name:        "pod_doesnt_exist",
			expectError: true,
		},
		{
			name:        "pod_exists_pending",
			podExists:   true,
			podPhase:    corev1.PodPending,
			expectError: true,
		},
		{
			name:        "pod_is_running",
			podExists:   true,
			podPhase:    corev1.PodRunning,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			f := testFactory{}
			if tc.podExists {
				f.expectedGet = &unstructured.Unstructured{
					Object: map[string]interface{}{
						"status": map[string]interface{}{
							"phase": tc.podPhase,
						},
					},
				}
			}

			err := ensurePodPortFwdAllowed(f, "ns/name")
			if tc.expectError {
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

func (t testFactory) Client() client.Connection {
	return nil
}

func (t testFactory) Get(string, string, bool, labels.Selector) (runtime.Object, error) {
	if t.expectedGet != nil {
		return t.expectedGet, nil
	}

	return nil, errors.New("not found")
}

func (t testFactory) List(string, string, bool, labels.Selector) ([]runtime.Object, error) {
	return nil, nil
}

func (t testFactory) ForResource(string, string) (informers.GenericInformer, error) {
	return nil, nil
}

func (t testFactory) CanForResource(string, string, []string) (informers.GenericInformer, error) {
	return nil, nil
}

func (t testFactory) Forwarders() watch.Forwarders {
	return nil
}

func (t testFactory) WaitForCacheSync() {}

func (t testFactory) DeleteForwarder(string) {}
