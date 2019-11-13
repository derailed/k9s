package model_test

import (
	"sort"
	"testing"

	"github.com/derailed/k9s/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestMenuHintOrder(t *testing.T) {
	h1 := model.MenuHint{Mnemonic: "b", Description: "Duh"}
	h2 := model.MenuHint{Mnemonic: "a", Description: "Blee"}
	h3 := model.MenuHint{Mnemonic: "1", Description: "Zorg"}

	hh := model.MenuHints{h1, h2, h3}
	sort.Sort(hh)

	assert.Equal(t, h3, hh[0])
	assert.Equal(t, h2, hh[1])
	assert.Equal(t, h1, hh[2])
}
