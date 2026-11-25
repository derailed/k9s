// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/config/mock"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func TestHeaderIndex(t *testing.T) {
	uu := map[string]struct {
		colName string
		cells   []string
		eok     bool
		e       int
	}{
		"simple": {
			cells:   []string{"NAMESPACE", "NAME", "AGE"},
			colName: "NAME",
			eok:     true,
			e:       1,
		},

		"missing": {
			cells:   []string{"NAMESPACE", "BLEE", "AGE"},
			colName: "NAME",
		},

		"decorated": {
			cells:   []string{"[#DADEE8::]NAMESPACE[::]", "[#DADEE8::]NAME[::]", "[#DADEE8::]AGE[::]"},
			colName: "NAME",
			eok:     true,
			e:       1,
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			ta := NewTable(client.NewGVR("test"))
			require.NoError(t, ta.Init(makeContext(t)))
			for i, c := range u.cells {
				ta.AddHeaderCell(i, model1.HeaderColumn{Name: c})
			}

			i, ok := ta.HeaderIndex(u.colName)
			assert.Equal(t, u.eok, ok)
			assert.Equal(t, u.e, i)
		})
	}
}

func TestTableSave(t *testing.T) {
	v := NewTable(client.NewGVR("test"))
	require.NoError(t, v.Init(makeContext(t)))
	v.SetTitle("k9s-test")

	require.NoError(t, ensureDumpDir("/tmp/test-dumps"))
	dir := v.app.Config.K9s.ContextScreenDumpDir()
	c1, _ := os.ReadDir(dir)
	v.saveCmd(nil)

	c2, _ := os.ReadDir(dir)
	assert.Len(t, c2, len(c1)+1)
}

func TestComputeFilenameKeepsDirCase(t *testing.T) {
	// Context names often have capitals, and their dump dir is created as is.
	dir := filepath.Join(t.TempDir(), "MyCluster", "Prod")

	f, err := computeFilename(dir, "default", "Pods", "")
	require.NoError(t, err)
	assert.Equal(t, dir, filepath.Dir(f))
	assert.True(t, strings.HasPrefix(filepath.Base(f), "pods-default-"), filepath.Base(f))

	f, err = saveTable(dir, "Pods", "", model1.NewTableData(client.PodGVR))
	require.NoError(t, err)
	assert.FileExists(t, f)
}

func TestTableNew(t *testing.T) {
	v := NewTable(client.NewGVR("test"))
	require.NoError(t, v.Init(makeContext(t)))

	data := model1.NewTableDataWithRows(
		client.NewGVR("test"),
		model1.Header{
			model1.HeaderColumn{Name: "NAMESPACE"},
			model1.HeaderColumn{Name: "NAME", Attrs: model1.Attrs{Align: tview.AlignRight}},
			model1.HeaderColumn{Name: "FRED"},
			model1.HeaderColumn{Name: "AGE", Attrs: model1.Attrs{Time: true, Decorator: render.AgeDecorator}},
		},
		model1.NewRowEventsWithEvts(
			model1.RowEvent{
				Row: model1.Row{
					Fields: model1.Fields{"ns1", "a", "10", "3m"},
				},
			},
			model1.RowEvent{
				Row: model1.Row{
					Fields: model1.Fields{"ns1", "b", "15", "1m"},
				},
			},
		),
	)
	cdata := v.Update(data, false)
	v.UpdateUI(cdata, data)

	assert.Equal(t, 3, v.GetRowCount())
}

func TestTableViewFilter(t *testing.T) {
	v := NewTable(client.NewGVR("test"))
	require.NoError(t, v.Init(makeContext(t)))
	v.SetModel(&mockTableModel{})
	v.Refresh()

	v.CmdBuff().SetActive(true)
	v.CmdBuff().SetText("blee", "", true)

	assert.Equal(t, 5, v.GetRowCount())
}

func TestTableDefaultEnvFilter(t *testing.T) {
	uu := map[string]struct {
		gvr           *client.GVR
		rowID         string
		fields        model1.Fields
		filter        string
		resetFilter   bool
		wantFilter    string
		wantRawFilter string
		wantNS        string
		wantName      string
	}{
		"no-filter-falls-back-to-selected-name": {
			gvr:           client.PodGVR,
			rowID:         "kube-system/coredns-559f6c778d-z5d7w",
			fields:        model1.Fields{"kube-system", "coredns-559f6c778d-z5d7w", "Running"},
			wantFilter:    "coredns-559f6c778d-z5d7w",
			wantRawFilter: "",
			wantNS:        "kube-system",
			wantName:      "coredns-559f6c778d-z5d7w",
		},
		"active-filter": {
			gvr:           client.PodGVR,
			rowID:         "kube-system/coredns-559f6c778d-z5d7w",
			fields:        model1.Fields{"kube-system", "coredns-559f6c778d-z5d7w", "Running"},
			filter:        "coredns",
			wantFilter:    "coredns",
			wantRawFilter: "coredns",
			wantNS:        "kube-system",
			wantName:      "coredns-559f6c778d-z5d7w",
		},
		"label-selector-filter": {
			gvr:           client.PodGVR,
			rowID:         "kube-system/coredns-559f6c778d-z5d7w",
			fields:        model1.Fields{"kube-system", "coredns-559f6c778d-z5d7w", "Running"},
			filter:        "-l k8s-app=kube-dns",
			wantFilter:    "-l k8s-app=kube-dns",
			wantRawFilter: "-l k8s-app=kube-dns",
			wantNS:        "kube-system",
			wantName:      "coredns-559f6c778d-z5d7w",
		},
		"inverse-filter": {
			gvr:           client.PodGVR,
			rowID:         "default/nginx-1",
			fields:        model1.Fields{"default", "nginx-1", "Running"},
			filter:        "!kube-system",
			wantFilter:    "!kube-system",
			wantRawFilter: "!kube-system",
			wantNS:        "default",
			wantName:      "nginx-1",
		},
		"reset-filter-restores-empty-raw-filter": {
			gvr:           client.PodGVR,
			rowID:         "kube-system/coredns-559f6c778d-z5d7w",
			fields:        model1.Fields{"kube-system", "coredns-559f6c778d-z5d7w", "Running"},
			filter:        "coredns",
			resetFilter:   true,
			wantFilter:    "coredns-559f6c778d-z5d7w",
			wantRawFilter: "",
			wantNS:        "kube-system",
			wantName:      "coredns-559f6c778d-z5d7w",
		},
		"cluster-scoped-no-filter": {
			gvr:           client.NodeGVR,
			rowID:         "worker-node-1",
			fields:        model1.Fields{"", "worker-node-1", "Ready"},
			wantFilter:    "worker-node-1",
			wantRawFilter: "",
			wantNS:        "",
			wantName:      "worker-node-1",
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			v := NewTable(u.gvr)
			require.NoError(t, v.Init(makeEnvContext(t)))
			v.CmdBuff().RemoveListener(v)

			td := makeEnvTableData(u.gvr, u.rowID, u.fields)
			v.SetModel(&envTableModel{data: td})
			v.UpdateUI(td, td)
			v.Select(1, 0)

			if u.filter != "" {
				v.CmdBuff().SetText(u.filter, "", true)
			}
			if u.resetFilter {
				v.CmdBuff().Reset()
			}

			env := v.EnvFn()()
			assert.Contains(t, env, "RAW_FILTER")
			assert.Equal(t, u.wantRawFilter, env["RAW_FILTER"])
			assert.Equal(t, u.wantFilter, env["FILTER"])
			assert.Equal(t, u.wantNS, env["NAMESPACE"])
			assert.Equal(t, u.wantName, env["NAME"])
			assert.Equal(t, u.gvr.G(), env["RESOURCE_GROUP"])
			assert.Equal(t, u.gvr.V(), env["RESOURCE_VERSION"])
			assert.Equal(t, u.gvr.R(), env["RESOURCE_NAME"])

			sub, err := env.Substitute(`FILTER=[$FILTER] RAW_FILTER=[$RAW_FILTER]`)
			require.NoError(t, err)
			assert.Equal(t, "FILTER=["+u.wantFilter+"] RAW_FILTER=["+u.wantRawFilter+"]", sub)
		})
	}
}

func TestTableViewSort(t *testing.T) {
	v := NewTable(client.NewGVR("test"))
	require.NoError(t, v.Init(makeContext(t)))
	v.SetModel(new(mockTableModel))

	uu := map[string]struct {
		sortCol  string
		sorted   []string
		reversed []string
	}{
		"by_name": {
			sortCol:  "NAME",
			sorted:   []string{"r0", "r1", "r2", "r3"},
			reversed: []string{"r3", "r2", "r1", "r0"},
		},
		"by_age": {
			sortCol:  "AGE",
			sorted:   []string{"r0", "r1", "r2", "r3"},
			reversed: []string{"r3", "r2", "r1", "r0"},
		},
		"by_fred": {
			sortCol:  "FRED",
			sorted:   []string{"r3", "r2", "r0", "r1"},
			reversed: []string{"r1", "r0", "r2", "r3"},
		},
	}

	for k := range uu {
		u := uu[k]
		v.SortColCmd(u.sortCol, true)(nil)
		assert.Len(t, u.sorted, v.GetRowCount()-1)
		for i, s := range u.sorted {
			assert.Equal(t, s, v.GetCell(i+1, 0).Text)
		}
		v.SortInvertCmd(nil)
		assert.Len(t, u.reversed, v.GetRowCount()-1)
		for i, s := range u.reversed {
			assert.Equal(t, s, v.GetCell(i+1, 0).Text)
		}
	}
}

// ----------------------------------------------------------------------------
// Helpers...

type mockTableModel struct{}

var _ ui.Tabular = (*mockTableModel)(nil)

func (*mockTableModel) SetViewSetting(context.Context, *config.ViewSetting) {}
func (*mockTableModel) SetInstance(string)                                  {}
func (*mockTableModel) SetLabelSelector(labels.Selector)                    {}
func (*mockTableModel) GetLabelSelector() labels.Selector                   { return nil }
func (*mockTableModel) Empty() bool                                         { return false }
func (*mockTableModel) RowCount() int                                       { return 1 }
func (*mockTableModel) HasMetrics() bool                                    { return true }
func (*mockTableModel) Peek() *model1.TableData                             { return makeTableData() }
func (*mockTableModel) Refresh(context.Context) error                       { return nil }
func (*mockTableModel) ClusterWide() bool                                   { return false }
func (*mockTableModel) GetNamespace() string                                { return "blee" }
func (*mockTableModel) SetNamespace(string)                                 {}
func (*mockTableModel) ToggleToast()                                        {}
func (*mockTableModel) AddListener(model.TableListener)                     {}
func (*mockTableModel) RemoveListener(model.TableListener)                  {}
func (*mockTableModel) Watch(context.Context) error                         { return nil }
func (*mockTableModel) Get(context.Context, string) (runtime.Object, error) {
	return nil, nil
}
func (*mockTableModel) Delete(context.Context, string, *metav1.DeletionPropagation, dao.Grace) error {
	return nil
}
func (*mockTableModel) Describe(context.Context, string) (string, error) {
	return "", nil
}
func (*mockTableModel) ToYAML(context.Context, string) (string, error) {
	return "", nil
}
func (*mockTableModel) InNamespace(string) bool      { return true }
func (*mockTableModel) SetRefreshRate(time.Duration) {}

func makeTableData() *model1.TableData {
	return model1.NewTableDataWithRows(
		client.NewGVR("test"),
		model1.Header{
			model1.HeaderColumn{Name: "NAMESPACE"},
			model1.HeaderColumn{Name: "NAME", Attrs: model1.Attrs{Align: tview.AlignRight}},
			model1.HeaderColumn{Name: "FRED"},
			model1.HeaderColumn{Name: "AGE", Attrs: model1.Attrs{Time: true}},
		},
		model1.NewRowEventsWithEvts(
			model1.RowEvent{
				Row: model1.Row{
					Fields: model1.Fields{"ns1", "r3", "10", "3y125d"},
				},
			},
			model1.RowEvent{
				Row: model1.Row{
					Fields: model1.Fields{"ns1", "r2", "15", "2y12d"},
				},
				Deltas: model1.DeltaRow{"", "", "20", ""},
			},
			model1.RowEvent{
				Row: model1.Row{
					Fields: model1.Fields{"ns1", "r1", "20", "19h"},
				},
			},
			model1.RowEvent{
				Row: model1.Row{
					Fields: model1.Fields{"ns1", "r0", "15", "10s"},
				},
			},
		),
	)
}

func makeContext(t *testing.T) context.Context {
	a := NewApp(mock.NewMockConfig(t))
	ctx := context.WithValue(context.Background(), internal.KeyApp, a)
	return context.WithValue(ctx, internal.KeyStyles, a.Styles)
}

type envTestConn struct {
	client.Connection

	cfg *client.Config
}

func (c envTestConn) Config() *client.Config {
	return c.cfg
}

type envTableModel struct {
	mockTableModel

	data *model1.TableData
}

func (m *envTableModel) Peek() *model1.TableData {
	return m.data
}

func makeEnvApp(t *testing.T) *App {
	t.Helper()

	cl, ct, cfgPath, usr := "cl-1", "ct-1-1", "cfg-1", "user-1"
	flags := genericclioptions.ConfigFlags{
		ClusterName:  &cl,
		Context:      &ct,
		AuthInfoName: &usr,
		KubeConfig:   &cfgPath,
	}
	cfg := mock.NewMockConfig(t)
	cfg.SetConnection(envTestConn{
		Connection: mock.NewMockConnection(),
		cfg:        client.NewConfig(&flags),
	})

	return NewApp(cfg)
}

func makeEnvContext(t *testing.T) context.Context {
	t.Helper()

	a := makeEnvApp(t)
	ctx := context.WithValue(context.Background(), internal.KeyApp, a)

	return context.WithValue(ctx, internal.KeyStyles, a.Styles)
}

func makeEnvTableData(gvr *client.GVR, rowID string, fields model1.Fields) *model1.TableData {
	return model1.NewTableDataWithRows(
		gvr,
		model1.Header{
			model1.HeaderColumn{Name: "NAMESPACE"},
			model1.HeaderColumn{Name: "NAME"},
			model1.HeaderColumn{Name: "STATUS"},
		},
		model1.NewRowEventsWithEvts(
			model1.RowEvent{
				Row: model1.Row{
					ID:     rowID,
					Fields: fields,
				},
			},
		),
	)
}

func ensureDumpDir(n string) error {
	config.AppDumpsDir = n
	if _, err := os.Stat(n); errors.Is(err, fs.ErrNotExist) {
		return os.Mkdir(n, 0700)
	}
	if err := os.RemoveAll(n); err != nil {
		return err
	}
	return os.Mkdir(n, 0700)
}
