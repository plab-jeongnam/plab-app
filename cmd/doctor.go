package cmd

import (
	"github.com/plab/plab-app/internal/doctor"
	"github.com/plab/plab-app/internal/platform"
	"github.com/spf13/cobra"
)

var flagSimulateOS string

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "개발 환경이 준비되었는지 확인해요",
	Long: `개발 환경 점검 도구.

필수 항목 (없으면 create 차단):
  - Git, Node.js (v18+), npm, gh CLI, GitHub 인증

권장 항목 (없어도 create 가능):
  - Claude Code

예시:
  plab-app doctor                        # 현재 환경 점검
  plab-app doctor --json                 # JSON 형식으로 결과 반환
  plab-app doctor --simulate-os windows  # Windows 환경 시뮬레이션`,
	Run: func(cmd *cobra.Command, args []string) {
		var report doctor.Report
		if flagSimulateOS != "" {
			plat := platform.ForOS(flagSimulateOS)
			report = doctor.RunWithPlatform(plat)
		} else {
			report = doctor.Run()
		}

		if flagJSON {
			PrintJSON(report.ToJSON())
			return
		}

		report.Print()
	},
}

func init() {
	doctorCmd.Flags().StringVar(&flagSimulateOS, "simulate-os", "", "다른 OS 환경을 시뮬레이션 (darwin, windows)")
}
