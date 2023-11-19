// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/rs/zerolog/log"
)

func clipboardWrite(text string) error {
	return clipboard.WriteAll(text)
}

func cpCmd(flash *model.Flash, v *tview.TextView) func(*tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		if err := clipboardWrite(v.GetText(true)); err != nil {
			flash.Err(err)
			return evt
		}
		flash.Info("Content copied to clipboard...")

		return nil
	}
}

func parsePFAnn(s string) (string, string, bool) {
	tokens := strings.Split(s, ":")
	if len(tokens) != 2 {
		return "", "", false
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
	}

	return Env{
		"CONTEXT":    ctx,
		"CLUSTER":    cluster,
		"USER":       user,
		"GROUPS":     strings.Join(groups, ","),
		"KUBECONFIG": cfg,
	}
}

func defaultEnv(c *client.Config, path string, header render.Header, row render.Row) Env {
	env := k8sEnv(c)
	env["NAMESPACE"], env["NAME"] = client.Namespaced(path)
	for _, col := range header.Columns(true) {
		i := header.IndexOf(col, true)
		if i >= 0 && i < len(row.Fields) {
			env["COL-"+col] = row.Fields[i]
		}
	}

	return env
}

func describeResource(app *App, m ui.Tabular, gvr, path string) {
	v := NewLiveView(app, "Describe", model.NewDescribe(client.NewGVR(gvr), path))
	if err := app.inject(v, false); err != nil {
		app.Flash().Err(err)
	}
}

func showPodsWithLabels(app *App, path string, sel map[string]string) {
	labels := make([]string, 0, len(sel))
	for k, v := range sel {
		labels = append(labels, fmt.Sprintf("%s=%s", k, v))
	}
	showPods(app, path, strings.Join(labels, ","), "")
}

func showPods(app *App, path, labelSel, fieldSel string) {
	if err := app.switchNS(client.AllNamespaces); err != nil {
		app.Flash().Err(err)
		return
	}

	v := NewPod(client.NewGVR("v1/pods"))
	v.SetContextFn(podCtx(app, path, labelSel, fieldSel))

	ns, _ := client.Namespaced(path)
	if err := app.Config.SetActiveNamespace(ns); err != nil {
		log.Error().Err(err).Msg("Config NS set failed!")
	}
	if err := app.inject(v, false); err != nil {
		app.Flash().Err(err)
	}
}

func podCtx(app *App, path, labelSel, fieldSel string) ContextFunc {
	return func(ctx context.Context) context.Context {
		ctx = context.WithValue(ctx, internal.KeyPath, path)
		ctx = context.WithValue(ctx, internal.KeyLabels, labelSel)

		ns, _ := client.Namespaced(path)
		mx := client.NewMetricsServer(app.factory.Client())
		nmx, err := mx.FetchPodsMetrics(ctx, ns)
		if err != nil {
			log.Debug().Err(err).Msgf("No pods metrics")
		}
		ctx = context.WithValue(ctx, internal.KeyMetrics, nmx)

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

	return 0, fmt.Errorf("no matching key found %s", key)
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
func urlFor(cfg config.BenchConfig, port string) string {
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

func decorateCpuMemHeaderRows(app *App, data *render.TableData) {
	for colIndex, header := range data.Header {
		var check string
		if header.Name == "%CPU/L" {
			check = "cpu"
		}
		if header.Name == "%MEM/L" {
			check = "memory"
		}
		if len(check) == 0 {
			continue
		}
		for _, re := range data.RowEvents {
			if re.Row.Fields[colIndex] == render.NAValue {
				continue
			}
			n, err := strconv.Atoi(re.Row.Fields[colIndex])
			if err != nil {
				continue
			}
			if n > 100 {
				n = 100
			}
			severity := app.Config.K9s.Thresholds.LevelFor(check, n)
			if severity == config.SeverityLow {
				continue
			}
			color := app.Config.K9s.Thresholds.SeverityColor(check, n)
			if len(color) > 0 {
				re.Row.Fields[colIndex] = "[" + color + "::b]" + re.Row.Fields[colIndex]
			}
		}
	}
}
