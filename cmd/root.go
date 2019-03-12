package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/views"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const (
	defaultRefreshRate = 2 // secs
	defaultLogLevel    = "info"
)

var (
	version     = "dev"
	commit      = "dev"
	date        = "n/a"
	refreshRate int
	logLevel    string
	k8sFlags    *genericclioptions.ConfigFlags

	rootCmd = &cobra.Command{
		Use:   "k9s",
		Short: "A graphical CLI for your Kubernetes cluster management.",
		Long:  `K9s is a CLI to view and manage your Kubernetes clusters.`,
		Run:   run,
	}
	_ config.KubeSettings = &k8s.Config{}
)

func init() {
	rootCmd.AddCommand(versionCmd(), infoCmd())

	rootCmd.Flags().IntVarP(
		&refreshRate,
		"refresh", "r",
		defaultRefreshRate,
		"Specifies the default refresh rate as an integer (sec)",
	)
	rootCmd.Flags().StringVarP(
		&logLevel,
		"logLevel", "l",
		defaultLogLevel,
		"Specify a log level (info, warn, debug, error, fatal, panic, trace)",
	)

	initK8sFlags()
}

func initK9s() {
	log.Info().Msg("🐶 K9s starting up...")

	// Load K9s config file...
	cfg := k8s.NewConfig(k8sFlags)
	config.Root = config.NewConfig(cfg)
	initK9sConfig()

	// Init K8s connection...
	k8s.InitConnectionOrDie(cfg)
	log.Info().Msg("✅ Kubernetes connectivity")

	config.Root.Save()
}

func initK9sConfig() {
	if err := config.Root.Load(config.K9sConfigFile); err != nil {
		log.Warn().Msg("Unable to locate K9s config. Generating new configuration...")
	}
	config.Root.K9s.RefreshRate = refreshRate

	cfg, err := k8sFlags.ToRawKubeConfigLoader().RawConfig()
	if err != nil {
		panic("Invalid configuration. Unable to connect to api")
	}

	ctx := cfg.CurrentContext
	switch {
	case isSet(k8sFlags.Context):
		ctx = *k8sFlags.Context
		config.Root.K9s.CurrentContext = ctx
	case isSet(&config.Root.K9s.CurrentContext):
		k8sFlags.Context = &config.Root.K9s.CurrentContext
	default:
		config.Root.K9s.CurrentContext = ctx
		if isSet(&cfg.Contexts[ctx].Namespace) {
			config.Root.SetActiveNamespace(cfg.Contexts[ctx].Namespace)
		}
	}
	log.Debug().Msgf("Active Context `%v`", ctx)

	if isSet(k8sFlags.Namespace) {
		config.Root.SetActiveNamespace(*k8sFlags.Namespace)
	}

	if isSet(k8sFlags.ClusterName) {
		config.Root.K9s.CurrentCluster = *k8sFlags.ClusterName
	}

	if c, ok := cfg.Contexts[ctx]; ok {
		config.Root.K9s.CurrentCluster = c.Cluster
		if len(c.Namespace) != 0 {
			config.Root.SetActiveNamespace(c.Namespace)
		}
	} else {
		panic(fmt.Sprintf("The specified context `%s does not exists in kubeconfig", cfg.CurrentContext))
	}
}

func isSet(s *string) bool {
	return s != nil && len(*s) > 0
}

// Execute root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Panic().Err(err)
	}
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

func run(cmd *cobra.Command, args []string) {
	zerolog.SetGlobalLevel(parseLevel(logLevel))

	initK9s()

	app := views.NewApp()
	{
		app.Init(version, refreshRate, k8sFlags)
		defer func() {
			clearScreen()
			if err := recover(); err != nil {
				app.Stop()
				log.Error().Msgf("Boom! %#v", err)
				debug.PrintStack()
			}
		}()
		app.Run()
	}
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
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

	rootCmd.Flags().StringVarP(
		k8sFlags.Namespace,
		"namespace",
		"n",
		"",
		"If present, the namespace scope for this CLI request",
	)
}
