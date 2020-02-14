package dialog

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
)

func TestPortForwards(t *testing.T) {
	p := ui.NewPages()

	cbFunc := func(path, co string, t dao.Tunnel) {}
	ShowPortForwards(p, config.NewStyles(), "fred", []string{"8080"}, cbFunc)

	d := p.GetPrimitive(portForwardKey).(*tview.ModalForm)
	assert.NotNil(t, d)

	DismissPortForwards(p)
	assert.Nil(t, p.GetPrimitive(portForwardKey))
}

func TestExtractPort(t *testing.T) {
	uu := map[string]struct {
		port, e string
	}{
		"full": {
			"fred:8000", "8000",
		},
		"port": {
			"8000", "8000",
		},
		"protocol": {
			"dns:53╱UDP", "53",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, extractPort(u.port))
		})
	}
}
