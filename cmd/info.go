package cmd

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/config"
	"github.com/spf13/cobra"
)

func infoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Print configuration information",
		Long:  "Print configuration information",
		Run: func(cmd *cobra.Command, args []string) {
			const (
				cyan    = "\033[1;36m%s\033[0m"
				green   = "\033[1;32m%s\033[0m"
				magenta = "\033[1;35m%s\033[0m"
			)
			fmt.Printf(cyan+"\n", strings.Repeat("-", 80))
			fmt.Printf(green+"\n", "🐶 K9s Information")
			fmt.Printf(magenta, fmt.Sprintf("%-10s", "LogFile:"))
			fmt.Printf("%s\n", config.K9sLogs)
			fmt.Printf(magenta, fmt.Sprintf("%-10s", "Config:"))
			fmt.Printf("%s\n", config.K9sConfigFile)
			fmt.Printf(cyan+"\n", strings.Repeat("-", 80))
		},
	}
}
