package view

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

type (
	TableInfo interface {
		Header() render.HeaderRow
		GetCache() render.RowEvents
		SetCache(render.RowEvents)
	}

	// Subject presents a user/group viewer.
	Subject struct {
		*Table

		subjectKind string
		cache       render.RowEvents
	}
)

// NewSubject returns a new subject viewer.
func NewSubject(title, _ string, _ resource.List) ResourceViewer {
	return &Subject{Table: NewTable(title)}
}

func (*Subject) SetContextFn(ContextFunc) {}

// GVR returns a resource descriptor.
func (s *Subject) GVR() string {
	return "n/a"
}

// GetTable returns the table view.
func (s *Subject) GetTable() *Table { return s.Table }

// SetEnvFn sets up K9s env vars.
func (s *Subject) SetEnvFn(EnvFunc) {}

// List returns the resource lister.
func (s *Subject) List() resource.List { return nil }

// SetPath sets parent selector.
func (s *Subject) SetPath(_ string) {}

// Init initializes the view.
func (s *Subject) Init(ctx context.Context) error {
	app, err := extractApp(ctx)
	if err != nil {
		return err
	}
	s.subjectKind = mapCmdSubject(app.Config.K9s.ActiveCluster().View.Active)
	s.Table = NewTable(s.subjectKind)
	s.SetColorerFn(render.Subject{}.ColorerFunc())
	if err := s.Table.Init(ctx); err != nil {
		return err
	}
	s.SetSortCol(1, len(s.Header()), true)
	s.SelectRow(1, true)
	s.bindKeys()
	s.refresh()

	return nil
}

// Start runs the refresh loop.
func (s *Subject) Start() {
	s.Stop()

	var ctx context.Context
	ctx, s.cancelFn = context.WithCancel(context.Background())
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				log.Debug().Msgf("Subject:%s Watch bailing out!", s.subjectKind)
				return
			case <-time.After(time.Duration(s.app.Config.K9s.GetRefreshRate()) * time.Second):
				s.refresh()
			}
		}
	}(ctx)
}

// Name returns the component name
func (s *Subject) Name() string {
	return "subject"
}

func (s *Subject) bindKeys() {
	s.Actions().Delete(ui.KeyShiftA, ui.KeyShiftP, tcell.KeyCtrlSpace, ui.KeySpace)
	s.Actions().Add(ui.KeyActions{
		tcell.KeyEnter:  ui.NewKeyAction("Policies", s.policyCmd, true),
		tcell.KeyEscape: ui.NewKeyAction("Back", s.resetCmd, false),
		ui.KeySlash:     ui.NewKeyAction("Filter", s.activateCmd, false),
		ui.KeyShiftK:    ui.NewKeyAction("Sort Kind", s.SortColCmd(1, true), false),
	})
}

// SetSubject sets the subject name.
func (s *Subject) SetSubject(n string) {
	s.subjectKind = mapSubject(n)
}

func (s *Subject) refresh() {
	log.Debug().Msgf("Refreshing Subject...")
	data, err := s.reconcile()
	if err != nil {
		log.Error().Err(err).Msgf("Refresh for %s", s.subjectKind)
		s.app.Flash().Err(err)
	}
	s.app.QueueUpdateDraw(func() {
		s.Update(data)
	})
}

func (s *Subject) policyCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !s.RowSelected() {
		return evt
	}

	_, n := namespaced(s.GetSelectedItem())
	subject, err := mapFuSubject(s.subjectKind)
	if err != nil {
		s.app.Flash().Err(err)
		return nil
	}
	s.app.inject(NewPolicy(s.app, subject, n))

	return nil
}

func (s *Subject) resetCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !s.SearchBuff().Empty() {
		s.SearchBuff().Reset()
		return nil
	}

	return s.backCmd(evt)
}

func (s *Subject) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	if s.SearchBuff().IsActive() {
		s.SearchBuff().Reset()
		return nil
	}

	return s.app.PrevCmd(evt)
}

func (s *Subject) reconcile() (render.TableData, error) {
	var table render.TableData
	if s.app.Conn() == nil {
		return table, nil
	}

	rows, err := s.fetchClusterRoleBindings()
	if err != nil {
		return table, err
	}

	nrows, err := s.fetchRoleBindings()
	if err != nil {
		return table, err
	}
	for k, v := range nrows {
		rows[k] = v
	}

	return buildTable(s, rows), nil
}

func (s *Subject) Header() render.HeaderRow {
	return render.Subject{}.Header(render.AllNamespaces)
}

func (s *Subject) GetCache() render.RowEvents {
	return s.cache
}

func (s *Subject) SetCache(rows render.RowEvents) {
	s.cache = rows
}

func buildTable(c TableInfo, rows render.Rows) render.TableData {
	table := render.TableData{
		Header:    c.Header(),
		Namespace: "*",
	}

	cache := c.GetCache()
	if len(cache) == 0 {
		cache := make(render.RowEvents, 0, len(rows))
		for _, row := range rows {
			cache = append(cache, render.RowEvent{Kind: render.EventAdd, Row: row})
		}
		table.RowEvents = cache
		return table
	}

	for _, row := range rows {
		idx, ok := cache.FindIndex(row.ID)
		if !ok {
			cache = append(cache, render.RowEvent{Kind: render.EventAdd, Row: row})
			continue
		}

		old := cache[idx].Row
		deltas := make(render.DeltaRow, len(row.Fields))
		if reflect.DeepEqual(old, row) {
			cache[idx].Kind = render.EventUnchanged
			cache[idx].Deltas = deltas
			continue
		}

		cache[idx].Kind = render.EventUpdate
		for i, field := range old.Fields {
			if field != row.Fields[i] {
				deltas[i] = field
			}
		}
		cache[idx].Deltas = deltas
	}

	for _, row := range rows {
		if _, ok := cache.FindIndex(row.ID); !ok {
			cache.Delete(row.ID)
		}
	}
	table.RowEvents = cache

	return table
}

func (s *Subject) fetchClusterRoleBindings() (render.Rows, error) {
	s.app.factory.Preload(render.ClusterWide, "rbac.authorization.k8s.io/v1/clusterroles")
	oo, err := s.app.factory.List(render.ClusterWide, "rbac.authorization.k8s.io/v1/clusterrolebindings", labels.Everything())
	if err != nil {
		return nil, err
	}

	rows := make(render.Rows, 0, len(oo))
	for _, o := range oo {
		var crb rbacv1.ClusterRoleBinding
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &crb)
		if err != nil {
			return nil, err
		}
		for _, subject := range crb.Subjects {
			if subject.Kind != s.subjectKind {
				continue
			}
			rows = append(rows, render.Row{
				ID:     subject.Name,
				Fields: render.Fields{subject.Name, "ClusterRoleBinding", crb.Name},
			})
		}
	}

	return rows, nil
}

func (s *Subject) fetchRoleBindings() (render.Rows, error) {
	s.app.factory.Preload(render.ClusterWide, "rbac.authorization.k8s.io/v1/clusterroles")
	oo, err := s.app.factory.List(render.ClusterWide, "rbac.authorization.k8s.io/v1/rolebindings", labels.Everything())
	if err != nil {
		return nil, err
	}

	rows := make(render.Rows, 0, len(oo))
	for _, o := range oo {
		var rb rbacv1.RoleBinding
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &rb)
		if err != nil {
			return nil, err
		}
		for _, subject := range rb.Subjects {
			if subject.Kind == s.subjectKind {
				rows = append(rows, render.Row{
					ID:     subject.Name,
					Fields: render.Fields{subject.Name, "RoleBinding", rb.Name},
				})
			}
		}
	}

	return rows, nil
}

func mapCmdSubject(subject string) string {
	switch subject {
	case "groups":
		return group
	case "sas":
		return sa
	default:
		return user
	}
}

func mapFuSubject(subject string) (string, error) {
	switch subject {
	case group:
		return "g", nil
	case sa:
		return "s", nil
	case user:
		return "u", nil
	default:
		return "", fmt.Errorf("Unknown subject %q should be one of user, group, serviceaccount", subject)
	}
}
