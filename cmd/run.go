package cmd

import (
	"github.com/lovelaze/nebula-sync/internal/config"
	"github.com/lovelaze/nebula-sync/internal/service"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var envFile string

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run sync",
	Run: func(cmd *cobra.Command, args []string) {
		readEnvFile()

		conf := config.Config{}
		conf.Load()

		service := service.NewService(conf)
		service.Run()
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
		log.Fatal().Err(err).Msg("error loading env file")
	}
}
