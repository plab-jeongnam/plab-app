package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "개발 서버를 실행하고 브라우저를 열어요",
	Long: `개발 서버(npm run dev)를 실행하고 자동으로 브라우저를 열어요.

예시:
  cd plab-landing && plab-app dev`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, _ := os.Getwd()

		if _, err := os.Stat(filepath.Join(cwd, "package.json")); err != nil {
			return fmt.Errorf("package.json이 없어요. plab 프로젝트 디렉토리에서 실행해 주세요")
		}

		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		accentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("13"))

		fmt.Println()
		fmt.Println(accentStyle.Render("  개발 서버를 시작해요..."))
		fmt.Printf("  %s\n", dimStyle.Render("종료하려면 Control + C"))
		fmt.Println()

		go func() {
			time.Sleep(3 * time.Second)
			openBrowser("http://localhost:3000")
		}()

		devCmd := exec.Command("npm", "run", "dev")
		devCmd.Dir = cwd
		devCmd.Stdout = os.Stdout
		devCmd.Stderr = os.Stderr
		devCmd.Stdin = os.Stdin
		return devCmd.Run()
	},
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	cmd.Run()
}

func init() {
	rootCmd.AddCommand(devCmd)
}
