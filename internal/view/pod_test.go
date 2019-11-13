package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestPodNew(t *testing.T) {
	po := view.NewPod("test", "blee", resource.NewPodList(nil, ""))

	assert.Equal(t, "po", po.Name())
}
