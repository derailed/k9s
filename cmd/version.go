package cmd

import (
	"fmt"

	"github.com/derailed/k9s/internal/printer"
	"github.com/spf13/cobra"
)

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version info",
		Long:  "Prints version info",
		Run: func(cmd *cobra.Command, args []string) {
			const secFmt = "%-10s"
			fmt.Printf(printer.Colorize(fmt.Sprintf(secFmt, "Version:"), printer.ColorMagenta))
			fmt.Println(printer.Colorize(version, printer.ColorDarkGray))
			fmt.Printf(printer.Colorize(fmt.Sprintf(secFmt, "Commit:"), printer.ColorMagenta))
			fmt.Println(printer.Colorize(commit, printer.ColorDarkGray))
			fmt.Printf(printer.Colorize(fmt.Sprintf(secFmt, "Date:"), printer.ColorMagenta))
			fmt.Println(printer.Colorize(date, printer.ColorDarkGray))
		},
	}
}
