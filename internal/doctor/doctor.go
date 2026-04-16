package doctor

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/plab/plab-app/internal/platform"
)

type CheckResult struct {
	Name     string
	OK       bool
	Version  string
	Message  string
	Required bool
}

type Report struct {
	Results  []CheckResult
	Platform platform.Platform
}

func (r Report) RequiredPassed() bool {
	for _, result := range r.Results {
		if result.Required && !result.OK {
			return false
		}
	}
	return true
}

func (r Report) AllPassed() bool {
	for _, result := range r.Results {
		if !result.OK {
			return false
		}
	}
	return true
}

func (r Report) FailedRequiredCount() int {
	count := 0
	for _, result := range r.Results {
		if result.Required && !result.OK {
			count++
		}
	}
	return count
}

type JSONReport struct {
	OK       bool         `json:"ok"`
	Platform string       `json:"platform"`
	Results  []JSONResult `json:"results"`
}

type JSONResult struct {
	Name     string `json:"name"`
	OK       bool   `json:"ok"`
	Required bool   `json:"required"`
	Version  string `json:"version,omitempty"`
	Error    string `json:"error,omitempty"`
	Fix      string `json:"fix,omitempty"`
}

func (r Report) ToJSON() JSONReport {
	results := make([]JSONResult, len(r.Results))
	for i, result := range r.Results {
		jr := JSONResult{
			Name:     result.Name,
			OK:       result.OK,
			Required: result.Required,
		}
		if result.OK {
			jr.Version = result.Version
		} else {
			jr.Error = result.Message
			jr.Fix = r.Platform.InstallCommand(toolKey(result.Name))
		}
		results[i] = jr
	}
	return JSONReport{
		OK:       r.RequiredPassed(),
		Platform: r.Platform.OS,
		Results:  results,
	}
}

func (r Report) FailedOptionalCount() int {
	count := 0
	for _, result := range r.Results {
		if !result.Required && !result.OK {
			count++
		}
	}
	return count
}

func Run() Report {
	return RunWithPlatform(platform.Detect())
}

func RunWithPlatform(plat platform.Platform) Report {
	results := []CheckResult{
		withRequired(checkGit(), true),
		withRequired(checkNode(), true),
		withRequired(checkNpm(), true),
		withRequired(checkGhCLI(), true),
		withRequired(checkGhAuth(), true),
		withRequired(checkVercelCLI(), false),
		withRequired(checkClaudeCode(), false),
	}
	return Report{Results: results, Platform: plat}
}

func withRequired(r CheckResult, required bool) CheckResult {
	r.Required = required
	return r
}

func (r Report) Print() {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	okStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	failStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	fmt.Println()
	fmt.Println(titleStyle.Render("  개발 환경 점검 중..."))
	fmt.Println()

	for _, result := range r.Results {
		tag := ""
		if !result.Required {
			tag = dimStyle.Render(" (권장)")
		}

		if result.OK {
			fmt.Printf("  %s %-20s %s\n",
				okStyle.Render("✓"),
				result.Name+tag,
				dimStyle.Render(result.Version),
			)
		} else {
			mark := failStyle.Render("✗")
			msg := failStyle.Render(result.Message)
			if !result.Required {
				mark = warnStyle.Render("△")
				msg = warnStyle.Render(result.Message)
			}
			fmt.Printf("  %s %-20s %s\n", mark, result.Name+tag, msg)
		}
	}

	fmt.Println()

	failedRequired := r.FailedRequiredCount()
	failedOptional := r.FailedOptionalCount()

	if r.AllPassed() {
		fmt.Println(okStyle.Render("  모든 환경이 준비되었어요!"))
	} else {
		if failedRequired > 0 {
			fmt.Printf(failStyle.Render("  %d개 필수 항목을 설정해야 해요.\n"), failedRequired)
		}
		if failedOptional > 0 {
			fmt.Printf(warnStyle.Render("  %d개 권장 항목이 없지만, 프로젝트 생성은 가능해요.\n"), failedOptional)
		}
		fmt.Println()

		hasGhFailed := false
		for _, result := range r.Results {
			if result.OK {
				continue
			}
			key := toolKey(result.Name)
			installCmd := r.Platform.InstallCommand(key)
			if installCmd != "" {
				label := "설치:"
				if !result.Required {
					label = "권장:"
				}
				fmt.Printf("  %s %s\n", dimStyle.Render(label), installCmd)
			}
			if key == "gh" {
				hasGhFailed = true
			}
		}

		if hasGhFailed {
			fmt.Println()
			fmt.Println(dimStyle.Render("  GitHub 계정이 없다면 먼저 가입해 주세요:"))
			fmt.Println(dimStyle.Render("  https://github.com/signup"))
			fmt.Println(dimStyle.Render("  가입 후 gh CLI를 설치하고 gh auth login 을 실행해 주세요."))
		}
	}
	fmt.Println()
}

func checkGit() CheckResult {
	return checkTool("Git", "git", "--version")
}

func checkNode() CheckResult {
	result := checkTool("Node.js", "node", "--version")
	if result.OK {
		ver := strings.TrimPrefix(result.Version, "v")
		parts := strings.Split(ver, ".")
		if len(parts) > 0 {
			major := parts[0]
			if major < "18" {
				return CheckResult{
					Name:    "Node.js",
					OK:      false,
					Message: fmt.Sprintf("v18 이상 필요 (현재: v%s)", ver),
				}
			}
		}
	}
	return result
}

func checkNpm() CheckResult {
	return checkTool("npm", "npm", "--version")
}

func checkGhCLI() CheckResult {
	return checkTool("gh (GitHub CLI)", "gh", "--version")
}

func checkVercelCLI() CheckResult {
	return checkTool("Vercel CLI", "vercel", "--version")
}

func checkClaudeCode() CheckResult {
	return checkTool("Claude Code", "claude", "--version")
}

func checkGhAuth() CheckResult {
	cmd := exec.Command("gh", "auth", "status")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return CheckResult{
			Name:    "GitHub 인증",
			OK:      false,
			Message: "미인증 (gh auth login 필요)",
		}
	}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Logged in") {
			return CheckResult{
				Name:    "GitHub 인증",
				OK:      true,
				Version: "인증됨",
			}
		}
	}
	return CheckResult{
		Name:    "GitHub 인증",
		OK:      true,
		Version: "인증됨",
	}
}

func checkTool(name, binary string, args ...string) CheckResult {
	cmd := exec.Command(binary, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return CheckResult{
			Name:    name,
			OK:      false,
			Message: "미설치",
		}
	}
	version := strings.TrimSpace(string(output))
	if strings.Contains(version, "\n") {
		version = strings.Split(version, "\n")[0]
	}
	return CheckResult{
		Name:    name,
		OK:      true,
		Version: version,
	}
}

func toolKey(name string) string {
	switch name {
	case "Git":
		return "git"
	case "Node.js":
		return "node"
	case "npm":
		return "npm"
	case "gh (GitHub CLI)":
		return "gh"
	case "Vercel CLI":
		return "vercel"
	case "Claude Code":
		return "claude"
	}
	return ""
}
