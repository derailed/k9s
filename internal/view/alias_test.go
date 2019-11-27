package view_test

import (
	"context"
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/view"
	"github.com/gdamore/tcell"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestAliasNew(t *testing.T) {
	v := view.NewAlias()
	v.Init(makeContext())

	assert.Equal(t, 3, v.GetColumnCount())
	assert.Equal(t, 41, v.GetRowCount())
	assert.Equal(t, "Aliases", v.Name())
	assert.Equal(t, 9, len(v.Hints()))
}

func TestAliasSearch(t *testing.T) {
	v := view.NewAlias()
	v.Init(makeContext())
	v.SearchBuff().SetActive(true)
	v.SearchBuff().Set("dump")

	v.SendKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))

	assert.Equal(t, 3, v.GetColumnCount())
	assert.Equal(t, 1, v.GetRowCount())
}

func TestAliasGoto(t *testing.T) {
	v := view.NewAlias()
	v.Init(makeContext())
	v.Select(0, 0)

	b := buffL{}
	v.SearchBuff().SetActive(true)
	v.SearchBuff().AddListener(&b)
	v.SendKey(tcell.NewEventKey(tcell.KeyEnter, 256, tcell.ModNone))

	assert.True(t, v.SearchBuff().IsActive())
}

// Helpers...

type buffL struct {
	active  int
	changed int
}

func (b *buffL) BufferChanged(s string) {
	b.changed++
}
func (b *buffL) BufferActive(state bool, kind ui.BufferKind) {
	b.active++
}

func makeContext() context.Context {
	a := view.NewApp(config.NewConfig(ks{}))
	ctx := context.WithValue(context.Background(), ui.KeyApp, a)
	return context.WithValue(ctx, ui.KeyStyles, a.Styles)
}

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
