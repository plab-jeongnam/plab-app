package main

import (
	"os"

	"github.com/plab/plab-app/cmd"
	"github.com/plab/plab-app/internal/tracking"
)

var version = "dev"

func main() {
	cmd.SetVersion(version)
	tracking.SetVersion(version)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
