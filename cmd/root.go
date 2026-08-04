package cmd

import (
	"os"

	"github.com/bestruirui/octopus/internal/conf"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   conf.APP_NAME,
	Short: conf.APP_DESC,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
