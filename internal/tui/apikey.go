package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/plab/plab-app/internal/config"
)

func RunAPIKeySetup(projectDir string) error {
	fmt.Println()
	fmt.Println(TitleStyle.Render("  플랩 데이터 연동 설정"))
	fmt.Println()
	fmt.Println(DimStyle.Render("  API 키가 있어야 플랩 데이터를 사용할 수 있어요."))
	fmt.Println(DimStyle.Render("  아래 사이트에서 발급받을 수 있어요:"))
	fmt.Println()
	fmt.Println(AccentStyle.Render("  " + config.PlabAPIURL))
	fmt.Println(DimStyle.Render("  → 로그인 → 대시보드 → API 키 → 새 API 키 생성"))
	fmt.Println()

	var hasKey bool
	err := huh.NewConfirm().
		Title("API 키를 발급받으셨나요?").
		Affirmative("네, 입력할게요").
		Negative("나중에 할게요").
		Value(&hasKey).
		Run()

	if err != nil {
		if IsUserAborted(err) {
			return nil
		}
		return err
	}

	if !hasKey {
		fmt.Println()
		fmt.Println(DimStyle.Render("  나중에 .env.local 파일에서 PLAB_API_KEY 를 설정해 주세요."))
		fmt.Println()
		return nil
	}

	var apiKey string
	err = huh.NewInput().
		Title("API 키를 입력해 주세요").
		Placeholder("plb_xxxxxxxxxxxxxxxx").
		Value(&apiKey).
		Validate(func(s string) error {
			s = strings.TrimSpace(s)
			if s == "" {
				return fmt.Errorf("API 키를 입력해 주세요")
			}
			return nil
		}).
		Run()

	if err != nil {
		if IsUserAborted(err) {
			return nil
		}
		return err
	}

	apiKey = strings.TrimSpace(apiKey)

	envContent := fmt.Sprintf(`# 플랩 API 키
PLAB_API_KEY=%s
PLAB_API_URL=https://vibe.techin.pe.kr
`, apiKey)

	envPath := filepath.Join(projectDir, ".env.local")
	if err := os.WriteFile(envPath, []byte(envContent), 0644); err != nil {
		return fmt.Errorf(".env.local 저장 실패: %w", err)
	}

	fmt.Println()
	fmt.Printf("  %s API 키가 .env.local 에 저장되었어요!\n", SuccessStyle.Render("✓"))
	fmt.Println()

	return nil
}
