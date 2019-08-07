package views

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
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
	v := newHelpView(a, nil)
	v.Init(nil, "")

	const e = "🏠 General\n   :<cmd> Command mode\n  /<term> Filter mode\n      esc Clear filter\n      tab Next term match\n  backtab Previous term match\n   Ctrl-r Refresh\n  Shift-i Invert Sort\n        p Previous resource view\n       :q Quit\n\n🤖 View Navigation\n        g Goto Top\n        G Goto Bottom\n   Ctrl-b Page Down\n   Ctrl-f Page Up\n        h Left\n        l Right\n        k Up\n        j Down\n️️\n😱 Help\n        ? Help\n   Ctrl-a Aliases view\n"
	assert.Equal(t, e, v.GetText(true))
	assert.Equal(t, "Help", v.getTitle())
}
