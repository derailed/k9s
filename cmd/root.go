// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package cmd

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/derailed/k9s/internal/config/data"
	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/color"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/view"
	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
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

	rootCmd.SetFlagErrorFunc(func(command *cobra.Command, err error) error {
		return flagError{err: err}
	})

	rootCmd.AddCommand(versionCmd(), infoCmd())
	initK9sFlags()
	initK8sFlags()
}

// Execute root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if !errors.As(err, &flagError{}) {
			panic(err)
		}
	}
}

func run(cmd *cobra.Command, args []string) error {
	if err := config.InitLocs(); err != nil {
		return err
	}
	file, err := os.OpenFile(
		*k9sFlags.LogFile,
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		data.DefaultFileMod,
	)
	if err != nil {
		return fmt.Errorf("Log file %q init failed: %w", *k9sFlags.LogFile, err)
	}
	defer func() {
		if file != nil {
			_ = file.Close()
		}
	}()
	defer func() {
		if err := recover(); err != nil {
			log.Error().Msgf("Boom! %v", err)
			log.Error().Msg(string(debug.Stack()))
			printLogo(color.Red)
			fmt.Printf("%s", color.Colorize("Boom!! ", color.Red))
			fmt.Printf("%v.\n", err)
		}
	}()

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: file})
	zerolog.SetGlobalLevel(parseLevel(*k9sFlags.LogLevel))

	cfg, err := loadConfiguration()
	if err != nil {
		log.Error().Err(err).Msgf("Fail to load global/context configuration")
	}
	app := view.NewApp(cfg)
	if err := app.Init(version, *k9sFlags.RefreshRate); err != nil {
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
	log.Info().Msg("🐶 K9s starting up...")

	k8sCfg := client.NewConfig(k8sFlags)
	k9sCfg := config.NewConfig(k8sCfg)
	var errs error

	if err := k9sCfg.Load(config.AppConfigFile, false); err != nil {
		errs = errors.Join(errs, err)
	}
	k9sCfg.K9s.Override(k9sFlags)
	if err := k9sCfg.Refine(k8sFlags, k9sFlags, k8sCfg); err != nil {
		log.Error().Err(err).Msgf("config refine failed")
		errs = errors.Join(errs, err)
	}

	conn, err := client.InitConnection(k8sCfg)

	if err != nil {
		errs = errors.Join(errs, err)
	}

	// Try to access server version if that fail. Connectivity issue?
	if !conn.CheckConnectivity() {
		errs = errors.Join(errs, fmt.Errorf("cannot connect to context: %s", k9sCfg.K9s.ActiveContextName()))
	}

	if !conn.ConnectionOK() {
		errs = errors.Join(errs, fmt.Errorf("k8s connection failed for context: %s", k9sCfg.K9s.ActiveContextName()))
	}

	k9sCfg.SetConnection(conn)

	log.Info().Msg("✅ Kubernetes connectivity")
	if err := k9sCfg.Save(false); err != nil {
		log.Error().Err(err).Msg("Config save")
		errs = errors.Join(errs, err)
	}

	return k9sCfg, errs
}

func parseLevel(level string) zerolog.Level {
	switch level {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}

func initK9sFlags() {
	k9sFlags = config.NewFlags()
	rootCmd.Flags().IntVarP(
		k9sFlags.RefreshRate,
		"refresh", "r",
		config.DefaultRefreshRate,
		"Specify the default refresh rate as an integer (sec)",
	)
	rootCmd.Flags().StringVarP(
		k9sFlags.LogLevel,
		"logLevel", "l",
		config.DefaultLogLevel,
		"Specify a log level (error, warn, info, debug, trace)",
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

	_ = rootCmd.RegisterFlagCompletionFunc("namespace", func(cmd *cobra.Command, args []string, s string) ([]string, cobra.ShellCompDirective) {
		conn := client.NewConfig(k8sFlags)
		if c, err := client.InitConnection(conn); err == nil {
			if nss, err := c.ValidNamespaceNames(); err == nil {
				return filterFlagCompletions(nss, s)
			}
		}

		return nil, cobra.ShellCompDirectiveError
	})
}

func k8sFlagCompletion[T any](picker k8sPickerFn[T]) completeFn {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		conn := client.NewConfig(k8sFlags)
		cfg, err := conn.RawConfig()
		if err != nil {
			log.Error().Err(err).Msgf("k8s config getter failed")
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
