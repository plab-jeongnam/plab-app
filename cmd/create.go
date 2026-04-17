package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/plab/plab-app/internal/config"
	"github.com/plab/plab-app/internal/doctor"
	"github.com/plab/plab-app/internal/generator"
	"github.com/plab/plab-app/internal/model"
	"github.com/plab/plab-app/internal/tracking"
	"github.com/plab/plab-app/internal/tui"
	"github.com/plab/plab-app/internal/updater"
	"github.com/spf13/cobra"
)

var (
	flagName            string
	flagPlabData        bool
	flagAPIKey          string
	flagResearchersOnly bool
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "새 프로젝트를 만들어요",
	Long: `새 프로젝트를 생성합니다.

플래그 없이 실행하면 대화형 TUI 모드로 시작합니다.
--name 플래그를 지정하면 CLI 모드로 즉시 생성합니다.

CLI 모드 (LLM/자동화):
  plab-app create --name landing                       # plab-landing 생성
  plab-app create --name admin --plab-data             # 플랩 데이터 연동 포함
  plab-app create --name api --plab-data --api-key KEY  # API 키도 함께 설정
  plab-app create --name dash --researchers-only       # 리서처 전용 (Google 로그인)
  plab-app create --name test --json                   # JSON 형식 결과 반환

대화형 TUI 모드:
  plab-app create                                      # 질문에 답하며 생성

동작:
  1. 환경 점검 (doctor 자동 실행)
  2. Next.js + TypeScript + Tailwind 프로젝트 생성
  3. npm install + npm run build 검증
  4. GitHub 저장소 자동 생성 + push

생성되는 프로젝트 구조:
  plab-{name}/
  ├── app/             # Next.js App Router 페이지
  ├── lib/             # 유틸리티 (플랩 API 클라이언트 등)
  ├── public/          # 정적 파일
  ├── package.json     # 의존성
  ├── tsconfig.json    # TypeScript 설정
  ├── vercel.json      # Vercel 배포 설정
  └── .env.local       # 환경 변수 (플랩 API 키)

Exit Codes:
  0  성공
  1  환경 미비 (plab-app doctor 실행 필요)
  1  디렉토리 이미 존재
  1  빌드 실패`,
	RunE: func(cmd *cobra.Command, args []string) error {
		startTime := time.Now()
		updater.CheckAndUpdate(appVersion)

		report := doctor.Run()
		if !report.RequiredPassed() {
			if flagJSON {
				PrintJSON(map[string]interface{}{
					"success": false,
					"error":   "required_tools_missing",
					"message": "필수 도구가 설치되어 있지 않습니다.",
					"fix":     "plab-app doctor --json 으로 누락 항목을 확인하세요.",
					"doctor":  report.ToJSON(),
				})
				os.Exit(1)
			}
			report.Print()
			fmt.Println(tui.ErrorStyle.Render("  먼저 위의 항목들을 설정해 주세요."))
			fmt.Println(tui.DimStyle.Render("  plab-app doctor 명령으로 다시 확인할 수 있어요."))
			fmt.Println()
			os.Exit(1)
			return nil
		}

		var project *model.Project

		if flagName != "" {
			name := strings.TrimSpace(strings.ToLower(flagName))
			name = strings.ReplaceAll(name, " ", "-")
			if !strings.HasPrefix(name, "plab-") {
				name = "plab-" + name
			}
			project = &model.Project{
				Name:            name,
				DisplayName:     flagName,
				UsePlabData:     flagPlabData,
				ResearchersOnly: flagResearchersOnly,
			}

			// LLM/automation path must not silently create a non-functional project.
			// If --plab-data is requested without --api-key, fail fast with a
			// structured error so the orchestrator can ask the user for the key.
			if project.UsePlabData && flagAPIKey == "" && AutoConfirm() {
				return cliError(
					"apikey_required",
					"플랩 API 키가 필요해요.",
					"--api-key 플래그로 플랩 API 키를 함께 전달해 주세요.",
					fmt.Sprintf("plab-app create --name %s --plab-data --api-key KEY --json", flagName),
				)
			}

			if !flagJSON {
				fmt.Println()
				fmt.Printf("  프로젝트: %s\n", tui.AccentStyle.Render(project.Name))
				fmt.Printf("  플랩 데이터: %s\n", tui.AccentStyle.Render(boolToKorean(project.UsePlabData)))
				fmt.Println()
			}
		} else {
			tui.PrintBanner(appVersion)

			var err error
			project, err = tui.RunCreateForm()
			if err != nil {
				if tui.IsUserAborted(err) {
					fmt.Println()
					fmt.Println(tui.DimStyle.Render("  종료했어요. 다시 시작하려면 plab-app create 를 입력해 주세요!"))
					fmt.Println()
					return nil
				}
				return fmt.Errorf("프로젝트 설정 중 오류: %w", err)
			}
		}

		cwd, err := os.Getwd()
		if err != nil {
			return cliError("cwd_failed", "현재 디렉토리 확인 실패", err.Error(), "")
		}
		outputDir := filepath.Join(cwd, project.Name)

		if _, err := os.Stat(outputDir); err == nil {
			return cliError(
				"dir_exists",
				fmt.Sprintf("%s 폴더가 이미 있어요.", project.Name),
				"다른 이름을 사용하거나 기존 폴더를 삭제해 주세요.",
				fmt.Sprintf("rm -rf %s", project.Name),
			)
		}

		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return cliError("mkdir_failed", "디렉토리 생성 실패", err.Error(), "")
		}

		gen := generator.New(*project, outputDir)
		if err := gen.Generate(); err != nil {
			return cliError("generate_failed", "프로젝트 생성 실패", err.Error(), "")
		}

		if project.UsePlabData {
			if flagAPIKey != "" {
				if err := writeAPIKey(outputDir, flagAPIKey); err != nil {
					return cliError("apikey_write_failed", "API 키 저장 실패", err.Error(), "")
				}
			} else if !flagJSON && flagName == "" {
				if err := tui.RunAPIKeySetup(outputDir); err != nil {
					return fmt.Errorf("API 키 설정 중 오류: %w", err)
				}
			}
		}

		postResult := generator.PostCreate(*project, outputDir)

		repoURL := tracking.GetRepoURL(outputDir)
		tracking.TrackProjectCreated(project.Name, repoURL, project.UsePlabData, project.ResearchersOnly)

		if flagJSON {
			steps := make([]map[string]interface{}, 0, len(postResult.Steps))
			for _, s := range postResult.Steps {
				entry := map[string]interface{}{
					"key":   s.Key,
					"label": s.Label,
					"ok":    s.OK,
				}
				if s.StderrHead != "" {
					entry["stderr_head"] = s.StderrHead
				}
				steps = append(steps, entry)
			}

			nextSteps := []map[string]string{
				{"label": "개발 서버 실행", "command": fmt.Sprintf("cd %s && plab-app dev", project.Name)},
				{"label": "프로덕션 배포", "command": fmt.Sprintf("cd %s && plab-app deploy --prod --json --yes", project.Name)},
			}

			result := map[string]interface{}{
				"success":          !postResult.Failed,
				"project":          project.Name,
				"path":             outputDir,
				"plab_data":        project.UsePlabData,
				"researchers_only": project.ResearchersOnly,
				"steps":            steps,
				"next_steps":       nextSteps,
			}
			if repoURL != "" {
				result["repo_url"] = repoURL
			}
			if missing := missingEnvVars(project, flagAPIKey); len(missing) > 0 {
				result["missing_env"] = missing
			}
			if postResult.Failed {
				result["warning"] = postResult.Err().Error()
			}
			PrintJSON(result)
			return nil
		}

		if postResult.Failed {
			fmt.Println()
			return nil
		}

		elapsed := time.Since(startTime).Round(time.Second)
		generator.PrintCompletionWithTime(*project, outputDir, elapsed)

		if flagName == "" {
			var runDev bool
			err := huh.NewConfirm().
				Title("바로 실행해 볼까요?").
				Description("개발 서버를 시작하고 브라우저를 열어요").
				Affirmative("네!").
				Negative("나중에 할게요").
				Value(&runDev).
				Run()
			if err == nil && runDev {
				devCmd := exec.Command("npm", "run", "dev")
				devCmd.Dir = outputDir
				devCmd.Stdout = os.Stdout
				devCmd.Stderr = os.Stderr
				devCmd.Stdin = os.Stdin

				go func() {
					time.Sleep(3 * time.Second)
					openBrowser("http://localhost:3000")
				}()

				return devCmd.Run()
			}
		}

		return nil
	},
}

func init() {
	createCmd.Flags().StringVarP(&flagName, "name", "n", "", "프로젝트 이름 (예: landing, admin)")
	createCmd.Flags().BoolVar(&flagPlabData, "plab-data", false, "플랩 데이터 연동 포함")
	createCmd.Flags().StringVar(&flagAPIKey, "api-key", "", "플랩 API 키 (--plab-data 와 함께 사용)")
	createCmd.Flags().BoolVar(&flagResearchersOnly, "researchers-only", false, "리서처 전용 (Google 로그인 포함)")
}

func boolToKorean(b bool) string {
	if b {
		return "예"
	}
	return "아니오"
}

func cliError(code, message, fix, command string) error {
	if flagJSON {
		PrintCLIError(code, message, fix, command)
		os.Exit(1)
	}
	if command != "" {
		return fmt.Errorf("%s\n  해결: %s\n  실행: %s", message, fix, command)
	}
	return fmt.Errorf("%s", message)
}

func writeAPIKey(projectDir, apiKey string) error {
	envContent := fmt.Sprintf("# 플랩 API 키\nPLAB_API_KEY=%s\nPLAB_API_URL=%s\n", apiKey, config.PlabAPIURL)
	return os.WriteFile(filepath.Join(projectDir, ".env.local"), []byte(envContent), 0644)
}

// missingEnvVars reports env vars whose values the LLM still needs to collect
// from the user before the generated project can actually run.
func missingEnvVars(project *model.Project, apiKey string) []string {
	var missing []string
	if project.UsePlabData && apiKey == "" {
		missing = append(missing, "PLAB_API_KEY")
	}
	if project.ResearchersOnly {
		missing = append(missing, "GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET", "NEXTAUTH_SECRET")
	}
	return missing
}
