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
		Short: "Print configuration information",
		Long:  "Print configuration information",
		Run: func(cmd *cobra.Command, args []string) {
			printInfo()
		},
	}
}

func printInfo() {
	for _, l := range views.LogoSmall {
		fmt.Println(printer.Colorize(l, printer.ColorCyan))
	}
	fmt.Println()
	fmt.Printf(printer.Colorize(fmt.Sprintf("%-15s", "Configuration:"), printer.ColorCyan))
	fmt.Println(printer.Colorize(config.K9sConfigFile, printer.ColorWhite))

	fmt.Printf(printer.Colorize(fmt.Sprintf("%-15s", "Logs:"), printer.ColorCyan))
	fmt.Println(printer.Colorize(config.K9sLogs, printer.ColorWhite))
}
