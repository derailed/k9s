// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"strings"
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	dao.MetaAccess.RegisterMeta(client.NodeGVR.String(), &metav1.APIResource{
		Name:         "nodes",
		SingularName: "node",
		Kind:         "Node",
		Verbs:        []string{"get", "list", "watch", "delete"},
		Categories:   []string{"k9s"},
	})
}

func TestResourceSortKeys(t *testing.T) {
	uu := map[string]struct {
		viewer ResourceViewer
	}{
		"pods": {
			viewer: NewPod(client.PodGVR),
		},
		"containers": {
			viewer: NewContainer(client.CoGVR),
		},
		"nodes": {
			viewer: NewNode(client.NodeGVR),
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			require.NoError(t, u.viewer.Init(makeContext(t)))

			for key, desc := range map[tcell.Key]string{
				ui.KeyShiftC: "Sort CPU",
				ui.KeyShiftM: "Sort MEM",
			} {
				a, ok := u.viewer.Actions().Get(key)
				require.True(t, ok, "missing %s binding", tcell.KeyNames[key])
				assert.Equal(t, desc, a.Description)
			}
		})
	}
}

func TestResourceSortOrder(t *testing.T) {
	v := NewNode(client.NodeGVR)
	require.NoError(t, v.Init(makeContext(t)))
	v.GetTable().SetModel(new(mockUsageModel))
	v.GetTable().Refresh()

	press := func(k tcell.Key) {
		a, ok := v.Actions().Get(k)
		require.True(t, ok, "missing %s binding", tcell.KeyNames[k])
		a.Action(nil)
	}
	rows := func() []string {
		ss := make([]string, 0, v.GetTable().GetRowCount()-1)
		for i := 1; i < v.GetTable().GetRowCount(); i++ {
			ss = append(ss, strings.TrimSpace(v.GetTable().GetCell(i, 0).Text))
		}
		return ss
	}

	assert.Equal(t, []string{"r0", "r1", "r2"}, rows())

	press(ui.KeyShiftC)
	assert.Equal(t, []string{"r1", "r2", "r0"}, rows(), "first CPU sort should list the heaviest rows first")
	press(ui.KeyShiftC)
	assert.Equal(t, []string{"r0", "r2", "r1"}, rows(), "second CPU sort should invert the order")

	press(ui.KeyShiftM)
	assert.Equal(t, []string{"r0", "r2", "r1"}, rows(), "first MEM sort should list the heaviest rows first")
	press(ui.KeyShiftM)
	assert.Equal(t, []string{"r1", "r2", "r0"}, rows(), "second MEM sort should invert the order")
}

// ----------------------------------------------------------------------------
// Helpers...

type mockUsageModel struct {
	mockTableModel
}

var _ ui.Tabular = (*mockUsageModel)(nil)

func (*mockUsageModel) Peek() *model1.TableData {
	return model1.NewTableDataWithRows(
		client.NewGVR("test"),
		model1.Header{
			model1.HeaderColumn{Name: "NAMESPACE"},
			model1.HeaderColumn{Name: "NAME"},
			model1.HeaderColumn{Name: cpuCol, Attrs: model1.Attrs{Align: tview.AlignRight, MX: true}},
			model1.HeaderColumn{Name: memCol, Attrs: model1.Attrs{Align: tview.AlignRight, MX: true}},
		},
		model1.NewRowEventsWithEvts(
			model1.RowEvent{Row: model1.Row{ID: "ns1/r0", Fields: model1.Fields{"ns1", "r0", "9", "300"}}},
			model1.RowEvent{Row: model1.Row{ID: "ns1/r1", Fields: model1.Fields{"ns1", "r1", "700", "8"}}},
			model1.RowEvent{Row: model1.Row{ID: "ns1/r2", Fields: model1.Fields{"ns1", "r2", "80", "40"}}},
		),
	)
}
