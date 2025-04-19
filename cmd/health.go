package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/lovelaze/nebula-sync/internal/health"
)

const healthURL = "http://127.0.0.1:8080/health"

var healthCmd = &cobra.Command{
	Use: "healthcheck",
	Run: func(cmd *cobra.Command, args []string) {
		if err := health.Check(healthURL); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	healthCmd.Hidden = true
	rootCmd.AddCommand(healthCmd)
}
