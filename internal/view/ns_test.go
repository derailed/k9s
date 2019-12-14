package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestNSCleanser(t *testing.T) {
	ns := view.NewNamespace(dao.GVR("v1/namespaces"))
	ns.Init(makeCtx())

	assert.Equal(t, "ns", ns.Name())
	assert.Equal(t, 19, len(ns.Hints()))
}
