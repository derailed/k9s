package view_test

import (
	"context"
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestPodNew(t *testing.T) {
	po := view.NewPod(client.GVR("v1/pods"))

	assert.Nil(t, po.Init(makeCtx()))
	assert.Equal(t, "Pods", po.Name())
	assert.Equal(t, 16, len(po.Hints()))
}

// Helpers...

func makeCtx() context.Context {
	cfg := config.NewConfig(ks{})
	return context.WithValue(context.Background(), internal.KeyApp, view.NewApp(cfg))
}
