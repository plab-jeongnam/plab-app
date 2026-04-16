package generator

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/plab/plab-app/internal/model"
	"github.com/plab/plab-app/internal/tui"
)

var (
	progressStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	hintStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Italic(true)
	errorHint     = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)

type step struct {
	label   string
	fn      func() (string, error) // (stderr output, error)
	onError func(stderr string)
}

func PostCreate(project model.Project, dir string) error {
	fmt.Println()
	fmt.Println(progressStyle.Render("  프로젝트를 만들고 있어요..."))
	fmt.Println()

	steps := []step{
		{
			label: "Git 초기화",
			fn: func() (string, error) {
				if _, err := runCapture(dir, "git", "init"); err != nil {
					return "", err
				}
				if _, err := runCapture(dir, "git", "add", "."); err != nil {
					return "", err
				}
				return runCapture(dir, "git", "commit", "-m", "chore: initialize "+project.Name)
			},
		},
		{
			label: "필요한 패키지 설치",
			fn: func() (string, error) {
				return runCapture(dir, "npm", "install")
			},
		},
		{
			label: "프로젝트 빌드 검증",
			fn: func() (string, error) {
				return runCapture(dir, "npm", "run", "build")
			},
			onError: func(stderr string) {
				fmt.Printf("    %s\n", errorHint.Render("빌드에 실패했어요. 에러를 확인해 주세요."))
				fmt.Printf("    %s\n", hintStyle.Render("cd "+project.Name+" && npm run build 로 직접 확인할 수 있어요."))
			},
		},
		{
			label: "GitHub 저장소 생성",
			fn: func() (string, error) {
				return runCapture(dir, "gh", "repo", "create", project.Name, "--private", "--source=.", "--push")
			},
			onError: func(stderr string) {
				printGitHubErrorGuide(project.Name, stderr)
			},
		},
	}

	hasError := false
	for _, s := range steps {
		sp := tui.NewSpinner(s.label)
		sp.Start()
		stderr, err := s.fn()
		sp.Stop(err == nil)
		if err != nil {
			hasError = true
			if s.onError != nil {
				s.onError(stderr)
			}
		}
	}

	if hasError {
		return fmt.Errorf("일부 작업이 실패했어요")
	}
	return nil
}

func printGitHubErrorGuide(projectName, stderr string) {
	lower := strings.ToLower(stderr)

	fmt.Println()
	switch {
	case strings.Contains(lower, "already exists"):
		fmt.Printf("    %s\n", errorHint.Render("같은 이름의 저장소가 이미 있어요."))
		fmt.Println()
		fmt.Printf("    %s\n", hintStyle.Render("해결 방법:"))
		fmt.Printf("    %s\n", hintStyle.Render("  1. 다른 프로젝트 이름으로 다시 만들거나"))
		fmt.Printf("    %s\n", hintStyle.Render("  2. 기존 저장소를 삭제 후 다시 시도해 주세요"))
		fmt.Printf("    %s\n", hintStyle.Render("     gh repo delete "+projectName+" --yes"))

	case strings.Contains(lower, "not logged") || strings.Contains(lower, "authentication"):
		fmt.Printf("    %s\n", errorHint.Render("GitHub에 로그인되어 있지 않아요."))
		fmt.Println()
		fmt.Printf("    %s\n", hintStyle.Render("해결 방법:"))
		fmt.Printf("    %s\n", hintStyle.Render("  1. gh auth login 을 실행해 주세요"))
		fmt.Printf("    %s\n", hintStyle.Render("  2. 로그인 후 아래 명령으로 저장소를 만들 수 있어요:"))
		fmt.Printf("    %s\n", hintStyle.Render("     cd "+projectName+" && gh repo create "+projectName+" --private --source=. --push"))

	case strings.Contains(lower, "permission") || strings.Contains(lower, "forbidden"):
		fmt.Printf("    %s\n", errorHint.Render("저장소를 만들 권한이 없어요."))
		fmt.Println()
		fmt.Printf("    %s\n", hintStyle.Render("해결 방법:"))
		fmt.Printf("    %s\n", hintStyle.Render("  1. GitHub 계정 권한을 확인해 주세요"))
		fmt.Printf("    %s\n", hintStyle.Render("  2. 조직(org) 저장소라면 관리자에게 권한을 요청해 주세요"))
		fmt.Printf("    %s\n", hintStyle.Render("  3. 개인 계정으로 만들려면:"))
		fmt.Printf("    %s\n", hintStyle.Render("     cd "+projectName+" && gh repo create "+projectName+" --private --source=. --push"))

	case strings.Contains(lower, "could not resolve") || strings.Contains(lower, "network"):
		fmt.Printf("    %s\n", errorHint.Render("인터넷 연결을 확인해 주세요."))
		fmt.Println()
		fmt.Printf("    %s\n", hintStyle.Render("해결 방법:"))
		fmt.Printf("    %s\n", hintStyle.Render("  1. Wi-Fi 또는 네트워크 연결 상태를 확인해 주세요"))
		fmt.Printf("    %s\n", hintStyle.Render("  2. 연결 후 아래 명령으로 다시 시도할 수 있어요:"))
		fmt.Printf("    %s\n", hintStyle.Render("     cd "+projectName+" && gh repo create "+projectName+" --private --source=. --push"))

	default:
		fmt.Printf("    %s\n", errorHint.Render("저장소 생성에 실패했어요."))
		if trimmed := strings.TrimSpace(stderr); trimmed != "" {
			fmt.Printf("    %s %s\n", hintStyle.Render("원인:"), trimmed)
		}
		fmt.Println()
		fmt.Printf("    %s\n", hintStyle.Render("프로젝트 파일은 정상적으로 만들어졌어요!"))
		fmt.Printf("    %s\n", hintStyle.Render("나중에 직접 저장소를 만들 수 있어요:"))
		fmt.Printf("    %s\n", hintStyle.Render("  cd "+projectName+" && gh repo create "+projectName+" --private --source=. --push"))
	}
	fmt.Println()
}

func PrintCompletion(project model.Project, dir string) {
	PrintCompletionWithTime(project, dir, 0)
}

func PrintCompletionWithTime(project model.Project, dir string, elapsed time.Duration) {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	accentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("13"))

	fmt.Println()
	if elapsed > 0 {
		fmt.Println(titleStyle.Render(fmt.Sprintf("  완료! (%s)", elapsed)))
	} else {
		fmt.Println(titleStyle.Render("  완료!"))
	}
	fmt.Println()
	fmt.Printf("  %s ./%s\n", dimStyle.Render("로컬:"), project.Name)
	fmt.Println()
	fmt.Println(dimStyle.Render("  시작하려면:"))
	fmt.Printf("  %s\n", accentStyle.Render(fmt.Sprintf("cd %s", project.Name)))
	fmt.Printf("  %s\n", accentStyle.Render("npm run dev"))
	fmt.Println()
}

func runCapture(dir string, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stdout = nil
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stderr.String(), err
}
