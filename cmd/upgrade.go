package cmd

import (
	"github.com/plab/plab-app/internal/updater"
	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "plab-app을 최신 버전으로 업데이트해요",
	Long: `plab-app 바이너리를 최신 버전으로 업데이트합니다.

GitHub Releases에서 최신 버전을 확인하고,
현재 OS에 맞는 바이너리를 다운로드하여 교체합니다.

예시:
  plab-app upgrade`,
	Run: func(cmd *cobra.Command, args []string) {
		updater.CheckAndUpdate(appVersion)
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
