package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/plab/plab-app/internal/doctor"
	"github.com/plab/plab-app/internal/platform"
	"github.com/plab/plab-app/internal/tui"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "개발 환경을 처음부터 세팅해요",
	Long: `개발 환경을 처음부터 단계별로 세팅합니다.

비개발자를 위한 원스텝 온보딩이에요.
필요한 도구를 자동으로 확인하고, 없는 것은 설치를 도와줘요.

동작:
  1. 현재 환경 점검
  2. 누락된 필수 도구 설치 안내 + 자동 설치
  3. GitHub 로그인 안내
  4. 완료 후 plab-app create 안내

예시:
  plab-app setup                               # 대화형
  plab-app setup --yes                         # 모든 확인 자동 승인
  plab-app setup --json                        # JSON 결과 (LLM/자동화용, --yes 포함)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		plat := platform.Detect()
		auto := AutoConfirm()

		if flagJSON {
			return runSetupJSON(plat)
		}

		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
		okMark := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("✓")
		failMark := lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("✗")
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		accentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("13"))

		tui.PrintBanner(appVersion)

		fmt.Println(titleStyle.Render("  개발 환경 세팅을 시작할게요!"))
		fmt.Println(dimStyle.Render("  하나씩 확인하면서 없는 것은 설치해 드릴게요."))
		fmt.Println()

		// Step 1: 패키지 매니저 확인
		fmt.Println(titleStyle.Render("  1. 패키지 매니저 확인"))
		fmt.Println()

		if plat.OS == "darwin" {
			if _, err := exec.LookPath("brew"); err != nil {
				fmt.Printf("  %s Homebrew가 설치되어 있지 않아요.\n", failMark)
				fmt.Println()
				fmt.Println(dimStyle.Render("  Homebrew는 macOS에서 프로그램을 쉽게 설치하는 도구에요."))
				fmt.Println(dimStyle.Render("  아래 명령을 복사해서 터미널에 붙여넣기 해주세요:"))
				fmt.Println()
				fmt.Println(accentStyle.Render(`  /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`))
				fmt.Println()
				fmt.Println(dimStyle.Render("  설치가 끝나면 터미널을 껐다 켜고 plab-app setup 을 다시 실행해 주세요."))
				fmt.Println()
				return nil
			}
			fmt.Printf("  %s Homebrew 설치됨\n", okMark)
		} else {
			fmt.Printf("  %s Windows 환경 (winget 사용)\n", okMark)
		}
		fmt.Println()

		// Step 2: 필수 도구 확인 + 설치
		tools := setupTools(plat)

		fmt.Println(titleStyle.Render("  2. 필수 도구 확인"))
		fmt.Println()

		needsInstall := []toolCheck{}
		for _, t := range tools {
			if _, err := exec.LookPath(t.binary); err != nil {
				fmt.Printf("  %s %s — 설치 필요\n", failMark, t.name)
				needsInstall = append(needsInstall, t)
			} else {
				fmt.Printf("  %s %s\n", okMark, t.name)
			}
		}
		fmt.Println()

		if len(needsInstall) > 0 {
			doInstall := auto
			if !auto {
				err := huh.NewConfirm().
					Title(fmt.Sprintf("%d개 도구를 설치할까요?", len(needsInstall))).
					Description("자동으로 설치할 수 있어요.").
					Affirmative("네, 설치해 주세요").
					Negative("아니오, 직접 할게요").
					Value(&doInstall).
					Run()

				if err != nil {
					if tui.IsUserAborted(err) {
						return nil
					}
					return err
				}
			}

			if doInstall {
				fmt.Println()
				for _, t := range needsInstall {
					fmt.Printf("  %s %s 설치 중...\n", dimStyle.Render("⏳"), t.name)
					installCmd := buildInstallCommand(plat.OS, t.installCmd)
					c := exec.Command(installCmd[0], installCmd[1:]...)
					c.Stdout = os.Stdout
					c.Stderr = os.Stderr
					if err := c.Run(); err != nil {
						fmt.Printf("\033[1A\033[2K")
						fmt.Printf("  %s %s 설치 실패\n", failMark, t.name)
						fmt.Printf("    %s 직접 설치해 주세요: %s\n", dimStyle.Render("→"), t.installCmd)
					} else {
						fmt.Printf("\033[1A\033[2K")
						fmt.Printf("  %s %s 설치 완료\n", okMark, t.name)
					}
				}
				fmt.Println()
			} else {
				fmt.Println()
				fmt.Println(dimStyle.Render("  아래 명령을 하나씩 실행해 주세요:"))
				fmt.Println()
				for _, t := range needsInstall {
					fmt.Printf("  %s\n", accentStyle.Render(t.installCmd))
				}
				fmt.Println()
				fmt.Println(dimStyle.Render("  설치 후 plab-app setup 을 다시 실행해 주세요."))
				fmt.Println()
				return nil
			}
		}

		// Step 3: GitHub 로그인
		fmt.Println(titleStyle.Render("  3. GitHub 로그인"))
		fmt.Println()

		report := doctor.Run()
		ghAuthOK := false
		for _, r := range report.Results {
			if r.Name == "GitHub 인증" && r.OK {
				ghAuthOK = true
			}
		}

		if ghAuthOK {
			fmt.Printf("  %s GitHub 로그인 완료\n", okMark)
		} else {
			fmt.Printf("  %s GitHub에 로그인되어 있지 않아요.\n", failMark)
			fmt.Println()

			if _, err := exec.LookPath("gh"); err == nil {
				fmt.Println(dimStyle.Render("  지금 GitHub 로그인을 시작할게요."))
				fmt.Println(dimStyle.Render("  브라우저가 열리면 로그인해 주세요."))
				fmt.Println()

				doLogin := auto
				if !auto {
					err := huh.NewConfirm().
						Title("GitHub 로그인을 시작할까요?").
						Description("GitHub 계정이 없다면 https://github.com/signup 에서 먼저 가입해 주세요.").
						Affirmative("네, 로그인할게요").
						Negative("나중에 할게요").
						Value(&doLogin).
						Run()
					if err != nil {
						doLogin = false
					}
				}

				if doLogin {
					loginCmd := exec.Command("gh", "auth", "login", "--web")
					loginCmd.Stdin = os.Stdin
					loginCmd.Stdout = os.Stdout
					loginCmd.Stderr = os.Stderr
					loginCmd.Run()
				}
			} else {
				fmt.Println(dimStyle.Render("  gh CLI 설치 후 아래 명령을 실행해 주세요:"))
				fmt.Println(accentStyle.Render("  gh auth login"))
			}
		}
		fmt.Println()

		// Step 4: 최종 확인
		fmt.Println(titleStyle.Render("  4. 최종 확인"))
		fmt.Println()

		finalReport := doctor.Run()
		finalReport.Print()

		if finalReport.RequiredPassed() {
			okStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
			fmt.Println(okStyle.Render("  세팅 완료! 이제 프로젝트를 만들 수 있어요."))
			fmt.Println()
			fmt.Println(dimStyle.Render("  프로젝트를 만들려면:"))
			fmt.Printf("  %s\n", accentStyle.Render("plab-app create"))
			fmt.Println()
		} else {
			fmt.Println(tui.ErrorStyle.Render("  아직 설정이 필요한 항목이 있어요."))
			fmt.Println(dimStyle.Render("  위의 안내를 따라 설치 후 plab-app setup 을 다시 실행해 주세요."))
			fmt.Println()
		}

		return nil
	},
}

type toolCheck struct {
	name       string
	binary     string
	installCmd string
	required   bool
}

func setupTools(plat platform.Platform) []toolCheck {
	return []toolCheck{
		{"Git", "git", plat.InstallCommand("git"), true},
		{"Node.js", "node", plat.InstallCommand("node"), true},
		{"GitHub CLI", "gh", plat.InstallCommand("gh"), true},
	}
}

// runSetupJSON performs setup in non-interactive mode and emits structured JSON
// for LLMs/automation. Brew missing is treated as a blocking error because
// piping curl|bash under automation is unsafe — we return the install command
// and let the orchestrator surface it to the user.
func runSetupJSON(plat platform.Platform) error {
	result := map[string]interface{}{
		"success":  false,
		"platform": plat.OS,
	}

	if plat.OS == "darwin" {
		if _, err := exec.LookPath("brew"); err != nil {
			PrintCLIError(
				"brew_required",
				"Homebrew가 설치되어 있지 않아요.",
				"아래 명령을 터미널에 복사해 실행한 뒤 plab-app setup 을 다시 실행해 주세요.",
				`/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`,
			)
			os.Exit(1)
		}
	}

	tools := setupTools(plat)
	installed := []string{}
	failed := []map[string]string{}
	missing := []string{}

	for _, t := range tools {
		if _, err := exec.LookPath(t.binary); err == nil {
			continue
		}
		installCmd := buildInstallCommand(plat.OS, t.installCmd)
		c := exec.Command(installCmd[0], installCmd[1:]...)
		if err := c.Run(); err != nil {
			failed = append(failed, map[string]string{
				"name":        t.name,
				"binary":      t.binary,
				"install_cmd": t.installCmd,
				"error":       err.Error(),
			})
			missing = append(missing, t.name)
		} else {
			installed = append(installed, t.name)
		}
	}

	report := doctor.Run()
	ghAuthOK := false
	for _, r := range report.Results {
		if r.Name == "GitHub 인증" && r.OK {
			ghAuthOK = true
		}
	}

	result["installed"] = installed
	result["missing"] = missing
	if len(failed) > 0 {
		result["install_failed"] = failed
	}
	result["gh_auth"] = ghAuthOK

	if !ghAuthOK {
		result["gh_auth_command"] = "gh auth login --web"
		result["requires_user_action"] = true
		result["user_action_reason"] = "GitHub 로그인은 브라우저 상호작용이 필요해요. 유저에게 'gh auth login --web' 실행을 안내해 주세요."
	}

	if report.RequiredPassed() && ghAuthOK {
		result["success"] = true
		result["next_command"] = "plab-app create --json"
	}

	PrintJSON(result)
	if !result["success"].(bool) {
		os.Exit(1)
	}
	return nil
}

func buildInstallCommand(goos, installCmd string) []string {
	if goos == "darwin" {
		return []string{"sh", "-c", installCmd}
	}
	if runtime.GOOS == "windows" {
		return []string{"cmd", "/c", installCmd}
	}
	return []string{"sh", "-c", installCmd}
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
