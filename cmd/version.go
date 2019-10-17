package cmd

import (
	"fmt"

	"github.com/derailed/k9s/internal/color"
	"github.com/spf13/cobra"
)

func versionCmd() *cobra.Command {
	var short bool

	command := cobra.Command{
		Use:   "version",
		Short: "Print version/build info",
		Long:  "Print version/build information",
		Run: func(cmd *cobra.Command, args []string) {
			printVersion(short)
		},
	}

	command.PersistentFlags().BoolVarP(&short, "short", "s", false, "Simplified print version for resumed printing")

	return &command
}

func printVersion(short bool) {
	const secFmt = "%-10s "
	var outputColor color.Paint

	if short {
		outputColor = -1
	} else {
		outputColor = color.Cyan
		printLogo(outputColor)
	}
	printTuple(secFmt, "Version", version, outputColor)
	printTuple(secFmt, "Commit", commit, outputColor)
	printTuple(secFmt, "Date", date, outputColor)
}

func printTuple(format, section, value string, outputColor color.Paint) {
	if outputColor != -1 {
		section = color.Colorize(fmt.Sprintf(section+":"), outputColor)
		value = color.Colorize(value, color.White)
	}
	fmt.Println(fmt.Sprintf(format, section), value)
}
