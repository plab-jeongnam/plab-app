package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/plab/plab-app/internal/doctor"
	"github.com/plab/plab-app/internal/platform"
	"github.com/plab/plab-app/internal/tui"
	"github.com/spf13/cobra"
)

var (
	flagSimulateOS string
	flagFix        bool
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "개발 환경이 준비되었는지 확인해요",
	Long: `개발 환경 점검 도구.

필수 항목 (없으면 create 차단):
  - Git, Node.js (v18+), npm, gh CLI, GitHub 인증

권장 항목 (없어도 create 가능):
  - Vercel CLI, Claude Code

예시:
  plab-app doctor                        # 현재 환경 점검
  plab-app doctor --fix                  # 누락 도구 자동 설치
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

		if flagFix && !report.AllPassed() {
			plat := platform.Detect()
			fmt.Println(tui.TitleStyle.Render("  누락된 도구를 설치합니다..."))
			fmt.Println()

			for _, r := range report.Results {
				if r.OK {
					continue
				}
				installCmd := plat.InstallCommand(r.Name)
				if installCmd == "" {
					continue
				}

				sp := tui.NewSpinner(r.Name + " 설치")
				sp.Start()
				c := exec.Command("sh", "-c", installCmd)
				c.Stdout = os.Stdout
				c.Stderr = os.Stderr
				err := c.Run()
				sp.Stop(err == nil)
			}

			fmt.Println()
			fmt.Println(tui.DimStyle.Render("  설치 후 plab-app doctor 로 다시 확인해 주세요."))
			fmt.Println()
		}
	},
}

func init() {
	doctorCmd.Flags().StringVar(&flagSimulateOS, "simulate-os", "", "다른 OS 환경을 시뮬레이션 (darwin, windows)")
	doctorCmd.Flags().BoolVar(&flagFix, "fix", false, "누락된 도구를 자동으로 설치")
}
