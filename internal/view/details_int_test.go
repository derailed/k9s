package view

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestDetailsDecorateLines(t *testing.T) {
	buff := `
	I love blee
	blee is much [blue::]cooler [green::]than foo!
	`
	exp := `
	I love ["0"]blee[""]
	["1"]blee[""] is much [blue::]cooler [green::]than foo!
	`

	app := NewApp(config.NewConfig(ks{}))
	v := NewDetails(app, nil)

	assert.Equal(t, exp, v.decorateLines(buff, "blee"))
}
