package config_test

import (
	"testing"

	"github.com/derailed/k9s/config"
	"github.com/stretchr/testify/assert"
)

func TestViewValidate(t *testing.T) {
	v := config.NewView()
	ci := NewMockClusterInfo()

	v.Validate(ci)
	assert.Equal(t, "po", v.Active)

	v.Active = "fred"
	v.Validate(ci)
	assert.Equal(t, "fred", v.Active)
}
