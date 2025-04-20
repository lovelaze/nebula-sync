package cmd

import (
	"github.com/spf13/cobra"

	"github.com/lovelaze/nebula-sync/internal/log"
	"github.com/lovelaze/nebula-sync/version"
)

var rootCmd = &cobra.Command{
	Use:     "nebula-sync",
	Version: version.Version,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(log.Init)
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
}
