package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/plab/plab-app/internal/tui"
	"github.com/spf13/cobra"
)

var updateTemplateCmd = &cobra.Command{
	Use:   "update-template",
	Short: "프로젝트의 공통 설정을 최신으로 업데이트해요",
	Long: `프로젝트의 공통 설정 파일을 최신 plab-app 템플릿 기준으로 업데이트합니다.

업데이트 항목:
  - ESLint 설정
  - Prettier 설정
  - PostCSS 설정
  - TypeScript 설정 (tsconfig.json)
  - Vercel 설정 (vercel.json)
  - devDependencies 버전 업데이트

업데이트하지 않는 항목 (사용자 코드):
  - app/ 디렉토리 (페이지, 컴포넌트)
  - lib/ 디렉토리 (유틸리티)
  - public/ 디렉토리 (정적 파일)
  - .env.local (환경 변수)

예시:
  cd plab-landing && plab-app update-template
  plab-app update-template --force
  plab-app update-template --json`,
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
				Title("공통 설정을 최신으로 업데이트할까요?").
				Description("ESLint, Prettier, tsconfig 등 설정 파일이 업데이트돼요. 앱 코드는 변경되지 않아요.").
				Affirmative("네, 업데이트해 주세요").
				Negative("취소").
				Value(&confirm).
				Run()

			if err != nil {
				if tui.IsUserAborted(err) {
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

		type updateStep struct {
			label string
			fn    func() error
		}

		steps := []updateStep{
			{
				label: "Prettier 설정 업데이트",
				fn: func() error {
					return writeConfig(cwd, ".prettierrc", `{
  "semi": true,
  "singleQuote": false,
  "tabWidth": 2,
  "trailingComma": "all"
}
`)
				},
			},
			{
				label: "PostCSS 설정 업데이트",
				fn: func() error {
					return writeConfig(cwd, "postcss.config.mjs", `/** @type {import('postcss-load-config').Config} */
const config = {
  plugins: {
    "@tailwindcss/postcss": {},
  },
};

export default config;
`)
				},
			},
			{
				label: "Vercel 설정 업데이트",
				fn: func() error {
					return writeConfig(cwd, "vercel.json", `{
  "$schema": "https://openapi.vercel.sh/vercel.json",
  "framework": "nextjs"
}
`)
				},
			},
			{
				label: "devDependencies 업데이트",
				fn: func() error {
					return updateDevDeps(cwd)
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
			fmt.Println(dimStyle.Render("  설정을 업데이트하고 있어요..."))
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
			fmt.Println(okStyle.Render("  업데이트 완료!"))
			fmt.Println(dimStyle.Render("  빌드도 정상적으로 통과했어요."))
		} else {
			fmt.Println(tui.ErrorStyle.Render("  일부 단계가 실패했어요. 위의 에러를 확인해 주세요."))
		}
		fmt.Println()
		return nil
	},
}

func writeConfig(dir, filename, content string) error {
	return os.WriteFile(filepath.Join(dir, filename), []byte(content), 0644)
}

func updateDevDeps(dir string) error {
	pkgPath := filepath.Join(dir, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return err
	}

	var pkg map[string]interface{}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return err
	}

	latestDevDeps := map[string]string{
		"@tailwindcss/postcss": "^4.0.0",
		"@types/node":         "^22.0.0",
		"@types/react":        "^19.0.0",
		"@types/react-dom":    "^19.0.0",
		"eslint":              "^9.0.0",
		"eslint-config-next":  "^15.0.0",
		"postcss":             "^8.0.0",
		"prettier":            "^3.0.0",
		"tailwindcss":         "^4.0.0",
		"typescript":          "^5.0.0",
	}

	devDeps, ok := pkg["devDependencies"].(map[string]interface{})
	if !ok {
		devDeps = make(map[string]interface{})
	}
	for k, v := range latestDevDeps {
		devDeps[k] = v
	}
	pkg["devDependencies"] = devDeps

	updated, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(pkgPath, append(updated, '\n'), 0644)
}

func init() {
	updateTemplateCmd.Flags().BoolVar(&flagForce, "force", false, "확인 없이 바로 실행")
	rootCmd.AddCommand(updateTemplateCmd)
}
