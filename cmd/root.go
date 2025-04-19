package cmd

import (
	"github.com/lovelaze/nebula-sync/internal/log"
	"github.com/lovelaze/nebula-sync/version"
	"github.com/spf13/cobra"
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
