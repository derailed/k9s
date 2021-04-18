package model_test

import (
	"testing"

	"github.com/derailed/k9s/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestNavHistoryPrev(t *testing.T) {
	h := model.NewNavHistory(3)
	h.Push("a")
	h.Push("b")

	assert.Equal(t, "a", h.Prev())
}
