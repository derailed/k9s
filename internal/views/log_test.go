package views

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
)

func TestAnsi(t *testing.T) {
	buff := bytes.NewBufferString("")
	w := tview.ANSIWriter(buff)
	fmt.Fprintf(w, "[YELLOW] ok")
	assert.Equal(t, "[YELLOW] ok", buff.String())

	v := tview.NewTextView()
	v.SetDynamicColors(true)
	aw := tview.ANSIWriter(v)
	s := "[2019-03-27T15:05:15,246][INFO ][o.e.c.r.a.AllocationService] [es-0] Cluster health status changed from [YELLOW] to [GREEN] (reason: [shards started [[.monitoring-es-6-2019.03.27][0]]"
	fmt.Fprintf(aw, s)
	assert.Equal(t, s+"\n", v.GetText(false))
}
