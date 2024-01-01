// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view_test

import (
	"context"
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config/mock"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestPodNew(t *testing.T) {
	po := view.NewPod(client.NewGVR("v1/pods"))

	assert.Nil(t, po.Init(makeCtx()))
	assert.Equal(t, "Pods", po.Name())
	assert.Equal(t, 28, len(po.Hints()))
}

// Helpers...

func makeCtx() context.Context {
	cfg := mock.NewMockConfig()
	return context.WithValue(context.Background(), internal.KeyApp, view.NewApp(cfg))
}
