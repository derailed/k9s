package config_test

import (
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	version "k8s.io/apimachinery/pkg/version"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	disk "k8s.io/client-go/discovery/cached/disk"
	dynamic "k8s.io/client-go/dynamic"
	kubernetes "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	versioned "k8s.io/metrics/pkg/client/clientset/versioned"
)

func newMockConfig() *config.Config {
	cfg := config.NewConfig(newMockKubeSettings(&genericclioptions.ConfigFlags{}))
	cfg.SetConnection(newMockConnection())

	return cfg
}

type mockKubeSettings struct {
	flags *genericclioptions.ConfigFlags
}

func newMockKubeSettings(f *genericclioptions.ConfigFlags) mockKubeSettings {
	return mockKubeSettings{flags: f}
}
func (m mockKubeSettings) CurrentContextName() (string, error) {
	return "minikube", nil
}
func (m mockKubeSettings) CurrentClusterName() (string, error) {
	return "minikube", nil
}
func (m mockKubeSettings) CurrentNamespaceName() (string, error) {
	return "default", nil
}
func (m mockKubeSettings) ClusterNames() (map[string]struct{}, error) {
	return map[string]struct{}{
		"minikube": {},
		"fred":     {},
		"blee":     {},
	}, nil
}

type mockConnection struct{}

func newMockConnection() mockConnection {
	return mockConnection{}
}
func (m mockConnection) CanI(ns, gvr string, verbs []string) (bool, error) {
	return true, nil
}
func (m mockConnection) Config() *client.Config {
	return nil
}
func (m mockConnection) ConnectionOK() bool {
	return true
}
func (m mockConnection) Dial() (kubernetes.Interface, error) {
	return nil, nil
}
func (m mockConnection) DialLogs() (kubernetes.Interface, error) {
	return nil, nil
}
func (m mockConnection) SwitchContext(ctx string) error {
	return nil
}
func (m mockConnection) CachedDiscovery() (*disk.CachedDiscoveryClient, error) {
	return nil, nil
}
func (m mockConnection) RestConfig() (*restclient.Config, error) {
	return nil, nil
}
func (m mockConnection) MXDial() (*versioned.Clientset, error) {
	return nil, nil
}
func (m mockConnection) DynDial() (dynamic.Interface, error) {
	return nil, nil
}
func (m mockConnection) HasMetrics() bool {
	return false
}
func (m mockConnection) ValidNamespaceNames() (client.NamespaceNames, error) {
	return nil, nil
}
func (m mockConnection) IsValidNamespace(string) bool {
	return true
}
func (m mockConnection) ServerVersion() (*version.Info, error) {
	return nil, nil
}
func (m mockConnection) CheckConnectivity() bool {
	return false
}
func (m mockConnection) ActiveCluster() string {
	return ""
}
func (m mockConnection) ActiveNamespace() string {
	return ""
}
func (m mockConnection) IsActiveNamespace(string) bool {
	return false
}
