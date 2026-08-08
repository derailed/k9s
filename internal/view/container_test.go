// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContainerNew(t *testing.T) {
	c := view.NewContainer(client.CoGVR)

	require.NoError(t, c.Init(makeCtx(t)))
	assert.Equal(t, "Containers", c.Name())
	assert.Len(t, c.Hints(), 14)
}

func TestShortContainerImageName(t *testing.T) {
	t.Parallel()

	t.Run("digest and registry", func(t *testing.T) {
		in := "europe-west3-docker.pkg.dev/acme/ghcr-io-proxy/derailed/k9s:latest@sha256:947974108bb93df64469fedaa0ab7191cdf86c608217d6ff9b497c4e96ff1069"
		assert.Equal(t, "k9s:latest", view.ShortContainerImageName(in))
	})

	t.Run("keeps simple image", func(t *testing.T) {
		assert.Equal(t, "busybox:1.37", view.ShortContainerImageName("busybox:1.37"))
	})
}
