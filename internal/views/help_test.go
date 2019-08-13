package views

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ks struct{}

func (k ks) CurrentContextName() (string, error) {
	return "test", nil
}

func (k ks) CurrentClusterName() (string, error) {
	return "test", nil
}

func (k ks) CurrentNamespaceName() (string, error) {
	return "test", nil
}

func (k ks) ClusterNames() ([]string, error) {
	return []string{"test"}, nil
}

func (k ks) NamespaceNames(nn []v1.Namespace) []string {
	return []string{"test"}
}

func newNS(n string) v1.Namespace {
	return v1.Namespace{ObjectMeta: metav1.ObjectMeta{
		Name: n,
	}}
}

func TestNewHelpView(t *testing.T) {
	cfg := config.NewConfig(ks{})
	a := NewApp(cfg)

	v := newHelpView(a, nil, ui.Hints{{"blee", "duh"}})
	v.Init(nil, "")

	assert.Equal(t, "<blee>", v.GetCell(1, 0).Text)
	assert.Equal(t, "duh", v.GetCell(1, 1).Text)
}
