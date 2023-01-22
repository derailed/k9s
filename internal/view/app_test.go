package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestAppNew(t *testing.T) {
	a := view.NewApp(config.NewConfig(ks{}))
	_ = a.Init("blee", 10)

	assert.Equal(t, 11, len(a.GetActions()))
}
