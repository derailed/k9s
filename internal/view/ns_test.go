package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestNSCleanser(t *testing.T) {
	ns := view.NewNamespace("ns", "", resource.NewNamespaceList(nil, ""))
	ns.Init(makeCtx())

	assert.Equal(t, "ns", ns.Name())
	assert.Equal(t, 20, len(ns.Hints()))
}
