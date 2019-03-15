package cmd

import (
	"fmt"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/printer"
	"github.com/spf13/cobra"
)

func infoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Print configuration information",
		Long:  "Print configuration information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf(printer.Colorize(fmt.Sprintf("%-15s", "Configuration:"), printer.ColorMagenta))
			fmt.Println(printer.Colorize(config.K9sConfigFile, printer.ColorDarkGray))

			fmt.Printf(printer.Colorize(fmt.Sprintf("%-15s", "Logs:"), printer.ColorMagenta))
			fmt.Println(printer.Colorize(config.K9sLogs, printer.ColorDarkGray))
		},
	}
}
