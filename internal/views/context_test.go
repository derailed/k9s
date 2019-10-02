package views

import (
	"testing"

	"github.com/derailed/k9s/internal/k8s"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/resource"
	"github.com/stretchr/testify/assert"
)

func TestContextView(t *testing.T) {
	l := resource.NewContextList(nil, "fred", k8s.GVR{})
	v := newContextView("blee", NewApp(config.NewConfig(ks{})), l).(*contextView)

	assert.Equal(t, 3, len(v.hints()))
}

func TestCleaner(t *testing.T) {
	uu := map[string]struct {
		s, e string
	}{
		"normal":  {"fred", "fred"},
		"default": {"fred*", "fred"},
		"delta":   {"fred(𝜟)", "fred"},
	}

	v := contextView{}
	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, v.cleanser(u.s))
		})
	}
}
