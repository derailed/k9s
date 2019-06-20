package views

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/resource"
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

	subjectView struct {
		*tableView

		current      igniter
		cancel       context.CancelFunc
		subjectKind  string
		selectedItem string
		cache        resource.RowEvents
	}
)

func newSubjectView(ns string, app *appView, list resource.List) resourceViewer {
	v := subjectView{}
	{
		v.tableView = newTableView(app, v.getTitle())
		v.tableView.SetSelectionChangedFunc(v.selChanged)
		v.colorerFn = rbacColorer
		v.bindKeys()
	}

	if current, ok := app.content.GetPrimitive("main").(igniter); ok {
		v.current = current
	} else {
		v.current = &v
	}

	return &v
}

// Init the view.
func (v *subjectView) init(c context.Context, _ string) {
	if v.cancel != nil {
		v.cancel()
	}

	v.sortCol = sortColumn{1, len(rbacHeader), true}
	v.subjectKind = mapCmdSubject(v.app.config.K9s.ActiveCluster().View.Active)
	v.baseTitle = v.getTitle()

	ctx, cancel := context.WithCancel(c)
	v.cancel = cancel
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				log.Debug().Msgf("Subject:%s Watch bailing out!", v.subjectKind)
				return
			case <-time.After(time.Duration(v.app.config.K9s.RefreshRate) * time.Second):
				v.refresh()
				v.app.Draw()
			}
		}
	}(ctx)

	v.refresh()
	v.app.SetFocus(v)
}

func (v *subjectView) setExtraActionsFn(f actionsFn) {}
func (v *subjectView) setColorerFn(f colorerFn)      {}
func (v *subjectView) setEnterFn(f enterFn)          {}
func (v *subjectView) setDecorateFn(f decorateFn)    {}

func (v *subjectView) bindKeys() {
	// No time data or ns
	delete(v.actions, KeyShiftA)
	delete(v.actions, KeyShiftP)

	v.actions[tcell.KeyEnter] = newKeyAction("RBAC", v.rbackCmd, true)
	v.actions[tcell.KeyEscape] = newKeyAction("Reset", v.resetCmd, false)
	v.actions[KeySlash] = newKeyAction("Filter", v.activateCmd, false)
	v.actions[KeyP] = newKeyAction("Previous", v.app.prevCmd, false)
	v.actions[KeyShiftK] = newKeyAction("Sort Kind", v.sortColCmd(1), true)
}

func (v *subjectView) getTitle() string {
	return fmt.Sprintf(rbacTitleFmt, "Subject", v.subjectKind)
}

func (v *subjectView) selChanged(r, _ int) {
	if r == 0 {
		v.selectedItem = ""
		return
	}
	v.selectedItem = strings.TrimSpace(v.GetCell(r, 0).Text)
}

func (v *subjectView) SetSubject(s string) {
	v.subjectKind = mapSubject(s)
}

func (v *subjectView) refresh() {
	data, err := v.reconcile()
	if err != nil {
		log.Error().Err(err).Msgf("Unable to reconcile for %s", v.subjectKind)
	}
	v.update(data)
}

func (v *subjectView) rbackCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.selectedItem == "" {
		return evt
	}

	if v.cancel != nil {
		v.cancel()
	}

	_, n := namespaced(v.selectedItem)
	v.app.inject(newPolicyView(v.app, mapFuSubject(v.subjectKind), n))

	return nil
}

func (v *subjectView) resetCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.cmdBuff.empty() {
		v.cmdBuff.reset()
		return nil
	}

	return v.backCmd(evt)
}

func (v *subjectView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.cancel != nil {
		v.cancel()
	}

	if v.cmdBuff.isActive() {
		v.cmdBuff.reset()
		return nil
	}

	v.app.inject(v.current)

	return nil
}

func (v *subjectView) hints() hints {
	return v.actions.toHints()
}

func (v *subjectView) reconcile() (resource.TableData, error) {
	var table resource.TableData

	evts, err := v.clusterSubjects()
	if err != nil {
		return table, err
	}

	nevts, err := v.namespacedSubjects()
	if err != nil {
		return table, err
	}
	for k, v := range nevts {
		evts[k] = v
	}

	return buildTable(v, evts), nil
}

func (v *subjectView) header() resource.Row {
	return subjectHeader
}

func (v *subjectView) getCache() resource.RowEvents {
	return v.cache
}

func (v *subjectView) setCache(evts resource.RowEvents) {
	v.cache = evts
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

func (v *subjectView) clusterSubjects() (resource.RowEvents, error) {
	crbs, err := v.app.conn().DialOrDie().Rbac().ClusterRoleBindings().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	evts := make(resource.RowEvents, len(crbs.Items))
	for _, crb := range crbs.Items {
		for _, s := range crb.Subjects {
			if s.Kind != v.subjectKind {
				continue
			}
			evts[s.Name] = &resource.RowEvent{
				Fields: resource.Row{s.Name, "ClusterRoleBinding", crb.Name},
			}
		}
	}

	return evts, nil
}

func (v *subjectView) namespacedSubjects() (resource.RowEvents, error) {
	rbs, err := v.app.conn().DialOrDie().Rbac().RoleBindings("").List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	evts := make(resource.RowEvents, len(rbs.Items))
	for _, rb := range rbs.Items {
		for _, s := range rb.Subjects {
			if s.Kind == v.subjectKind {
				evts[s.Name] = &resource.RowEvent{
					Fields: resource.Row{s.Name, "RoleBinding", rb.Name},
				}
			}
		}
	}

	return evts, nil
}

func mapCmdSubject(subject string) string {
	switch subject {
	case "grp":
		return "Group"
	case "sas":
		return "ServiceAccount"
	default:
		return "User"
	}
}

func mapFuSubject(subject string) string {
	switch subject {
	case "Group":
		return "g"
	case "ServiceAccount":
		return "s"
	default:
		return "u"
	}
}
