package cmd

import (
	"flag"
	"fmt"
	"runtime/debug"

	"github.com/derailed/k9s/internal/color"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/views"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog"
)

const (
	appName      = "k9s"
	shortAppDesc = "A graphical CLI for your Kubernetes cluster management."
	longAppDesc  = "K9s is a CLI to view and manage your Kubernetes clusters."
)

var (
	version, commit, date = "dev", "dev", "n/a"
	k9sFlags              *config.Flags
	k8sFlags              *genericclioptions.ConfigFlags

	rootCmd = &cobra.Command{
		Use:   appName,
		Short: shortAppDesc,
		Long:  longAppDesc,
		Run:   run,
	}
	_ config.KubeSettings = &k8s.Config{}
)

func init() {
	rootCmd.AddCommand(versionCmd(), infoCmd())
	initK9sFlags()
	initK8sFlags()

	// Klogs (of course) want to print stuff to the screen ;(
	klog.InitFlags(nil)
	flag.Set("log_file", config.K9sLogs)
	flag.Set("stderrthreshold", "fatal")
	flag.Set("alsologtostderr", "false")
	flag.Set("logtostderr", "false")
}

// Execute root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Panic().Err(err)
	}
}

func run(cmd *cobra.Command, args []string) {
	defer func() {
		// views.ClearScreen()
		if err := recover(); err != nil {
			log.Error().Msgf("Boom! %v", err)
			log.Error().Msg(string(debug.Stack()))
			printLogo(color.Red)
			fmt.Printf(color.Colorize("Boom!! ", color.Red))
			fmt.Println(color.Colorize(fmt.Sprintf("%v.", err), color.White))
		}
	}()

	zerolog.SetGlobalLevel(parseLevel(*k9sFlags.LogLevel))
	cfg := loadConfiguration()
	app := views.NewApp(cfg)
	{
		defer app.BailOut()
		app.Init(version, *k9sFlags.RefreshRate)
		app.Run()
	}
}

func loadConfiguration() *config.Config {
	log.Info().Msg("🐶 K9s starting up...")

	// Load K9s config file...
	k8sCfg := k8s.NewConfig(k8sFlags)
	k9sCfg := config.NewConfig(k8sCfg)
	if err := k9sCfg.Load(config.K9sConfigFile); err != nil {
		log.Warn().Msg("Unable to locate K9s config. Generating new configuration...")
	}

	if *k9sFlags.RefreshRate != config.DefaultRefreshRate {
		k9sCfg.K9s.OverrideRefreshRate(*k9sFlags.RefreshRate)
	}

	if k9sFlags.Headless != nil {
		k9sCfg.K9s.OverrideHeadless(*k9sFlags.Headless)
	}

	if k9sFlags.Command != nil {
		k9sCfg.K9s.OverrideCommand(*k9sFlags.Command)
	}

	if k9sFlags.AllNamespaces != nil && *k9sFlags.AllNamespaces {
		k9sCfg.SetActiveNamespace(resource.AllNamespaces)
	}

	if err := k9sCfg.Refine(k8sFlags); err != nil {
		log.Panic().Err(err).Msg("Unable to locate kubeconfig file")
	}
	k9sCfg.SetConnection(k8s.InitConnectionOrDie(k8sCfg, log.Logger))

	// Try to access server version if that fail. Connectivity issue?
	if _, err := k9sCfg.GetConnection().ServerVersion(); err != nil {
		log.Panic().Err(err).Msg("K9s can't connect to cluster")
	}
	log.Info().Msg("✅ Kubernetes connectivity")
	k9sCfg.Save()

	return k9sCfg
}

func parseLevel(level string) zerolog.Level {
	switch level {
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
		"Specify a log level (info, warn, debug, error, fatal, panic, trace)",
	)
	rootCmd.Flags().BoolVar(
		k9sFlags.Headless,
		"headless",
		false,
		"Turn K9s header off",
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
		"Specify the default command to view when the application launches",
	)
}

func initK8sFlags() {
	k8sFlags = genericclioptions.NewConfigFlags(false)

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
