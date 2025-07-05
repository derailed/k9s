// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/slogs"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/sahilm/fuzzy"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
)

func isBailoutEvt(evt *tcell.EventKey) bool {
	return evt.Name() == "Ctrl+C"
}

func aliases(m *v1.APIResource, aa sets.Set[string]) sets.Set[string] {
	ss := sets.New(aa.UnsortedList()...)
	ss.Insert(m.Name)
	ss.Insert(m.ShortNames...)
	if m.SingularName != "" {
		ss.Insert(m.SingularName)
	}

	return ss
}

func clipboardWrite(text string) error {
	if text != "" {
		return clipboard.WriteAll(text)
	}

	return nil
}

var bracketRX = regexp.MustCompile(`\[(.+)\[\]`)

func sanitizeEsc(s string) string {
	return bracketRX.ReplaceAllString(s, `[$1]`)
}

func cpCmd(flash *model.Flash, v *tview.TextView) func(*tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		if err := clipboardWrite(sanitizeEsc(v.GetText(true))); err != nil {
			flash.Err(err)
			return evt
		}
		flash.Info("Content copied to clipboard...")

		return nil
	}
}

func parsePFAnn(s string) (port, lport string, ok bool) {
	tokens := strings.Split(s, ":")
	if len(tokens) != 2 {
		return
	}

	return tokens[0], tokens[1], true
}

func k8sEnv(c *client.Config) Env {
	ctx, err := c.CurrentContextName()
	if err != nil {
		ctx = render.NAValue
	}
	cluster, err := c.CurrentClusterName()
	if err != nil {
		cluster = render.NAValue
	}
	user, err := c.CurrentUserName()
	if err != nil {
		user = render.NAValue
	}
	groups, err := c.CurrentGroupNames()
	if err != nil {
		groups = []string{render.NAValue}
	}

	var cfg string
	kcfg := c.Flags().KubeConfig
	if kcfg != nil && *kcfg != "" {
		cfg = *kcfg
	} else {
		cfg = os.Getenv("KUBECONFIG")
	}

	return Env{
		"CONTEXT":    ctx,
		"CLUSTER":    cluster,
		"USER":       user,
		"GROUPS":     strings.Join(groups, ","),
		"KUBECONFIG": cfg,
	}
}

func defaultEnv(c *client.Config, path string, header model1.Header, row *model1.Row) Env {
	env := k8sEnv(c)
	env["NAMESPACE"], env["NAME"] = client.Namespaced(path)
	if row == nil {
		return env
	}
	for _, col := range header.ColumnNames(true) {
		idx, ok := header.IndexOf(col, true)
		if ok && idx < len(row.Fields) {
			env["COL-"+col] = row.Fields[idx]
		}
	}

	return env
}

func describeResource(app *App, _ ui.Tabular, gvr *client.GVR, path string) {
	v := NewLiveView(app, "Describe", model.NewDescribe(gvr, path))
	if err := app.inject(v, false); err != nil {
		app.Flash().Err(err)
	}
}

func showReplicasets(app *App, path string, labelSel labels.Selector, fieldSel string) {
	v := NewReplicaSet(client.RsGVR)
	v.SetContextFn(func(ctx context.Context) context.Context {
		ctx = context.WithValue(ctx, internal.KeyPath, path)
		return context.WithValue(ctx, internal.KeyFields, fieldSel)
	})
	v.SetLabelSelector(labelSel)

	ns, _ := client.Namespaced(path)
	if err := app.Config.SetActiveNamespace(ns); err != nil {
		slog.Error("Unable to set active namespace during show replicasets", slogs.Error, err)
	}
	if err := app.inject(v, false); err != nil {
		app.Flash().Err(err)
	}
}

func showPods(app *App, path string, labelSel labels.Selector, fieldSel string) {
	v := NewPod(client.PodGVR)
	v.SetContextFn(podCtx(app, path, fieldSel))
	v.SetLabelSelector(labelSel)

	ns, _ := client.Namespaced(path)
	if err := app.Config.SetActiveNamespace(ns); err != nil {
		slog.Error("Unable to set active namespace during show pods", slogs.Error, err)
	}
	if err := app.inject(v, false); err != nil {
		app.Flash().Err(err)
	}
}

func podCtx(_ *App, path, fieldSel string) ContextFunc {
	return func(ctx context.Context) context.Context {
		ctx = context.WithValue(ctx, internal.KeyPath, path)
		return context.WithValue(ctx, internal.KeyFields, fieldSel)
	}
}

func extractApp(ctx context.Context) (*App, error) {
	app, ok := ctx.Value(internal.KeyApp).(*App)
	if !ok {
		return nil, errors.New("no application found in context")
	}

	return app, nil
}

// AsKey maps a string representation of a key to a tcell key.
func asKey(key string) (tcell.Key, error) {
	for k, v := range tcell.KeyNames {
		if key == v {
			return k, nil
		}
	}

	return 0, fmt.Errorf("invalid key specified: %q", key)
}

// FwFQN returns a fully qualified ns/name:container id.
func fwFQN(po, co string) string {
	return po + "|" + co
}

func isTCPPort(p string) bool {
	return !strings.Contains(p, "UDP")
}

// ContainerID computes container ID based on ns/po/co.
func containerID(path, co string) string {
	ns, n := client.Namespaced(path)
	po := strings.Split(n, "-")[0]

	return ns + "/" + po + ":" + co
}

// UrlFor computes fq url for a given benchmark configuration.
func urlFor(cfg *config.BenchConfig, port string) string {
	host := "localhost"
	if cfg.HTTP.Host != "" {
		host = cfg.HTTP.Host
	}

	path := "/"
	if cfg.HTTP.Path != "" {
		path = cfg.HTTP.Path
	}

	return "http://" + host + ":" + port + path
}

func fqn(ns, n string) string {
	if ns == "" {
		return n
	}
	return ns + "/" + n
}

func decorateCpuMemHeaderRows(app *App, data *model1.TableData) {
	for colIndex, header := range data.Header() {
		var check string
		if header.Name == "%CPU/L" {
			check = config.CPU
		}
		if header.Name == "%MEM/L" {
			check = config.MEM
		}
		if check == "" {
			continue
		}
		data.RowsRange(func(_ int, re model1.RowEvent) bool {
			if re.Row.Fields[colIndex] == render.NAValue {
				return true
			}
			n, err := strconv.Atoi(re.Row.Fields[colIndex])
			if err != nil {
				return true
			}
			if n > 100 {
				n = 100
			}
			severity := app.Config.K9s.Thresholds.LevelFor(check, n)
			if severity == config.SeverityLow {
				return true
			}
			color := app.Config.K9s.Thresholds.SeverityColor(check, n)
			if color != "" {
				re.Row.Fields[colIndex] = "[" + color + "::b]" + re.Row.Fields[colIndex]
			}

			return true
		})
	}
}

func matchTag(i int, s string) string {
	return `<<<"search_` + strconv.Itoa(i) + `">>>` + s + `<<<"">>>`
}

func linesWithRegions(lines []string, matches fuzzy.Matches) []string {
	ll := make([]string, len(lines))
	copy(ll, lines)
	offsetForLine := make(map[int]int)
	for i, m := range matches {
		for _, loc := range dao.ContinuousRanges(m.MatchedIndexes) {
			start, end := loc[0]+offsetForLine[m.Index], loc[1]+offsetForLine[m.Index]
			line := ll[m.Index]
			if end > len(line) {
				end = len(line)
			}
			regionStr := matchTag(i, line[start:end])
			ll[m.Index] = line[:start] + regionStr + line[end:]
			offsetForLine[m.Index] += len(regionStr) - (end - start)
		}
	}
	return ll
}
