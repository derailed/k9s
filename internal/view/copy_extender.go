// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/slogs"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const copyKey = "copyResource"

var errNoTargetNS = errors.New("a target namespace is required")

// CopyExtender provides for copying resources to another namespace.
type CopyExtender struct {
	ResourceViewer
}

// NewCopyExtender returns a new extender.
func NewCopyExtender(r ResourceViewer) ResourceViewer {
	c := CopyExtender{ResourceViewer: r}
	c.AddBindKeysFn(c.bindKeys)

	return &c
}

func (c *CopyExtender) bindKeys(aa *ui.KeyActions) {
	if c.App().Config.IsReadOnly() || !c.isCopyable() {
		return
	}
	aa.Add(ui.KeyShiftC, ui.NewKeyAction("Copy To Ns", c.copyCmd, true))
}

// isCopyable checks that the resource is namespaced and supports creation. This
// weeds out cluster scoped resources as well as the k9s pseudo resources which
// carry no create verb. NOTE: client.Can is deliberately not used here as it
// treats nil verbs as permissive, which would opt in the pseudo resources.
func (c *CopyExtender) isCopyable() bool {
	m, err := dao.MetaAccess.MetaFor(c.GVR())
	if err != nil {
		return false
	}

	return m.Namespaced && slices.Contains(m.Verbs, client.CreateVerb)
}

func (c *CopyExtender) copyCmd(evt *tcell.EventKey) *tcell.EventKey {
	paths := c.GetTable().GetSelectedItems()
	if len(paths) == 0 {
		return evt
	}

	c.Stop()
	defer c.Start()
	c.showCopyDialog(paths)

	return nil
}

func (c *CopyExtender) showCopyDialog(paths []string) {
	confirm := tview.NewModalForm("<Copy>", c.makeCopyForm(paths))
	if len(paths) == 1 {
		confirm.SetText(fmt.Sprintf("Copy %s %s to", c.GVR(), paths[0]))
	} else {
		confirm.SetText(fmt.Sprintf("Copy %d %s resources to", len(paths), c.GVR()))
	}
	confirm.SetDoneFunc(func(int, string) {
		c.dismissDialog()
	})
	c.App().Content.AddPage(copyKey, confirm, false, false)
	c.App().Content.ShowPage(copyKey)
}

func (c *CopyExtender) makeCopyForm(paths []string) *tview.Form {
	srcNS, srcName := client.Namespaced(paths[0])
	opts := dao.CopyOpts{Namespace: srcNS}
	if len(paths) == 1 {
		opts.Name = srcName
	}

	styles := c.App().Styles.Dialog()
	f := tview.NewForm().
		SetItemPadding(0).
		SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(styles.ButtonBgColor.Color()).
		SetButtonTextColor(styles.ButtonFgColor.Color()).
		SetLabelColor(styles.LabelFgColor.Color()).
		SetFieldTextColor(styles.FieldFgColor.Color())

	f.AddInputField("Namespace:", opts.Namespace, 40, nil, func(v string) {
		opts.Namespace = strings.TrimSpace(v)
	})
	if field, ok := f.GetFormItemByLabel("Namespace:").(*tview.InputField); ok {
		nn := c.namespaces()
		field.SetAutocompleteFunc(func(s string) []string {
			return matchNamespaces(nn, s)
		})
	}

	// A rename only makes sense for a single resource. On a multi selection all
	// copies keep their source name.
	if len(paths) == 1 {
		f.AddInputField("Name:", opts.Name, 40, nil, func(v string) {
			opts.Name = strings.TrimSpace(v)
		})
	}

	f.AddCheckbox("Overwrite:", opts.Overwrite, func(_ string, flag bool) {
		opts.Overwrite = flag
	})

	// Hand editing only makes sense for a single resource.
	var editSpec bool
	if len(paths) == 1 {
		f.AddCheckbox("Edit spec:", editSpec, func(_ string, flag bool) {
			editSpec = flag
		})
	}

	f.AddButton("OK", func() {
		c.dismissDialog()
		o := opts
		if editSpec {
			// The editor suspends the UI so this has to stay on the event loop.
			c.runCopyEdit(paths[0], &o)
			return
		}
		// Each copy is a round trip to the api-server, so keep a multi selection
		// off the event loop. Flash is channel based and safe from here.
		go c.runCopy(paths, &o)
	})
	f.AddButton("Cancel", func() {
		c.dismissDialog()
	})

	for i := range f.GetButtonCount() {
		f.GetButton(i).
			SetBackgroundColorActivated(styles.ButtonFocusBgColor.Color()).
			SetLabelColorActivated(styles.ButtonFocusFgColor.Color())
	}

	return f
}

func (c *CopyExtender) dismissDialog() {
	c.App().Content.RemovePage(copyKey)
}

func (c *CopyExtender) copier() (dao.Copier, error) {
	res, err := dao.AccessorFor(c.App().factory, c.GVR())
	if err != nil {
		return nil, err
	}
	cp, ok := res.(dao.Copier)
	if !ok {
		return nil, fmt.Errorf("resource %s cannot be copied", c.GVR())
	}

	return cp, nil
}

// runCopyEdit pops the user editor on the sanitized copy before creating it.
func (c *CopyExtender) runCopyEdit(path string, opts *dao.CopyOpts) {
	if opts.Namespace == "" {
		c.App().Flash().Err(errNoTargetNS)
		return
	}
	cp, err := c.copier()
	if err != nil {
		c.App().Flash().Err(err)
		return
	}

	u, err := c.prepCopy(cp, path, opts)
	if err != nil {
		c.App().Flash().Err(err)
		return
	}
	// NOTE: not on a timeout as the user owns the editor session.
	if u, err = c.editObj(u); err != nil {
		c.App().Flash().Err(err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.App().Conn().Config().CallTimeout())
	defer cancel()
	fqn, err := cp.CopyCreate(ctx, u, opts.Overwrite)
	if err != nil {
		slog.Error("Unable to copy resource",
			slogs.GVR, c.GVR(),
			slogs.FQN, path,
			slogs.Namespace, opts.Namespace,
			slogs.Error, err,
		)
		c.App().Flash().Err(err)
		return
	}
	c.App().Flash().Infof("Copied %s to %s", c.GVR(), fqn)
}

func (c *CopyExtender) prepCopy(cp dao.Copier, path string, opts *dao.CopyOpts) (*unstructured.Unstructured, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.App().Conn().Config().CallTimeout())
	defer cancel()

	return cp.CopyPrep(ctx, path, opts)
}

// editObj round trips a resource through the user editor.
func (c *CopyExtender) editObj(u *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	raw, err := dao.ToYAML(u, true)
	if err != nil {
		return nil, err
	}

	f, err := os.CreateTemp("", "k9s-copy-*.yaml")
	if err != nil {
		return nil, err
	}
	defer func() {
		if e := os.Remove(f.Name()); e != nil {
			slog.Warn("Unable to remove copy scratch file",
				slogs.Path, f.Name(),
				slogs.Error, e,
			)
		}
	}()
	_, err = f.WriteString(raw)
	if e := f.Close(); err == nil {
		err = e
	}
	if err != nil {
		return nil, err
	}

	if !edit(c.App(), &shellOpts{clear: true, args: []string{f.Name()}}) {
		return nil, errors.New("copy aborted: unable to launch editor")
	}

	bb, err := os.ReadFile(f.Name())
	if err != nil {
		return nil, err
	}
	jj, err := yaml.ToJSON(bb)
	if err != nil {
		return nil, fmt.Errorf("invalid yaml: %w", err)
	}
	var out unstructured.Unstructured
	if err := out.UnmarshalJSON(jj); err != nil {
		return nil, fmt.Errorf("invalid resource: %w", err)
	}

	return &out, nil
}

func (c *CopyExtender) runCopy(paths []string, opts *dao.CopyOpts) {
	if opts.Namespace == "" {
		c.App().Flash().Err(errNoTargetNS)
		return
	}

	cp, err := c.copier()
	if err != nil {
		c.App().Flash().Err(err)
		return
	}

	var (
		fqn    string
		copied int
	)
	for _, path := range paths {
		o := *opts
		if len(paths) > 1 {
			o.Name = ""
		}
		if fqn, err = c.copyOne(cp, path, &o); err != nil {
			slog.Error("Unable to copy resource",
				slogs.GVR, c.GVR(),
				slogs.FQN, path,
				slogs.Namespace, o.Namespace,
				slogs.Error, err,
			)
			c.App().Flash().Err(err)
			continue
		}
		copied++
	}

	switch copied {
	case 0:
		return
	case 1:
		c.App().Flash().Infof("Copied %s to %s", c.GVR(), fqn)
	default:
		c.App().Flash().Infof("Copied %d %s resources to namespace %s", copied, c.GVR(), opts.Namespace)
	}
}

func (c *CopyExtender) copyOne(cp dao.Copier, path string, opts *dao.CopyOpts) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.App().Conn().Config().CallTimeout())
	defer cancel()

	return cp.Copy(ctx, path, opts)
}

// namespaces returns the cluster namespaces for autocompletion. Falls back on
// the configured favorites when the user cannot list namespaces.
func (c *CopyExtender) namespaces() []string {
	oo, err := c.App().factory.List(client.NsGVR, client.ClusterScope, false, labels.Everything())
	if err != nil {
		slog.Warn("Unable to list namespaces. Using favorites",
			slogs.Error, err,
		)
		return c.App().Config.FavNamespaces()
	}

	nn := make([]string, 0, len(oo))
	for _, o := range oo {
		u, ok := o.(*unstructured.Unstructured)
		if !ok {
			continue
		}
		nn = append(nn, u.GetName())
	}
	sort.Strings(nn)

	return nn
}

func matchNamespaces(nn []string, s string) []string {
	if s == "" {
		return nil
	}

	mm := make([]string, 0, len(nn))
	for _, ns := range nn {
		if strings.HasPrefix(ns, s) && ns != s {
			mm = append(mm, ns)
		}
	}

	return mm
}
