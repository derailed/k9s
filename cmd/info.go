package cmd

import (
	"fmt"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/printer"
	"github.com/derailed/k9s/internal/views"
	"github.com/spf13/cobra"
)

func infoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Print configuration info",
		Long:  "Print configuration information",
		Run: func(cmd *cobra.Command, args []string) {
			printInfo()
		},
	}
}

func printInfo() {
	const secFmt = "%-15s "

	printLogo()
	printTuple(secFmt, "Configuration", config.K9sConfigFile)
	printTuple(secFmt, "Logs", config.K9sLogs)
}

func printLogo() {
	for _, l := range views.LogoSmall {
		fmt.Println(printer.Colorize(l, printer.ColorCyan))
	}
	fmt.Println()
}
