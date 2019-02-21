package cmd

import (
	"fmt"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/views"
	"github.com/gdamore/tcell"
	"github.com/k8sland/tview"
	log "github.com/sirupsen/logrus"
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
)

var _ config.KubeSettings = &k8s.Config{}

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

func initK9s() {
	log.Info("🐶 K9s starting up...")

	// Load K9s config file...
	cfg := k8s.NewConfig(k8sFlags)
	config.Root = config.NewConfig(cfg)
	initK9sConfig()

	// Init K8s connection...
	k8s.InitConnectionOrDie(cfg)
	log.Info("✅ Kubernetes connectivity")

	config.Root.Save()
}

func initK9sConfig() {
	if err := config.Root.Load(config.K9sConfigFile); err != nil {
		log.Warnf("Unable to locate K9s config. Generating new configuration...")
	}
	config.Root.K9s.RefreshRate = refreshRate

	cfg, err := k8sFlags.ToRawKubeConfigLoader().RawConfig()
	if err != nil {
		panic("Invalid configuration. Unable to connect to api")
	}
	ctx := cfg.CurrentContext
	if isSet(k8sFlags.Context) {
		ctx = *k8sFlags.Context
	}
	config.Root.K9s.CurrentContext = ctx

	log.Debugf("Active Context `%v`", ctx)

	if isSet(k8sFlags.Namespace) {
		config.Root.SetActiveNamespace(*k8sFlags.Namespace)
	}

	if c, ok := cfg.Contexts[ctx]; ok {
		config.Root.K9s.CurrentCluster = c.Cluster
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
		log.Panic(err)
	}
}

func run(cmd *cobra.Command, args []string) {
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		level = log.DebugLevel
	}
	log.SetLevel(level)
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true, ForceColors: true})

	initK9s()
	initStyles()
	initKeys()

	app := views.NewApp()
	{
		app.Init(version, refreshRate, k8sFlags)
		app.Run()
	}
}

func initKeys() {
	tcell.KeyNames[tcell.Key(views.Key0)] = "0"
	tcell.KeyNames[tcell.Key(views.Key1)] = "1"
	tcell.KeyNames[tcell.Key(views.Key2)] = "2"
	tcell.KeyNames[tcell.Key(views.Key3)] = "3"
	tcell.KeyNames[tcell.Key(views.Key4)] = "4"
	tcell.KeyNames[tcell.Key(views.Key5)] = "5"
	tcell.KeyNames[tcell.Key(views.Key6)] = "6"
	tcell.KeyNames[tcell.Key(views.Key7)] = "7"
	tcell.KeyNames[tcell.Key(views.Key8)] = "8"
	tcell.KeyNames[tcell.Key(views.Key9)] = "9"
	tcell.KeyNames[tcell.Key(views.KeyA)] = "a"
	tcell.KeyNames[tcell.Key(views.KeyB)] = "b"
	tcell.KeyNames[tcell.Key(views.KeyC)] = "c"
	tcell.KeyNames[tcell.Key(views.KeyD)] = "d"
	tcell.KeyNames[tcell.Key(views.KeyE)] = "e"
	tcell.KeyNames[tcell.Key(views.KeyF)] = "f"
	tcell.KeyNames[tcell.Key(views.KeyG)] = "g"
	tcell.KeyNames[tcell.Key(views.KeyH)] = "h"
	tcell.KeyNames[tcell.Key(views.KeyI)] = "i"
	tcell.KeyNames[tcell.Key(views.KeyJ)] = "j"
	tcell.KeyNames[tcell.Key(views.KeyK)] = "k"
	tcell.KeyNames[tcell.Key(views.KeyL)] = "l"
	tcell.KeyNames[tcell.Key(views.KeyM)] = "m"
	tcell.KeyNames[tcell.Key(views.KeyN)] = "n"
	tcell.KeyNames[tcell.Key(views.KeyO)] = "o"
	tcell.KeyNames[tcell.Key(views.KeyP)] = "p"
	tcell.KeyNames[tcell.Key(views.KeyQ)] = "q"
	tcell.KeyNames[tcell.Key(views.KeyR)] = "r"
	tcell.KeyNames[tcell.Key(views.KeyS)] = "s"
	tcell.KeyNames[tcell.Key(views.KeyT)] = "t"
	tcell.KeyNames[tcell.Key(views.KeyU)] = "u"
	tcell.KeyNames[tcell.Key(views.KeyV)] = "v"
	tcell.KeyNames[tcell.Key(views.KeyX)] = "x"
	tcell.KeyNames[tcell.Key(views.KeyY)] = "y"
	tcell.KeyNames[tcell.Key(views.KeyZ)] = "z"
	tcell.KeyNames[tcell.Key(views.KeyHelp)] = "?"
}

func initStyles() {
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorBlack
	tview.Styles.ContrastBackgroundColor = tcell.ColorBlack
	tview.Styles.FocusColor = tcell.ColorLightSkyBlue
	tview.Styles.BorderColor = tcell.ColorDodgerBlue
}
