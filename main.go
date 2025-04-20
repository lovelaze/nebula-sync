package main

import (
	"os"

	"github.com/lovelaze/nebula-sync/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
