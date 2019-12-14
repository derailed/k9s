package view

import (
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
)

type (
	TableInfo interface {
		Header() render.HeaderRow
		GetCache() render.RowEvents
		SetCache(render.RowEvents)
	}

	// Subject presents a user/group viewer.
	Subject struct {
		ResourceViewer

		subjectKind string
		cache       render.RowEvents
	}
)

// NewSubject returns a new subject viewer.
func NewSubject(gvr dao.GVR) ResourceViewer {
	s := Subject{ResourceViewer: NewBrowser(gvr)}
	s.GetTable().SetColorerFn(render.Subject{}.ColorerFunc())
	// s.GetTable().SetSortCol(1, len(s.Header()), true)
	s.SetBindKeysFn(s.bindKeys)

	return &s
}

// BOZO!!
// // Start runs the refresh loop.
// func (s *Subject) Start() {
// 	s.Stop()

// 	var ctx context.Context
// 	ctx, s.cancelFn = context.WithCancel(context.Background())
// 	go func(ctx context.Context) {
// 		for {
// 			select {
// 			case <-ctx.Done():
// 				log.Debug().Msgf("Subject:%s Watch bailing out!", s.subjectKind)
// 				return
// 			case <-time.After(time.Duration(s.App().Config.K9s.GetRefreshRate()) * time.Second):
// 				s.refresh()
// 			}
// 		}
// 	}(ctx)
// }

// Name returns the component name
func (s *Subject) Name() string {
	return "subjects"
}

func (s *Subject) bindKeys(aa ui.KeyActions) {
	aa.Delete(ui.KeyShiftA, ui.KeyShiftP, tcell.KeyCtrlSpace, ui.KeySpace)
	aa.Add(ui.KeyActions{
		tcell.KeyEnter: ui.NewKeyAction("Policies", s.policyCmd, true),
		// BOZO!!
		// tcell.KeyEscape: ui.NewKeyAction("Back", s.resetCmd, false),
		// ui.KeySlash:     ui.NewKeyAction("Filter", s.activateCmd, false),
		ui.KeyShiftK: ui.NewKeyAction("Sort Kind", s.GetTable().SortColCmd(1, true), false),
	})
}

// SetSubject sets the subject name.
func (s *Subject) SetSubject(n string) {
	s.subjectKind = mapSubject(n)
}

// BOZO!!
// func (s *Subject) refresh() {
// 	log.Debug().Msgf("Refreshing Subject...")
// 	data, err := s.reconcile()
// 	if err != nil {
// 		log.Error().Err(err).Msgf("Refresh for %s", s.subjectKind)
// 		s.App().Flash().Err(err)
// 	}
// 	s.App().QueueUpdateDraw(func() {
// 		s.GetTable().Update(data)
// 	})
// }

func (s *Subject) policyCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !s.GetTable().RowSelected() {
		return evt
	}

	// _, n := k8s.Namespaced(s.GetSelectedItem())
	// subject, err := mapFuSubject(s.subjectKind)
	// if err != nil {
	// 	s.App().Flash().Err(err)
	// 	return nil
	// }
	// BOZO!!
	// s.App().inject(NewPolicy(s.app, subject, n))

	return nil
}

// func (s *Subject) resetCmd(evt *tcell.EventKey) *tcell.EventKey {
// 	if !s.SearchBuff().Empty() {
// 		s.SearchBuff().Reset()
// 		return nil
// 	}

// 	return s.backCmd(evt)
// }

// func (s *Subject) backCmd(evt *tcell.EventKey) *tcell.EventKey {
// 	if s.SearchBuff().IsActive() {
// 		s.SearchBuff().Reset()
// 		return nil
// 	}

// 	return s.App().PrevCmd(evt)
// }

// func (s *Subject) reconcile() (render.TableData, error) {
// 	var table render.TableData
// 	if s.App().Conn() == nil {
// 		return table, nil
// 	}

// 	rows, err := s.fetchClusterRoleBindings()
// 	if err != nil {
// 		return table, err
// 	}

// 	nrows, err := s.fetchRoleBindings()
// 	if err != nil {
// 		return table, err
// 	}
// 	for k, v := range nrows {
// 		rows[k] = v
// 	}

// 	return buildTable(s, rows), nil
// }

// func (s *Subject) Header() render.HeaderRow {
// 	return render.Subject{}.Header(render.AllNamespaces)
// }

// func (s *Subject) GetCache() render.RowEvents {
// 	return s.cache
// }

// func (s *Subject) SetCache(rows render.RowEvents) {
// 	s.cache = rows
// }

// func buildTable(c TableInfo, rows render.Rows) render.TableData {
// 	table := render.TableData{
// 		Header:    c.Header(),
// 		Namespace: "*",
// 	}

// 	cache := c.GetCache()
// 	if len(cache) == 0 {
// 		cache := make(render.RowEvents, 0, len(rows))
// 		for _, row := range rows {
// 			cache = append(cache, render.RowEvent{Kind: render.EventAdd, Row: row})
// 		}
// 		table.RowEvents = cache
// 		return table
// 	}

// 	for _, row := range rows {
// 		idx, ok := cache.FindIndex(row.ID)
// 		if !ok {
// 			cache = append(cache, render.RowEvent{Kind: render.EventAdd, Row: row})
// 			continue
// 		}

// 		old := cache[idx].Row
// 		deltas := make(render.DeltaRow, len(row.Fields))
// 		if reflect.DeepEqual(old, row) {
// 			cache[idx].Kind = render.EventUnchanged
// 			cache[idx].Deltas = deltas
// 			continue
// 		}

// 		cache[idx].Kind = render.EventUpdate
// 		for i, field := range old.Fields {
// 			if field != row.Fields[i] {
// 				deltas[i] = field
// 			}
// 		}
// 		cache[idx].Deltas = deltas
// 	}

// 	for _, row := range rows {
// 		if _, ok := cache.FindIndex(row.ID); !ok {
// 			cache.Delete(row.ID)
// 		}
// 	}
// 	table.RowEvents = cache

// 	return table
// }

// func (s *Subject) fetchClusterRoleBindings() (render.Rows, error) {
// 	s.App().factory.Preload(render.ClusterWide, "rbac.authorization.k8s.io/v1/clusterroles")
// 	oo, err := s.App().factory.List(render.ClusterWide, "rbac.authorization.k8s.io/v1/clusterrolebindings", labels.Everything())
// 	if err != nil {
// 		return nil, err
// 	}

// 	rows := make(render.Rows, 0, len(oo))
// 	for _, o := range oo {
// 		var crb rbacv1.ClusterRoleBinding
// 		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &crb)
// 		if err != nil {
// 			return nil, err
// 		}
// 		for _, subject := range crb.Subjects {
// 			if subject.Kind != s.subjectKind {
// 				continue
// 			}
// 			rows = append(rows, render.Row{
// 				ID:     subject.Name,
// 				Fields: render.Fields{subject.Name, "ClusterRoleBinding", crb.Name},
// 			})
// 		}
// 	}

// 	return rows, nil
// }

// func (s *Subject) fetchRoleBindings() (render.Rows, error) {
// 	s.App().factory.Preload(render.ClusterWide, "rbac.authorization.k8s.io/v1/clusterroles")
// 	oo, err := s.App().factory.List(render.ClusterWide, "rbac.authorization.k8s.io/v1/rolebindings", labels.Everything())
// 	if err != nil {
// 		return nil, err
// 	}

// 	rows := make(render.Rows, 0, len(oo))
// 	for _, o := range oo {
// 		var rb rbacv1.RoleBinding
// 		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &rb)
// 		if err != nil {
// 			return nil, err
// 		}
// 		for _, subject := range rb.Subjects {
// 			if subject.Kind == s.subjectKind {
// 				rows = append(rows, render.Row{
// 					ID:     subject.Name,
// 					Fields: render.Fields{subject.Name, "RoleBinding", rb.Name},
// 				})
// 			}
// 		}
// 	}

// 	return rows, nil
// }

// func mapCmdSubject(subject string) string {
// 	switch subject {
// 	case "groups":
// 		return group
// 	case "sas":
// 		return sa
// 	default:
// 		return user
// 	}
// }

// func mapFuSubject(subject string) (string, error) {
// 	switch subject {
// 	case group:
// 		return "g", nil
// 	case sa:
// 		return "s", nil
// 	case user:
// 		return "u", nil
// 	default:
// 		return "", fmt.Errorf("Unknown subject %q should be one of user, group, serviceaccount", subject)
// 	}
// }
