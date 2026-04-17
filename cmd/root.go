package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	appVersion = "dev"
	flagJSON   bool
	flagYes    bool
)

// AutoConfirm returns true when the user or LLM wants to skip interactive
// confirmations (either explicit --yes, or --json which implies automation).
func AutoConfirm() bool {
	return flagYes || flagJSON
}

func SetVersion(v string) {
	appVersion = v
}

var rootCmd = &cobra.Command{
	Use:   "plab-app",
	Short: "plab-app - 표준화된 웹 프로젝트 스캐폴딩 도구",
	Long: `plab-app - 표준화된 웹 프로젝트 스캐폴딩 도구

비개발자도 쉽게 표준 프로젝트를 생성할 수 있어요.
모든 프로젝트는 plab- 접두사가 자동으로 붙어요.

처음 사용하시나요?
  plab-app setup                           # 환경 세팅부터 시작

프로젝트 만들기:
  plab-app create                          # 대화형 프로젝트 생성
  plab-app create --name landing           # CLI로 plab-landing 생성
  plab-app create --name api --plab-data   # 플랩 데이터 연동 포함
  plab-app create --name dash --researchers-only  # 리서처 전용 (Google 로그인)

프로젝트 관리:
  plab-app dev                             # 개발 서버 실행
  plab-app deploy                          # Vercel 배포
  plab-app status                          # 상태 확인
  plab-app reset                           # 문제 복구

Exit Codes:
  0  성공
  1  오류 발생 (에러 메시지 참고)`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "출력을 JSON 형식으로 반환 (LLM/자동화용)")
	rootCmd.PersistentFlags().BoolVar(&flagYes, "yes", false, "모든 확인 질문을 자동 승인 (LLM/자동화용, --json 이면 자동 활성)")
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(versionCmd)
}

// PrintJSON outputs structured JSON for LLM/automation consumers.
func PrintJSON(v interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

// CLIError is a structured error response for LLM consumers.
type CLIError struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Fix     string `json:"fix"`
	Command string `json:"command,omitempty"`
}

func PrintCLIError(code, message, fix, command string) {
	if flagJSON {
		PrintJSON(CLIError{
			Error:   message,
			Code:    code,
			Fix:     fix,
			Command: command,
		})
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", message)
		if fix != "" {
			fmt.Fprintf(os.Stderr, "Fix: %s\n", fix)
		}
		if command != "" {
			fmt.Fprintf(os.Stderr, "Run: %s\n", command)
		}
	}
}
