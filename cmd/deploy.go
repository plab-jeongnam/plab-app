package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/plab/plab-app/internal/gcp"
	"github.com/plab/plab-app/internal/generator"
	"github.com/plab/plab-app/internal/tracking"
	"github.com/plab/plab-app/internal/tui"
	"github.com/spf13/cobra"
)

var flagProd bool

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Vercel로 프로젝트를 배포해요",
	Long: `Vercel로 프로젝트를 배포합니다.

사전 점검:
  1. Vercel CLI 설치 여부
  2. Vercel 로그인 상태
  3. 빌드 검증 (npm run build)

동작:
  1. 사전 점검 통과 후 Vercel 배포
  2. 배포 URL 출력

예시:
  cd plab-landing && plab-app deploy           # 프리뷰 배포
  plab-app deploy --prod                       # 프로덕션 배포
  plab-app deploy --json                       # JSON 결과`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, _ := os.Getwd()

		if _, err := os.Stat(filepath.Join(cwd, "package.json")); err != nil {
			if flagJSON {
				PrintCLIError("not_project", "plab 프로젝트 디렉토리가 아니에요.", "package.json이 있는 프로젝트 폴더에서 실행해 주세요.", "")
				os.Exit(1)
			}
			return fmt.Errorf("package.json이 없어요. plab 프로젝트 디렉토리에서 실행해 주세요")
		}

		// Step 1: Vercel CLI 체크
		if err := checkVercelInstalled(); err != nil {
			return err
		}

		// Step 2: Vercel 로그인 체크
		if err := checkVercelLogin(); err != nil {
			return err
		}

		// Step 3: 배포 확인
		deployType := "프리뷰"
		if flagProd {
			deployType = "프로덕션"
		}

		if !AutoConfirm() {
			var confirm bool
			err := huh.NewConfirm().
				Title(fmt.Sprintf("%s 배포를 시작할까요?", deployType)).
				Affirmative("네, 배포할게요").
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

		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

		// Step 4: 환경변수 동기화
		envVars, _ := generator.ReadEnvLocal(cwd)
		if len(envVars) > 0 && !flagJSON {
			fmt.Println()
			fmt.Println(dimStyle.Render("  .env.local 환경변수를 Vercel에 동기화할게요."))
			fmt.Println(dimStyle.Render("  (비어있거나 placeholder 값은 건너뛰어요)"))
			fmt.Println()

			sp := tui.NewSpinner("환경변수 동기화")
			sp.Start()
			synced, _ := generator.SyncEnvToVercel(cwd, envVars)
			sp.Stop(true)
			generator.PrintEnvSyncReport(envVars, synced)
		}

		if !flagJSON {
			fmt.Println(dimStyle.Render(fmt.Sprintf("  %s 배포를 시작해요...", deployType)))
			fmt.Println()
		}

		// Step 5: 빌드 검증
		buildSp := tui.NewSpinner("빌드 검증")
		if !flagJSON {
			buildSp.Start()
		}
		buildCmd := exec.Command("npm", "run", "build")
		buildCmd.Dir = cwd
		buildCmd.Stdout = nil
		buildCmd.Stderr = nil
		if err := buildCmd.Run(); err != nil {
			if !flagJSON {
				buildSp.Stop(false)
				fmt.Println(tui.ErrorStyle.Render("    빌드에 실패했어요. npm run build로 에러를 확인해 주세요."))
			} else {
				PrintCLIError("build_failed", "빌드 실패", "npm run build 로 에러를 확인해 주세요.", "npm run build")
				os.Exit(1)
			}
			return nil
		}
		if !flagJSON {
			buildSp.Stop(true)
		}

		// Step 6: Vercel 배포
		deploySp := tui.NewSpinner("Vercel 배포")
		if !flagJSON {
			deploySp.Start()
		}

		vercelArgs := []string{"deploy"}
		if flagProd {
			vercelArgs = append(vercelArgs, "--prod")
		}
		vercelArgs = append(vercelArgs, "--yes")

		vercelCmd := exec.Command("vercel", vercelArgs...)
		vercelCmd.Dir = cwd
		var outBuf, errBuf bytes.Buffer
		vercelCmd.Stdout = &outBuf
		vercelCmd.Stderr = &errBuf

		if err := vercelCmd.Run(); err != nil {
			stderr := errBuf.String()
			if !flagJSON {
				deploySp.Stop(false)
				fmt.Println()
				printDeployErrorGuide(stderr)
			} else {
				PrintCLIError("deploy_failed", "Vercel 배포 실패", stderr, "vercel deploy")
				os.Exit(1)
			}
			return nil
		}

		deployURL := strings.TrimSpace(outBuf.String())

		projectName := filepath.Base(cwd)
		repoURL := tracking.GetRepoURL(cwd)
		tracking.TrackProjectDeployed(projectName, repoURL, deployURL, deployType)

		if flagJSON {
			result := map[string]interface{}{
				"success": true,
				"url":     deployURL,
				"type":    deployType,
				"next_steps": []map[string]string{
					{"label": "브라우저에서 확인", "command": "plab-app open"},
					{"label": "상태 확인", "command": "plab-app status --json"},
				},
			}
			if hasResearchersOnly(cwd) {
				redirectURI := gcp.RedirectURIForURL(deployURL)
				oauthErr := gcp.AddRedirectURI(deployURL)
				oauth := map[string]interface{}{
					"required":     true,
					"redirect_uri": redirectURI,
					"registered":   oauthErr == nil,
				}
				if oauthErr != nil {
					oauth["requires_user_action"] = true
					oauth["console_url"] = "https://console.cloud.google.com/apis/credentials"
					oauth["user_action_reason"] = "Google Cloud Console에서 OAuth 2.0 클라이언트의 '승인된 리디렉션 URI'에 redirect_uri 값을 추가해 주세요."
				}
				result["oauth"] = oauth
				// Backward-compatible flat fields (do not remove).
				result["oauth_redirect_uri"] = redirectURI
				result["oauth_redirect_registered"] = oauthErr == nil
			}
			PrintJSON(result)
			return nil
		}

		deploySp.Stop(true)

		// Step 7: Google OAuth redirect URI 자동 등록 (researchers-only 프로젝트)
		if hasResearchersOnly(cwd) {
			redirectURI := gcp.RedirectURIForURL(deployURL)

			oauthSp := tui.NewSpinner("Google OAuth redirect URI 등록")
			oauthSp.Start()
			err := gcp.AddRedirectURI(deployURL)
			if err != nil {
				oauthSp.Stop(false)
				errMsg := err.Error()
				if strings.Contains(errMsg, "gcloud_not_available") {
					fmt.Println()
					fmt.Println(dimStyle.Render("    gcloud CLI가 없어서 자동 등록을 건너뛰었어요."))
					fmt.Println(dimStyle.Render("    Google Cloud Console에서 직접 추가해 주세요:"))
					fmt.Println()
					fmt.Println(dimStyle.Render("    1. https://console.cloud.google.com/apis/credentials"))
					fmt.Println(dimStyle.Render("    2. OAuth 2.0 클라이언트 ID 클릭"))
					fmt.Println(dimStyle.Render("    3. 승인된 리디렉션 URI에 추가:"))
					fmt.Printf("       %s\n", tui.AccentStyle.Render(redirectURI))
					fmt.Println()
				} else {
					fmt.Println()
					fmt.Println(dimStyle.Render("    자동 등록에 실패했어요. 직접 추가해 주세요:"))
					fmt.Printf("       %s\n", tui.AccentStyle.Render(redirectURI))
					fmt.Println()
				}
			} else {
				oauthSp.Stop(true)
			}
		}

		fmt.Println()

		okStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
		accentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("13"))
		fmt.Println(okStyle.Render(fmt.Sprintf("  %s 배포 완료!", deployType)))
		fmt.Println()
		fmt.Printf("  %s %s\n", dimStyle.Render("URL:"), accentStyle.Render(deployURL))
		fmt.Println()
		return nil
	},
}

func checkVercelInstalled() error {
	if _, err := exec.LookPath("vercel"); err != nil {
		if flagJSON {
			PrintCLIError("vercel_not_found", "Vercel CLI가 설치되어 있지 않아요.", "npm install -g vercel 로 설치해 주세요.", "npm install -g vercel")
			os.Exit(1)
		}
		fmt.Println()
		fmt.Println(tui.ErrorStyle.Render("  Vercel CLI가 설치되어 있지 않아요."))
		fmt.Println()
		fmt.Println(tui.DimStyle.Render("  설치 방법:"))
		fmt.Println(tui.DimStyle.Render("  npm install -g vercel"))
		fmt.Println()
		os.Exit(1)
	}
	return nil
}

func checkVercelLogin() error {
	cmd := exec.Command("vercel", "whoami")
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err := cmd.Run()

	if err != nil || strings.TrimSpace(out.String()) == "" {
		if flagJSON {
			PrintJSON(map[string]interface{}{
				"success":              false,
				"error":                "Vercel에 로그인되어 있지 않아요.",
				"code":                 "vercel_not_logged_in",
				"fix":                  "vercel login 으로 로그인해 주세요. (브라우저 상호작용 필요)",
				"command":              "vercel login",
				"requires_user_action": true,
				"user_action_reason":   "Vercel 로그인은 브라우저에서 직접 수행해야 해요. 유저에게 터미널에서 'vercel login' 실행을 안내해 주세요.",
			})
			os.Exit(1)
		}
		fmt.Println()
		fmt.Println(tui.ErrorStyle.Render("  Vercel에 로그인되어 있지 않아요."))
		fmt.Println()
		fmt.Println(tui.DimStyle.Render("  Vercel 계정이 없다면:"))
		fmt.Println(tui.DimStyle.Render("  1. https://vercel.com/signup 에서 가입해 주세요"))
		fmt.Println(tui.DimStyle.Render("     (GitHub 계정으로 간편 가입 가능)"))
		fmt.Println()
		fmt.Println(tui.DimStyle.Render("  이미 계정이 있다면:"))
		fmt.Println(tui.DimStyle.Render("  2. vercel login 을 실행해 주세요"))
		fmt.Println(tui.DimStyle.Render("  3. 로그인 후 plab-app deploy 를 다시 실행해 주세요"))
		fmt.Println()
		os.Exit(1)
	}

	return nil
}

func printDeployErrorGuide(stderr string) {
	lower := strings.ToLower(stderr)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Italic(true)
	errorHintLocal := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))

	switch {
	case strings.Contains(lower, "not linked"):
		fmt.Println(errorHintLocal.Render("    이 프로젝트가 Vercel에 연결되어 있지 않아요."))
		fmt.Println()
		fmt.Println(dimStyle.Render("    해결 방법:"))
		fmt.Println(dimStyle.Render("    vercel link 를 실행해서 프로젝트를 연결해 주세요."))

	case strings.Contains(lower, "payment") || strings.Contains(lower, "billing"):
		fmt.Println(errorHintLocal.Render("    Vercel 요금제 관련 문제가 있어요."))
		fmt.Println()
		fmt.Println(dimStyle.Render("    해결 방법:"))
		fmt.Println(dimStyle.Render("    https://vercel.com/dashboard 에서 결제 정보를 확인해 주세요."))

	case strings.Contains(lower, "build failed") || strings.Contains(lower, "build error"):
		fmt.Println(errorHintLocal.Render("    Vercel 서버에서 빌드가 실패했어요."))
		fmt.Println()
		fmt.Println(dimStyle.Render("    해결 방법:"))
		fmt.Println(dimStyle.Render("    1. 환경 변수가 Vercel에 설정되어 있는지 확인해 주세요"))
		fmt.Println(dimStyle.Render("    2. https://vercel.com/dashboard 에서 빌드 로그를 확인해 주세요"))

	default:
		fmt.Println(errorHintLocal.Render("    배포에 실패했어요."))
		fmt.Println()
		fmt.Println(dimStyle.Render("    해결 방법:"))
		fmt.Println(dimStyle.Render("    1. vercel deploy 로 직접 에러를 확인해 주세요"))
		fmt.Println(dimStyle.Render("    2. 문제가 계속되면 vercel login 으로 다시 로그인해 주세요"))
	}
	fmt.Println()
}

func hasResearchersOnly(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "src", "middleware.ts"))
	return err == nil
}

func init() {
	deployCmd.Flags().BoolVar(&flagProd, "prod", false, "프로덕션 배포")
	rootCmd.AddCommand(deployCmd)
}
