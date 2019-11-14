package view_test

import (
	"context"
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestPodNew(t *testing.T) {
	po := view.NewPod("Pod", "blee", resource.NewPodList(nil, ""))
	po.Init(makeCtx())

	assert.Equal(t, "pods", po.Name())
	assert.Equal(t, 31, len(po.Hints()))
}

// Helpers...

func makeCtx() context.Context {
	cfg := config.NewConfig(ks{})
	return context.WithValue(context.Background(), ui.KeyApp, view.NewApp(cfg))
}
