package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/charmbracelet/lipgloss"
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

		if !flagFix || report.AllPassed() {
			return
		}

		plat := platform.Detect()
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11"))

		// 설치 가능한 항목 수집
		type fixItem struct {
			name       string
			key        string
			installCmd string
			required   bool
		}

		var fixable []fixItem
		for _, r := range report.Results {
			if r.OK {
				continue
			}
			key := doctor.ToolKey(r.Name)
			installCmd := plat.InstallCommand(key)
			if installCmd != "" {
				fixable = append(fixable, fixItem{
					name:       r.Name,
					key:        key,
					installCmd: installCmd,
					required:   r.Required,
				})
			}
		}

		if len(fixable) == 0 {
			fmt.Println(dimStyle.Render("  자동으로 설치할 수 있는 항목이 없어요."))
			fmt.Println()
			return
		}

		// 설치할 항목 안내
		fmt.Println(tui.TitleStyle.Render("  누락된 도구를 설치할게요"))
		fmt.Println()
		for _, f := range fixable {
			tag := ""
			if !f.required {
				tag = dimStyle.Render(" (권장)")
			}
			fmt.Printf("  %s %s%s\n", dimStyle.Render("→"), f.name, tag)
			fmt.Printf("    %s\n", dimStyle.Render(f.installCmd))
		}
		fmt.Println()

		// 설치 실행
		installed := 0
		failed := 0

		for _, f := range fixable {
			sp := tui.NewSpinner(f.name + " 설치 중")
			sp.Start()

			var c *exec.Cmd
			if runtime.GOOS == "windows" {
				c = exec.Command("cmd", "/c", f.installCmd)
			} else {
				c = exec.Command("sh", "-c", f.installCmd)
			}
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			err := c.Run()

			if err != nil {
				sp.Stop(false)
				fmt.Printf("    %s %s\n", warnStyle.Render("→"), dimStyle.Render("직접 설치해 주세요: "+f.installCmd))
				failed++
			} else {
				sp.Stop(true)
				installed++
			}
		}

		// 결과 요약
		fmt.Println()
		if installed > 0 {
			fmt.Printf("  %s %d개 도구 설치 완료\n", tui.SuccessStyle.Render("✓"), installed)
		}
		if failed > 0 {
			fmt.Printf("  %s %d개 도구 설치 실패\n", tui.ErrorStyle.Render("✗"), failed)
		}
		fmt.Println()

		// 재점검
		fmt.Println(tui.TitleStyle.Render("  다시 점검할게요..."))
		fmt.Println()

		recheck := doctor.Run()
		recheck.Print()

		if recheck.RequiredPassed() {
			okStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
			fmt.Println(okStyle.Render("  이제 프로젝트를 만들 수 있어요!"))
			fmt.Println(dimStyle.Render("  plab-app create"))
			fmt.Println()
		}
	},
}

func init() {
	doctorCmd.Flags().StringVar(&flagSimulateOS, "simulate-os", "", "다른 OS 환경을 시뮬레이션 (darwin, windows)")
	doctorCmd.Flags().BoolVar(&flagFix, "fix", false, "누락된 도구를 자동으로 설치")
}
