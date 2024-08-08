package cmd

import (
	"github.com/lovelaze/nebula-sync/internal/config"
	"github.com/lovelaze/nebula-sync/internal/service"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run sync",
	Run: func(cmd *cobra.Command, args []string) {
		conf := config.Config{}
		conf.Load()

		service := service.NewService(conf)
		service.Run()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
