package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/plab/plab-app/internal/tui"
	"github.com/spf13/cobra"
)

var flagForce bool

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "프로젝트를 깨끗한 상태로 복구해요",
	Long: `프로젝트를 깨끗한 상태로 복구합니다.

수행 작업:
  1. node_modules 삭제
  2. .next 빌드 캐시 삭제
  3. npm install (패키지 재설치)
  4. npm run build (빌드 검증)

예시:
  cd plab-landing && plab-app reset
  plab-app reset --force    # 확인 없이 바로 실행
  plab-app reset --json     # JSON 형식 결과 반환`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, _ := os.Getwd()

		if _, err := os.Stat(filepath.Join(cwd, "package.json")); err != nil {
			if flagJSON {
				PrintCLIError("not_project", "plab 프로젝트 디렉토리가 아니에요.", "package.json이 있는 프로젝트 폴더에서 실행해 주세요.", "")
				os.Exit(1)
			}
			return fmt.Errorf("package.json이 없어요. plab 프로젝트 디렉토리에서 실행해 주세요")
		}

		if !flagForce && !flagJSON {
			var confirm bool
			err := huh.NewConfirm().
				Title("프로젝트를 초기화할까요?").
				Description("node_modules와 빌드 캐시를 삭제하고 다시 설치해요.").
				Affirmative("네, 초기화해 주세요").
				Negative("취소").
				Value(&confirm).
				Run()

			if err != nil {
				if tui.IsUserAborted(err) {
					fmt.Println()
					fmt.Println(tui.DimStyle.Render("  취소했어요."))
					fmt.Println()
					return nil
				}
				return err
			}
			if !confirm {
				fmt.Println()
				fmt.Println(tui.DimStyle.Render("  취소했어요."))
				fmt.Println()
				return nil
			}
		}

		spinnerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
		okMark := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("✓")
		failMark := lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("✗")
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

		type resetStep struct {
			label string
			fn    func() error
		}

		steps := []resetStep{
			{
				label: "node_modules 삭제",
				fn: func() error {
					return os.RemoveAll(filepath.Join(cwd, "node_modules"))
				},
			},
			{
				label: ".next 캐시 삭제",
				fn: func() error {
					return os.RemoveAll(filepath.Join(cwd, ".next"))
				},
			},
			{
				label: "패키지 재설치",
				fn: func() error {
					cmd := exec.Command("npm", "install")
					cmd.Dir = cwd
					cmd.Stdout = nil
					cmd.Stderr = nil
					return cmd.Run()
				},
			},
			{
				label: "빌드 검증",
				fn: func() error {
					cmd := exec.Command("npm", "run", "build")
					cmd.Dir = cwd
					cmd.Stdout = nil
					cmd.Stderr = nil
					return cmd.Run()
				},
			},
		}

		if !flagJSON {
			fmt.Println()
			fmt.Println(dimStyle.Render("  프로젝트를 복구하고 있어요..."))
			fmt.Println()
		}

		results := make([]map[string]interface{}, len(steps))
		allOK := true

		for i, s := range steps {
			if !flagJSON {
				fmt.Printf("  %s %s\n", spinnerStyle.Render("⏳"), s.label)
			}
			err := s.fn()
			if !flagJSON {
				fmt.Print("\033[1A\033[2K")
			}

			results[i] = map[string]interface{}{
				"step": s.label,
				"ok":   err == nil,
			}

			if err != nil {
				allOK = false
				results[i]["error"] = err.Error()
				if !flagJSON {
					fmt.Printf("  %s %s\n", failMark, s.label)
				}
			} else {
				if !flagJSON {
					fmt.Printf("  %s %s\n", okMark, s.label)
				}
			}
		}

		if flagJSON {
			PrintJSON(map[string]interface{}{
				"success": allOK,
				"steps":   results,
			})
			return nil
		}

		fmt.Println()
		if allOK {
			okStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
			fmt.Println(okStyle.Render("  복구 완료! 프로젝트가 정상이에요."))
		} else {
			fmt.Println(tui.ErrorStyle.Render("  일부 단계가 실패했어요. 위의 에러를 확인해 주세요."))
		}
		fmt.Println()
		return nil
	},
}

func init() {
	resetCmd.Flags().BoolVar(&flagForce, "force", false, "확인 없이 바로 실행")
	rootCmd.AddCommand(resetCmd)
}
