package views

import (
	"testing"

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
	v := detailsView{}
	assert.Equal(t, exp, v.decorateLines(buff, "blee"))
}
