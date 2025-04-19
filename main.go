package main

import (
	"github.com/lovelaze/nebula-sync/cmd"
	"os"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
