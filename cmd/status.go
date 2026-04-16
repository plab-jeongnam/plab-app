package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "현재 프로젝트 상태를 확인해요",
	Long: `현재 디렉토리의 plab 프로젝트 상태를 한눈에 확인합니다.

점검 항목:
  - package.json 존재 여부
  - node_modules 설치 상태
  - 빌드 가능 여부 (npm run build)
  - Git 상태 (커밋되지 않은 변경, push 안 된 커밋)
  - GitHub 저장소 연결 상태

예시:
  cd plab-landing && plab-app status
  plab-app status --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, _ := os.Getwd()
		projectName := filepath.Base(cwd)

		checks := []statusCheck{
			checkPackageJSON(cwd),
			checkNodeModules(cwd),
			checkBuildable(cwd),
			checkGitStatus(cwd),
			checkGitRemote(cwd),
			checkUnpushed(cwd),
		}

		if flagJSON {
			results := make([]map[string]interface{}, len(checks))
			allOK := true
			for i, c := range checks {
				results[i] = map[string]interface{}{
					"name":   c.Name,
					"ok":     c.OK,
					"detail": c.Detail,
				}
				if !c.OK {
					allOK = false
					if c.Fix != "" {
						results[i]["fix"] = c.Fix
					}
				}
			}
			PrintJSON(map[string]interface{}{
				"success": allOK,
				"project": projectName,
				"path":    cwd,
				"checks":  results,
			})
			return nil
		}

		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
		okStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
		failStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

		fmt.Println()
		fmt.Printf("  %s %s\n", titleStyle.Render("프로젝트 상태"), dimStyle.Render(projectName))
		fmt.Println()

		allOK := true
		for _, c := range checks {
			if c.OK {
				fmt.Printf("  %s %-25s %s\n", okStyle.Render("✓"), c.Name, dimStyle.Render(c.Detail))
			} else {
				fmt.Printf("  %s %-25s %s\n", failStyle.Render("✗"), c.Name, failStyle.Render(c.Detail))
				if c.Fix != "" {
					fmt.Printf("    %s %s\n", dimStyle.Render("해결:"), dimStyle.Render(c.Fix))
				}
				allOK = false
			}
		}

		fmt.Println()
		if allOK {
			fmt.Println(okStyle.Render("  모든 상태가 정상이에요!"))
		} else {
			fmt.Println(dimStyle.Render("  문제가 있는 항목을 확인해 주세요."))
			fmt.Println(dimStyle.Render("  plab-app reset 으로 자동 복구할 수 있어요."))
		}
		fmt.Println()
		return nil
	},
}

type statusCheck struct {
	Name   string
	OK     bool
	Detail string
	Fix    string
}

func checkPackageJSON(dir string) statusCheck {
	_, err := os.Stat(filepath.Join(dir, "package.json"))
	if err != nil {
		return statusCheck{
			Name:   "package.json",
			OK:     false,
			Detail: "파일이 없어요",
			Fix:    "plab-app 프로젝트 디렉토리에서 실행해 주세요",
		}
	}
	return statusCheck{Name: "package.json", OK: true, Detail: "존재함"}
}

func checkNodeModules(dir string) statusCheck {
	info, err := os.Stat(filepath.Join(dir, "node_modules"))
	if err != nil || !info.IsDir() {
		return statusCheck{
			Name:   "패키지 설치",
			OK:     false,
			Detail: "node_modules 없음",
			Fix:    "npm install",
		}
	}
	return statusCheck{Name: "패키지 설치", OK: true, Detail: "설치됨"}
}

func checkBuildable(dir string) statusCheck {
	if _, err := os.Stat(filepath.Join(dir, "node_modules")); err != nil {
		return statusCheck{
			Name:   "빌드 가능",
			OK:     false,
			Detail: "패키지 미설치로 확인 불가",
			Fix:    "npm install 후 다시 확인해 주세요",
		}
	}

	cmd := exec.Command("npm", "run", "build")
	cmd.Dir = dir
	cmd.Stdout = nil
	cmd.Stderr = nil
	err := cmd.Run()
	if err != nil {
		return statusCheck{
			Name:   "빌드 가능",
			OK:     false,
			Detail: "빌드 실패",
			Fix:    "npm run build 로 에러를 확인해 주세요",
		}
	}
	return statusCheck{Name: "빌드 가능", OK: true, Detail: "성공"}
}

func checkGitStatus(dir string) statusCheck {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return statusCheck{
			Name:   "Git 상태",
			OK:     false,
			Detail: "Git 저장소가 아니에요",
			Fix:    "git init",
		}
	}
	output := strings.TrimSpace(out.String())
	if output != "" {
		lines := strings.Split(output, "\n")
		return statusCheck{
			Name:   "Git 상태",
			OK:     false,
			Detail: fmt.Sprintf("변경된 파일 %d개", len(lines)),
			Fix:    "git add . && git commit -m '변경사항 저장'",
		}
	}
	return statusCheck{Name: "Git 상태", OK: true, Detail: "깨끗함"}
}

func checkGitRemote(dir string) statusCheck {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return statusCheck{
			Name:   "GitHub 연결",
			OK:     false,
			Detail: "원격 저장소 없음",
			Fix:    "gh repo create <이름> --public --source=. --push",
		}
	}
	url := strings.TrimSpace(out.String())
	return statusCheck{Name: "GitHub 연결", OK: true, Detail: url}
}

func checkUnpushed(dir string) statusCheck {
	cmd := exec.Command("git", "log", "--oneline", "@{u}..HEAD")
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return statusCheck{Name: "최신 push", OK: true, Detail: "확인 불가 (원격 없음)"}
	}
	output := strings.TrimSpace(out.String())
	if output != "" {
		lines := strings.Split(output, "\n")
		return statusCheck{
			Name:   "최신 push",
			OK:     false,
			Detail: fmt.Sprintf("push 안 된 커밋 %d개", len(lines)),
			Fix:    "git push",
		}
	}
	return statusCheck{Name: "최신 push", OK: true, Detail: "최신 상태"}
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
