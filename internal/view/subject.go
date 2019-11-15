package view

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

var subjectHeader = resource.Row{"NAME", "KIND", "FIRST LOCATION"}

type (
	cachedEventer interface {
		header() resource.Row
		getCache() resource.RowEvents
		setCache(resource.RowEvents)
	}

	// Subject presents a user/group viewer.
	Subject struct {
		*Table

		cancel      context.CancelFunc
		subjectKind string
		cache       resource.RowEvents
	}
)

// NewSubject returns a new subject viewer.
func NewSubject(title, gvr string, list resource.List) ResourceViewer {
	return &Subject{
		Table: NewTable("Subject"),
	}
}

// Init initializes the view.
func (s *Subject) Init(ctx context.Context) {
	s.ActiveNS = "*"
	s.SetColorerFn(rbacColorer)
	s.Table.Init(ctx)
	s.bindKeys()
	s.SetSortCol(1, len(rbacHeader), true)
	s.subjectKind = mapCmdSubject(s.app.Config.K9s.ActiveCluster().View.Active)
	s.SetBaseTitle(s.subjectKind)
	s.SelectRow(1, true)

	s.refresh()
}

func (s *Subject) Start() {
	s.Stop()

	var ctx context.Context
	ctx, s.cancel = context.WithCancel(context.Background())
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				log.Debug().Msgf("Subject:%s Watch bailing out!", s.subjectKind)
				return
			case <-time.After(time.Duration(s.app.Config.K9s.GetRefreshRate()) * time.Second):
				s.app.QueueUpdateDraw(func() {
					s.refresh()
				})
			}
		}
	}(ctx)
}

func (s *Subject) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *Subject) Name() string {
	return "subject"
}

func (s *Subject) masterPage() *Table {
	return s.Table
}

func (s *Subject) bindKeys() {
	s.RmAction(ui.KeyShiftA)
	s.RmAction(ui.KeyShiftP)
	s.RmAction(tcell.KeyCtrlSpace)
	s.RmAction(ui.KeySpace)

	s.AddActions(ui.KeyActions{
		tcell.KeyEnter:  ui.NewKeyAction("Policies", s.policyCmd, true),
		tcell.KeyEscape: ui.NewKeyAction("Back", s.resetCmd, false),
		ui.KeySlash:     ui.NewKeyAction("Filter", s.activateCmd, false),
		ui.KeyShiftK:    ui.NewKeyAction("Sort Kind", s.SortColCmd(1), false),
	})
}

func (s *Subject) setExtraActionsFn(f ActionsFunc) {}
func (s *Subject) setColorerFn(f ui.ColorerFunc)   {}
func (s *Subject) setEnterFn(f enterFn)            {}
func (s *Subject) setDecorateFn(f decorateFn)      {}

func (s *Subject) SetSubject(n string) {
	s.subjectKind = mapSubject(n)
}

func (s *Subject) refresh() {
	data, err := s.reconcile()
	if err != nil {
		log.Error().Err(err).Msgf("Refresh for %s", s.subjectKind)
		s.app.Flash().Err(err)
	}
	s.Update(data)
}

func (s *Subject) policyCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !s.RowSelected() {
		return evt
	}

	if s.cancel != nil {
		s.cancel()
	}

	_, n := namespaced(s.GetSelectedItem())
	s.app.inject(NewPolicy(s.app, mapFuSubject(s.subjectKind), n))

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
	if s.cancel != nil {
		s.cancel()
	}

	if s.SearchBuff().IsActive() {
		s.SearchBuff().Reset()
		return nil
	}

	return s.app.PrevCmd(evt)
}

func (s *Subject) reconcile() (resource.TableData, error) {
	var table resource.TableData
	if s.app.Conn() == nil {
		return table, nil
	}

	evts, err := s.clusterSubjects()
	if err != nil {
		return table, err
	}

	nevts, err := s.namespacedSubjects()
	if err != nil {
		return table, err
	}
	for k, v := range nevts {
		evts[k] = v
	}

	return buildTable(s, evts), nil
}

func (s *Subject) header() resource.Row {
	return subjectHeader
}

func (s *Subject) getCache() resource.RowEvents {
	return s.cache
}

func (s *Subject) setCache(evts resource.RowEvents) {
	s.cache = evts
}

func buildTable(c cachedEventer, evts resource.RowEvents) resource.TableData {
	table := resource.TableData{
		Header:    c.header(),
		Rows:      make(resource.RowEvents, len(evts)),
		Namespace: "*",
	}

	noDeltas := make(resource.Row, len(c.header()))
	cache := c.getCache()
	if len(cache) == 0 {
		for k, ev := range evts {
			ev.Action = resource.New
			ev.Deltas = noDeltas
			table.Rows[k] = ev
		}
		c.setCache(evts)
		return table
	}

	for k, ev := range evts {
		table.Rows[k] = ev

		newr := ev.Fields
		if _, ok := cache[k]; !ok {
			ev.Action, ev.Deltas = watch.Added, noDeltas
			continue
		}
		oldr := cache[k].Fields
		deltas := make(resource.Row, len(newr))
		if !reflect.DeepEqual(oldr, newr) {
			ev.Action = watch.Modified
			for i, field := range oldr {
				if field != newr[i] {
					deltas[i] = field
				}
			}
			ev.Deltas = deltas
		} else {
			ev.Action = resource.Unchanged
			ev.Deltas = noDeltas
		}
	}

	for k := range evts {
		if _, ok := table.Rows[k]; !ok {
			delete(evts, k)
		}
	}
	c.setCache(evts)

	return table
}

func (s *Subject) clusterSubjects() (resource.RowEvents, error) {
	crbs, err := s.app.Conn().DialOrDie().RbacV1().ClusterRoleBindings().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	evts := make(resource.RowEvents, len(crbs.Items))
	for _, crb := range crbs.Items {
		for _, subject := range crb.Subjects {
			if subject.Kind != s.subjectKind {
				continue
			}
			evts[subject.Name] = &resource.RowEvent{
				Fields: resource.Row{subject.Name, "ClusterRoleBinding", crb.Name},
			}
		}
	}

	return evts, nil
}

func (s *Subject) namespacedSubjects() (resource.RowEvents, error) {
	rbs, err := s.app.Conn().DialOrDie().RbacV1().RoleBindings("").List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	evts := make(resource.RowEvents, len(rbs.Items))
	for _, rb := range rbs.Items {
		for _, subject := range rb.Subjects {
			if subject.Kind == s.subjectKind {
				evts[subject.Name] = &resource.RowEvent{
					Fields: resource.Row{subject.Name, "RoleBinding", rb.Name},
				}
			}
		}
	}

	return evts, nil
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

func mapFuSubject(subject string) string {
	switch subject {
	case group:
		return "g"
	case sa:
		return "s"
	case user:
		return "u"
	default:
		panic(fmt.Sprintf("Unknown FU subject %q", subject))
	}
}
