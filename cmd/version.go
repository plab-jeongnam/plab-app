package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "plab-app 버전을 확인해요",
	Run: func(cmd *cobra.Command, args []string) {
		if flagJSON {
			PrintJSON(map[string]string{"version": appVersion})
			return
		}
		fmt.Printf("plab-app %s\n", appVersion)
	},
}
