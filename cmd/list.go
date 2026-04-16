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

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "plab 프로젝트 목록을 보여줘요",
	Long: `현재 디렉토리에서 plab- 접두사를 가진 프로젝트 목록을 보여줍니다.
GitHub에서 내 plab 저장소도 함께 조회합니다.

예시:
  plab-app list
  plab-app list --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, _ := os.Getwd()

		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
		okStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		accentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("13"))

		// 로컬 프로젝트 찾기
		localProjects := findLocalProjects(cwd)

		// GitHub 프로젝트 찾기
		ghProjects := findGitHubProjects()

		if flagJSON {
			PrintJSON(map[string]interface{}{
				"local":  localProjects,
				"github": ghProjects,
			})
			return nil
		}

		fmt.Println()

		// 로컬 프로젝트
		fmt.Println(titleStyle.Render("  로컬 프로젝트"))
		fmt.Println()
		if len(localProjects) == 0 {
			fmt.Println(dimStyle.Render("  plab- 프로젝트가 없어요."))
			fmt.Println(dimStyle.Render("  plab-app create 로 만들어 보세요!"))
		} else {
			for _, p := range localProjects {
				hasPackageJSON := ""
				if _, err := os.Stat(filepath.Join(cwd, p, "package.json")); err == nil {
					hasPackageJSON = okStyle.Render("✓")
				} else {
					hasPackageJSON = dimStyle.Render("·")
				}
				fmt.Printf("  %s %s\n", hasPackageJSON, accentStyle.Render(p))
			}
		}

		fmt.Println()

		// GitHub 프로젝트
		fmt.Println(titleStyle.Render("  GitHub 저장소"))
		fmt.Println()
		if len(ghProjects) == 0 {
			fmt.Println(dimStyle.Render("  plab- 저장소가 없어요."))
		} else {
			for _, p := range ghProjects {
				fmt.Printf("  %s %s\n", dimStyle.Render("·"), accentStyle.Render(p))
			}
		}

		fmt.Println()
		return nil
	},
}

func findLocalProjects(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var projects []string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "plab-") {
			projects = append(projects, entry.Name())
		}
	}
	return projects
}

func findGitHubProjects() []string {
	cmd := exec.Command("gh", "repo", "list", "--limit", "50", "--json", "name", "--jq", ".[].name")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return nil
	}

	var projects []string
	for _, line := range strings.Split(strings.TrimSpace(out.String()), "\n") {
		name := strings.TrimSpace(line)
		if name != "" && strings.HasPrefix(name, "plab-") {
			projects = append(projects, name)
		}
	}
	return projects
}

func init() {
	rootCmd.AddCommand(listCmd)
}
