package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/lovelaze/nebula-sync/internal/config"
	"github.com/lovelaze/nebula-sync/internal/service"
)

var envFile string

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run sync",
	Run: func(cmd *cobra.Command, args []string) {
		readEnvFile()

		service, err := service.Init()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize service")
		}

		if err = service.Run(); err != nil {
			log.Fatal().Err(err).Msg("Sync failed")
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVar(&envFile, "env-file", "", "Read env from `.env` file")
}

func readEnvFile() {
	if envFile == "" {
		return
	}

	if err := config.LoadEnvFile(envFile); err != nil {
		log.Fatal().Err(err).Msg("Failed to load env file")
	}
}
