package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestAppNew(t *testing.T) {
	a := view.NewApp(config.NewConfig(ks{}))
	a.Init("blee", 10)

	assert.Equal(t, 9, len(a.GetActions()))
}
