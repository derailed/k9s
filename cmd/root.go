// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/color"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/config/data"
	"github.com/derailed/k9s/internal/slogs"
	"github.com/derailed/k9s/internal/view"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-colorable"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/clientcmd/api"
)

const (
	appName      = config.AppName
	shortAppDesc = "A graphical CLI for your Kubernetes cluster management."
	longAppDesc  = "K9s is a CLI to view and manage your Kubernetes clusters."
)

var _ data.KubeSettings = (*client.Config)(nil)

var (
	version, commit, date = "dev", "dev", client.NA
	k9sFlags              *config.Flags
	k8sFlags              *genericclioptions.ConfigFlags

	rootCmd = &cobra.Command{
		Use:   appName,
		Short: shortAppDesc,
		Long:  longAppDesc,
		RunE:  run,
	}

	out = colorable.NewColorableStdout()
)

type flagError struct{ err error }

func (e flagError) Error() string { return e.err.Error() }

func init() {
	if err := config.InitLogLoc(); err != nil {
		fmt.Printf("Fail to init k9s logs location %s\n", err)
	}

	rootCmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return flagError{err: err}
	})

	rootCmd.AddCommand(versionCmd(), infoCmd())
	initK9sFlags()
	initK8sFlags()
}

// Execute root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(*cobra.Command, []string) error {
	if err := config.InitLocs(); err != nil {
		return err
	}
	logFile, err := os.OpenFile(
		*k9sFlags.LogFile,
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		data.DefaultFileMod,
	)
	if err != nil {
		return fmt.Errorf("log file %q init failed: %w", *k9sFlags.LogFile, err)
	}
	defer func() {
		if logFile != nil {
			_ = logFile.Close()
		}
	}()
	defer func() {
		if err := recover(); err != nil {
			slog.Error("Boom!! k9s init failed", slogs.Error, err)
			slog.Error("", slogs.Stack, string(debug.Stack()))
			printLogo(color.Red)
			fmt.Printf("%s", color.Colorize("Boom!! ", color.Red))
			fmt.Printf("%v.\n", err)
		}
	}()

	slog.SetDefault(slog.New(tint.NewHandler(logFile, &tint.Options{
		Level:      parseLevel(*k9sFlags.LogLevel),
		TimeFormat: time.RFC3339,
	})))

	cfg, err := loadConfiguration()
	if err != nil {
		slog.Warn("Fail to load global/context configuration", slogs.Error, err)
	}
	app := view.NewApp(cfg)
	if app.Config.K9s.DefaultView != "" {
		app.Config.SetActiveView(app.Config.K9s.DefaultView)
	}

	if err := app.Init(version, int(*k9sFlags.RefreshRate)); err != nil {
		return err
	}
	if err := app.Run(); err != nil {
		return err
	}
	if view.ExitStatus != "" {
		return fmt.Errorf("view exit status %s", view.ExitStatus)
	}

	return nil
}

func loadConfiguration() (*config.Config, error) {
	slog.Info("🐶 K9s starting up...")

	k8sCfg := client.NewConfig(k8sFlags)
	k9sCfg := config.NewConfig(k8sCfg)
	var errs error

	conn, err := client.InitConnection(k8sCfg, slog.Default())
	if err != nil {
		errs = errors.Join(errs, err)
	}
	k9sCfg.SetConnection(conn)

	if err := k9sCfg.Load(config.AppConfigFile, false); err != nil {
		errs = errors.Join(errs, err)
	}
	k9sCfg.K9s.Override(k9sFlags)
	if err := k9sCfg.Refine(k8sFlags, k9sFlags, k8sCfg); err != nil {
		slog.Error("Fail to refine k9s config", slogs.Error, err)
		errs = errors.Join(errs, err)
	}

	// Try to access server version if that fail. Connectivity issue?
	if !conn.CheckConnectivity() {
		errs = errors.Join(errs, fmt.Errorf("cannot connect to context: %s", k9sCfg.K9s.ActiveContextName()))
	}
	if !conn.ConnectionOK() {
		slog.Warn("💣 Kubernetes connectivity toast!")
		errs = errors.Join(errs, fmt.Errorf("k8s connection failed for context: %s", k9sCfg.K9s.ActiveContextName()))
	} else {
		slog.Info("✅ Kubernetes connectivity OK")
	}

	if err := k9sCfg.Save(false); err != nil {
		slog.Error("K9s config save failed", slogs.Error, err)
		errs = errors.Join(errs, err)
	}

	return k9sCfg, errs
}

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func initK9sFlags() {
	k9sFlags = config.NewFlags()
	rootCmd.Flags().Float32VarP(
		k9sFlags.RefreshRate,
		"refresh", "r",
		config.DefaultRefreshRate,
		"Specify the default refresh rate as a float (sec)",
	)
	rootCmd.Flags().StringVarP(
		k9sFlags.LogLevel,
		"logLevel", "l",
		config.DefaultLogLevel,
		"Specify a log level (error, warn, info, debug)",
	)
	rootCmd.Flags().StringVarP(
		k9sFlags.LogFile,
		"logFile", "",
		config.AppLogFile,
		"Specify the log file",
	)
	rootCmd.Flags().BoolVar(
		k9sFlags.Headless,
		"headless",
		false,
		"Turn K9s header off",
	)
	rootCmd.Flags().BoolVar(
		k9sFlags.Logoless,
		"logoless",
		false,
		"Turn K9s logo off",
	)
	rootCmd.Flags().BoolVar(
		k9sFlags.Crumbsless,
		"crumbsless",
		false,
		"Turn K9s crumbs off",
	)
	rootCmd.Flags().BoolVar(
		k9sFlags.Splashless,
		"splashless",
		false,
		"Turn K9s splash screen off",
	)
	rootCmd.Flags().BoolVar(
		k9sFlags.Invert,
		"invert",
		false,
		"Invert skin (dark to light, light to dark), preserving colors",
	)
	rootCmd.Flags().BoolVarP(
		k9sFlags.AllNamespaces,
		"all-namespaces", "A",
		false,
		"Launch K9s in all namespaces",
	)
	rootCmd.Flags().StringVarP(
		k9sFlags.Command,
		"command", "c",
		config.DefaultCommand,
		"Overrides the default resource to load when the application launches",
	)
	rootCmd.Flags().BoolVar(
		k9sFlags.ReadOnly,
		"readonly",
		false,
		"Sets readOnly mode by overriding readOnly configuration setting",
	)
	rootCmd.Flags().BoolVar(
		k9sFlags.Write,
		"write",
		false,
		"Sets write mode by overriding the readOnly configuration setting",
	)
	rootCmd.Flags().StringVar(
		k9sFlags.ScreenDumpDir,
		"screen-dump-dir",
		"",
		"Sets a path to a dir for a screen dumps",
	)
	rootCmd.Flags()
}

func initK8sFlags() {
	k8sFlags = genericclioptions.NewConfigFlags(client.UsePersistentConfig)

	rootCmd.Flags().StringVar(
		k8sFlags.KubeConfig,
		"kubeconfig",
		"",
		"Path to the kubeconfig file to use for CLI requests",
	)

	rootCmd.Flags().StringVar(
		k8sFlags.Timeout,
		"request-timeout",
		"",
		"The length of time to wait before giving up on a single server request",
	)

	rootCmd.Flags().StringVar(
		k8sFlags.Context,
		"context",
		"",
		"The name of the kubeconfig context to use",
	)

	rootCmd.Flags().StringVar(
		k8sFlags.ClusterName,
		"cluster",
		"",
		"The name of the kubeconfig cluster to use",
	)

	rootCmd.Flags().StringVar(
		k8sFlags.AuthInfoName,
		"user",
		"",
		"The name of the kubeconfig user to use",
	)

	rootCmd.Flags().StringVarP(
		k8sFlags.Namespace,
		"namespace",
		"n",
		"",
		"If present, the namespace scope for this CLI request",
	)

	initAsFlags()
	initCertFlags()
	initK8sFlagCompletion()
}

func initAsFlags() {
	rootCmd.Flags().StringVar(
		k8sFlags.Impersonate,
		"as",
		"",
		"Username to impersonate for the operation",
	)

	rootCmd.Flags().StringArrayVar(
		k8sFlags.ImpersonateGroup,
		"as-group",
		[]string{},
		"Group to impersonate for the operation",
	)
}

func initCertFlags() {
	rootCmd.Flags().BoolVar(
		k8sFlags.Insecure,
		"insecure-skip-tls-verify",
		false,
		"If true, the server's caCertFile will not be checked for validity",
	)

	rootCmd.Flags().StringVar(
		k8sFlags.CAFile,
		"certificate-authority",
		"",
		"Path to a cert file for the certificate authority",
	)

	rootCmd.Flags().StringVar(
		k8sFlags.KeyFile,
		"client-key",
		"",
		"Path to a client key file for TLS",
	)

	rootCmd.Flags().StringVar(
		k8sFlags.CertFile,
		"client-certificate",
		"",
		"Path to a client certificate file for TLS",
	)

	rootCmd.Flags().StringVar(
		k8sFlags.BearerToken,
		"token",
		"",
		"Bearer token for authentication to the API server",
	)
}

type (
	k8sPickerFn[T any] func(cfg *api.Config) map[string]T
	completeFn         func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective)
)

func initK8sFlagCompletion() {
	_ = rootCmd.RegisterFlagCompletionFunc("context", k8sFlagCompletion(func(cfg *api.Config) map[string]*api.Context {
		return cfg.Contexts
	}))

	_ = rootCmd.RegisterFlagCompletionFunc("cluster", k8sFlagCompletion(func(cfg *api.Config) map[string]*api.Cluster {
		return cfg.Clusters
	}))

	_ = rootCmd.RegisterFlagCompletionFunc("user", k8sFlagCompletion(func(cfg *api.Config) map[string]*api.AuthInfo {
		return cfg.AuthInfos
	}))

	_ = rootCmd.RegisterFlagCompletionFunc("namespace", func(_ *cobra.Command, _ []string, s string) ([]string, cobra.ShellCompDirective) {
		conn := client.NewConfig(k8sFlags)
		if c, err := client.InitConnection(conn, slog.Default()); err == nil {
			if nss, err := c.ValidNamespaceNames(); err == nil {
				return filterFlagCompletions(nss, s)
			}
		}

		return nil, cobra.ShellCompDirectiveError
	})
}

func k8sFlagCompletion[T any](picker k8sPickerFn[T]) completeFn {
	return func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		conn := client.NewConfig(k8sFlags)
		cfg, err := conn.RawConfig()
		if err != nil {
			slog.Error("K8s raw config getter failed", slogs.Error, err)
		}

		return filterFlagCompletions(picker(&cfg), toComplete)
	}
}

func filterFlagCompletions[T any](m map[string]T, s string) ([]string, cobra.ShellCompDirective) {
	cc := make([]string, 0, len(m))
	for name := range m {
		if strings.HasPrefix(name, s) {
			cc = append(cc, name)
		}
	}

	return cc, cobra.ShellCompDirectiveNoFileComp
}
