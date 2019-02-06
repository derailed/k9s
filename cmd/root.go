package cmd

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/views"
	"github.com/gdamore/tcell"
	"github.com/k8sland/tview"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	defaultRefreshRate = 2 // secs
	defaultLogLevel    = "info"
	defaultNamespace   = ""
)

var (
	version     = "dev"
	commit      = "dev"
	date        = "n/a"
	refreshRate int
	logLevel    string
	namespace   string
	kubeconfig  string

	rootCmd = &cobra.Command{
		Use:   "k9s",
		Short: "A graphical CLI for your Kubernetes cluster management.",
		Long:  `K9s is a Kubernetes CLI to view and manage your Kubernetes clusters.`,
		Run:   run,
	}
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print k9s version info",
		Long:  "Prints k9s version info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version:%s GitCommit:%s On %s\n", version, commit, date)
		},
	}
	infoCmd = &cobra.Command{
		Use:   "info",
		Short: "Print k9s configuration information",
		Long:  "Print k9s configuration information",
		Run: func(cmd *cobra.Command, args []string) {
			const (
				cyan    = "\033[1;36m%s\033[0m"
				green   = "\033[1;32m%s\033[0m"
				magenta = "\033[1;35m%s\033[0m"
			)
			fmt.Printf(cyan+"\n", strings.Repeat("-", 80))
			fmt.Printf(green+"\n", "🐶 K9s Information")
			fmt.Printf(magenta, fmt.Sprintf("%-10s", "LogFile:"))
			fmt.Printf("%s\n", views.K9sLogs)
			fmt.Printf(magenta, fmt.Sprintf("%-10s", "Config:"))
			fmt.Printf("%s\n", views.K9sConfig)
			fmt.Printf(cyan+"\n", strings.Repeat("-", 80))
		},
	}
)

func init() {
	rootCmd.AddCommand(versionCmd, infoCmd)

	rootCmd.Flags().IntVarP(
		&refreshRate,
		"refresh", "r",
		defaultRefreshRate,
		"Specifies the default refresh rate as an integer (sec)",
	)

	rootCmd.Flags().StringVarP(
		&namespace,
		"namespace", "n",
		defaultNamespace,
		"Uses a given namespace versus all-namespaces",
	)

	rootCmd.Flags().StringVarP(
		&logLevel,
		"logLevel", "l",
		defaultLogLevel,
		"Specify a log level (info, warn, debug, error, fatal, panic, trace)",
	)
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

	initStyles()
	initKeys()

	app := views.NewApp(version, refreshRate, namespace)
	{
		app.Init()
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
}

func initStyles() {
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorBlack
	tview.Styles.ContrastBackgroundColor = tcell.ColorBlack
	tview.Styles.FocusColor = tcell.ColorLightSkyBlue
	tview.Styles.BorderColor = tcell.ColorDodgerBlue
}
