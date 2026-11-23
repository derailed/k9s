// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config/mock"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/watch"
	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func TestAppSuggestCommandConcurrentAliasReload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aliases.yaml")
	require.NoError(t, os.WriteFile(path, []byte("aliases:\n  pp: v1/pods\n"), 0600))

	kubeconfig := filepath.Join(t.TempDir(), "kubeconfig")
	a := NewApp(mock.NewMockConfig(t))
	a.factory = watch.NewFactory(&stubConn{
		cfg: client.NewConfig(&genericclioptions.ConfigFlags{KubeConfig: &kubeconfig}),
	})
	a.command = NewCommand(a)
	a.command.alias = dao.NewAlias(a.factory)
	suggest := a.suggestCommand()

	// Typing in the prompt while the cluster refresh reloads aliases.
	var wg sync.WaitGroup
	wg.Go(func() {
		for range 1_000 {
			suggest("p")
		}
	})
	wg.Go(func() {
		for range 1_000 {
			_ = a.command.alias.LoadFile(path)
		}
	})
	wg.Wait()
}

type stubConn struct {
	client.Connection

	cfg *client.Config
}

func (c *stubConn) Config() *client.Config { return c.cfg }

func (*stubConn) ValidNamespaceNames() (client.NamespaceNames, error) { return nil, nil }
