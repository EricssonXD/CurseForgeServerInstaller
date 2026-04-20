package main

import (
	"os"

	"github.com/ericsson/curseforge-server-installer/cmd/cfs/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
