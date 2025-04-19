// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package mock

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/require"
	version "k8s.io/apimachinery/pkg/version"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	disk "k8s.io/client-go/discovery/cached/disk"
	dynamic "k8s.io/client-go/dynamic"
	kubernetes "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
	versioned "k8s.io/metrics/pkg/client/clientset/versioned"
)

func EnsureDir(d string) error {
	if _, err := os.Stat(d); errors.Is(err, fs.ErrNotExist) {
		return os.MkdirAll(d, 0700)
	}
	if err := os.RemoveAll(d); err != nil {
		return err
	}

	return os.MkdirAll(d, 0700)
}

func NewMockConfig(t testing.TB) *config.Config {
	if _, err := os.Stat("/tmp/test"); err == nil {
		if e := os.RemoveAll("/tmp/test"); e != nil {
			require.NoError(t, e)
		}
	}
	config.AppContextsDir = "/tmp/test"
	cl, ct := "cl-1", "ct-1-1"
	flags := genericclioptions.ConfigFlags{
		ClusterName: &cl,
		Context:     &ct,
	}
	cfg := config.NewConfig(
		NewMockKubeSettings(&flags),
	)

	return cfg
}

type mockKubeSettings struct {
	flags *genericclioptions.ConfigFlags
	cts   map[string]*api.Context
}

func NewMockKubeSettings(f *genericclioptions.ConfigFlags) mockKubeSettings {
	_, idx, _ := strings.Cut(*f.ClusterName, "-")
	ctId := "ct-" + idx

	return mockKubeSettings{
		flags: f,
		cts: map[string]*api.Context{
			ctId + "-1": {
				Cluster:   *f.ClusterName,
				Namespace: "",
			},
			ctId + "-2": {
				Cluster:   *f.ClusterName,
				Namespace: "ns-2",
			},
			ctId + "-3": {
				Cluster:   *f.ClusterName,
				Namespace: client.DefaultNamespace,
			},
			"fred-blee": {
				Cluster:   "arn:aws:eks:eu-central-1:xxx:cluster/fred-blee",
				Namespace: client.DefaultNamespace,
			},
		},
	}
}
func (m mockKubeSettings) CurrentContextName() (string, error) {
	return *m.flags.Context, nil
}
func (m mockKubeSettings) CurrentClusterName() (string, error) {
	return *m.flags.ClusterName, nil
}
func (mockKubeSettings) CurrentNamespaceName() (string, error) {
	return "default", nil
}
func (m mockKubeSettings) GetContext(s string) (*api.Context, error) {
	ct, ok := m.cts[s]
	if !ok {
		return nil, fmt.Errorf("no context found for: %q", s)
	}
	return ct, nil
}
func (m mockKubeSettings) CurrentContext() (*api.Context, error) {
	return m.GetContext(*m.flags.Context)
}
func (m mockKubeSettings) ContextNames() (map[string]struct{}, error) {
	mm := make(map[string]struct{}, len(m.cts))
	for k := range m.cts {
		mm[k] = struct{}{}
	}

	return mm, nil
}

func (mockKubeSettings) SetProxy(func(*http.Request) (*url.URL, error)) {}

type mockConnection struct {
	ct string
}

func NewMockConnection() mockConnection {
	return mockConnection{}
}
func NewMockConnectionWithContext(ct string) mockConnection {
	return mockConnection{ct: ct}
}

func (mockConnection) CanI(string, *client.GVR, string, []string) (bool, error) {
	return true, nil
}
func (mockConnection) Config() *client.Config {
	return nil
}
func (mockConnection) ConnectionOK() bool {
	return false
}
func (mockConnection) Dial() (kubernetes.Interface, error) {
	return nil, nil
}
func (mockConnection) DialLogs() (kubernetes.Interface, error) {
	return nil, nil
}
func (mockConnection) SwitchContext(string) error {
	return nil
}
func (mockConnection) CachedDiscovery() (*disk.CachedDiscoveryClient, error) {
	return nil, nil
}
func (mockConnection) RestConfig() (*restclient.Config, error) {
	return nil, nil
}
func (mockConnection) MXDial() (*versioned.Clientset, error) {
	return nil, nil
}
func (mockConnection) DynDial() (dynamic.Interface, error) {
	return nil, nil
}
func (mockConnection) HasMetrics() bool {
	return false
}
func (mockConnection) ValidNamespaceNames() (client.NamespaceNames, error) {
	return nil, nil
}
func (mockConnection) IsValidNamespace(string) bool {
	return true
}
func (mockConnection) ServerVersion() (*version.Info, error) {
	return nil, nil
}
func (mockConnection) CheckConnectivity() bool {
	return false
}
func (m mockConnection) ActiveContext() string {
	return m.ct
}
func (mockConnection) ActiveNamespace() string {
	return ""
}
func (mockConnection) IsActiveNamespace(string) bool {
	return false
}
