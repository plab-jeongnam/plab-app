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
	"github.com/plab/plab-app/internal/tui"
	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open [target]",
	Short: "프로젝트 관련 페이지를 브라우저로 열어요",
	Long: `프로젝트 관련 페이지를 브라우저에서 열어요.

대상:
  github   GitHub 저장소 페이지
  vercel   Vercel 대시보드
  dev      로컬 개발 서버 (localhost:3000)

예시:
  plab-app open             # 대화형 선택
  plab-app open github      # GitHub 바로 열기
  plab-app open vercel
  plab-app open dev`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, _ := os.Getwd()

		if _, err := os.Stat(filepath.Join(cwd, "package.json")); err != nil {
			return fmt.Errorf("package.json이 없어요. plab 프로젝트 디렉토리에서 실행해 주세요")
		}

		target := ""
		if len(args) > 0 {
			target = args[0]
		}

		if target == "" {
			var selected string
			err := huh.NewSelect[string]().
				Title("어디를 열까요?").
				Options(
					huh.NewOption("GitHub 저장소", "github"),
					huh.NewOption("Vercel 대시보드", "vercel"),
					huh.NewOption("로컬 개발 서버", "dev"),
				).
				Value(&selected).
				Run()
			if err != nil {
				if tui.IsUserAborted(err) {
					return nil
				}
				return err
			}
			target = selected
		}

		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

		switch target {
		case "github":
			url := getGitRemoteURL(cwd)
			if url == "" {
				return fmt.Errorf("GitHub 원격 저장소가 연결되어 있지 않아요. gh repo create 로 먼저 만들어 주세요")
			}
			url = toHTTPSURL(url)
			fmt.Printf("  %s %s\n", dimStyle.Render("열기:"), url)
			openBrowser(url)

		case "vercel":
			projectName := filepath.Base(cwd)
			url := fmt.Sprintf("https://vercel.com/dashboard?search=%s", projectName)
			fmt.Printf("  %s %s\n", dimStyle.Render("열기:"), url)
			openBrowser(url)

		case "dev":
			fmt.Printf("  %s %s\n", dimStyle.Render("열기:"), "http://localhost:3000")
			openBrowser("http://localhost:3000")

		default:
			return fmt.Errorf("알 수 없는 대상: %s (github, vercel, dev 중 선택해 주세요)", target)
		}

		return nil
	},
}

func getGitRemoteURL(dir string) string {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return ""
	}
	return strings.TrimSpace(out.String())
}

func toHTTPSURL(gitURL string) string {
	if strings.HasPrefix(gitURL, "https://") {
		return gitURL
	}
	// git@github.com:user/repo.git → https://github.com/user/repo
	url := strings.TrimPrefix(gitURL, "git@")
	url = strings.Replace(url, ":", "/", 1)
	url = strings.TrimSuffix(url, ".git")
	return "https://" + url
}

func init() {
	rootCmd.AddCommand(openCmd)
}
