package view

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/printers"
)

func defaultK9sEnv(app *App, sel string, row render.Row) K9sEnv {
	ns, n := client.Namespaced(sel)
	ctx, err := app.Conn().Config().CurrentContextName()
	if err != nil {
		ctx = render.NAValue
	}
	cluster, err := app.Conn().Config().CurrentClusterName()
	if err != nil {
		cluster = render.NAValue
	}
	user, err := app.Conn().Config().CurrentUserName()
	if err != nil {
		user = render.NAValue
	}
	groups, err := app.Conn().Config().CurrentGroupNames()
	if err != nil {
		groups = []string{render.NAValue}
	}
	var cfg string
	kcfg := app.Conn().Config().Flags().KubeConfig
	if kcfg != nil && *kcfg != "" {
		cfg = *kcfg
	}

	env := K9sEnv{
		"NAMESPACE":  ns,
		"NAME":       n,
		"CONTEXT":    ctx,
		"CLUSTER":    cluster,
		"USER":       user,
		"GROUPS":     strings.Join(groups, ","),
		"KUBECONFIG": cfg,
	}

	for i, r := range row.Fields {
		env["COL"+strconv.Itoa(i)] = r
	}

	return env
}

func describeResource(app *App, _, gvr, path string) {
	ns, n := client.Namespaced(path)
	yaml, err := dao.Describe(app.Conn(), client.NewGVR(gvr), ns, n)
	if err != nil {
		app.Flash().Errf("Describe command failed: %s", err)
		return
	}

	details := NewDetails(app, "Describe", path).Update(yaml)
	if err := app.inject(details); err != nil {
		app.Flash().Err(err)
	}
}

func toYAML(o runtime.Object) (string, error) {
	var (
		buff bytes.Buffer
		p    printers.YAMLPrinter
	)
	err := p.PrintObj(o, &buff)
	if err != nil {
		log.Error().Msgf("Marshal Error %v", err)
		return "", err
	}

	return buff.String(), nil
}

func showPodsWithLabels(app *App, path string, sel map[string]string) {
	var labels []string
	for k, v := range sel {
		labels = append(labels, fmt.Sprintf("%s=%s", k, v))
	}
	showPods(app, path, strings.Join(labels, ","), "")
}

func showPods(app *App, path, labelSel, fieldSel string) {
	log.Debug().Msgf("SHOW PODS %q -- %q -- %q", path, labelSel, fieldSel)
	app.switchNS("")

	v := NewPod(client.NewGVR("v1/pods"))
	v.SetContextFn(podCtx(app, path, labelSel, fieldSel))
	v.GetTable().SetColorerFn(render.Pod{}.ColorerFunc())

	ns, _ := client.Namespaced(path)
	if err := app.Config.SetActiveNamespace(ns); err != nil {
		log.Error().Err(err).Msg("Config NS set failed!")
	}
	if err := app.inject(v); err != nil {
		app.Flash().Err(err)
	}
}

func podCtx(app *App, path, labelSel, fieldSel string) ContextFunc {
	return func(ctx context.Context) context.Context {
		ctx = context.WithValue(ctx, internal.KeyPath, path)
		ctx = context.WithValue(ctx, internal.KeyLabels, labelSel)

		ns, _ := client.Namespaced(path)
		log.Debug().Msgf("POD METRICS in NS %q", ns)
		mx := client.NewMetricsServer(app.factory.Client())
		nmx, err := mx.FetchPodsMetrics(ns)
		if err != nil {
			log.Warn().Err(err).Msgf("No pods metrics")
		}
		ctx = context.WithValue(ctx, internal.KeyMetrics, nmx)

		return context.WithValue(ctx, internal.KeyFields, fieldSel)
	}
}

func extractApp(ctx context.Context) (*App, error) {
	app, ok := ctx.Value(internal.KeyApp).(*App)
	if !ok {
		return nil, errors.New("No application found in context")
	}

	return app, nil
}

// AsKey maps a string representation of a key to a tcell key.
func asKey(key string) (tcell.Key, error) {
	for k, v := range tcell.KeyNames {
		if v == key {
			return k, nil
		}
	}

	return 0, fmt.Errorf("No matching key found %s", key)
}

// FwFQN returns a fully qualified ns/name:container id.
func fwFQN(po, co string) string {
	return po + ":" + co
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
