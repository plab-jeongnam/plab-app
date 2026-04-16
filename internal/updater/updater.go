package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/plab/plab-app/internal/config"
)

const (
	timeout = 2 * time.Second
)

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func CheckAndUpdate(currentVersion string) {
	if currentVersion == "dev" {
		return
	}

	latest, err := fetchLatestRelease()
	if err != nil {
		return
	}

	latestVersion := strings.TrimPrefix(latest.TagName, "v")
	current := strings.TrimPrefix(currentVersion, "v")

	if !isNewer(latestVersion, current) {
		return
	}

	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	fmt.Println()
	fmt.Println(infoStyle.Render(fmt.Sprintf("  새 버전이 있어요! %s → %s", currentVersion, latest.TagName)))

	var confirm bool
	err = huh.NewConfirm().
		Title("업데이트하시겠어요?").
		Affirmative("예").
		Negative("아니오").
		Value(&confirm).
		Run()

	if err != nil || !confirm {
		return
	}

	assetURL := findAssetURL(latest.Assets)
	if assetURL == "" {
		fmt.Println("  이 플랫폼에 맞는 바이너리를 찾을 수 없어요.")
		return
	}

	if err := downloadAndReplace(assetURL); err != nil {
		fmt.Printf("  업데이트 실패: %v\n", err)
		return
	}

	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	fmt.Println(successStyle.Render(fmt.Sprintf("  %s으로 업데이트 완료!", latest.TagName)))
	fmt.Println()
}

func isNewer(latest, current string) bool {
	lp := strings.Split(latest, ".")
	cp := strings.Split(current, ".")

	for i := 0; i < len(lp) && i < len(cp); i++ {
		l, r := 0, 0
		fmt.Sscanf(lp[i], "%d", &l)
		fmt.Sscanf(cp[i], "%d", &r)
		if l > r {
			return true
		}
		if l < r {
			return false
		}
	}
	return false
}

func fetchLatestRelease() (*githubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", config.GitHubOwner, config.GitHubRepo)

	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	return &release, nil
}

func findAssetURL(assets []githubAsset) string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	for _, asset := range assets {
		name := strings.ToLower(asset.Name)
		if strings.Contains(name, goos) && strings.Contains(name, goarch) {
			return asset.BrowserDownloadURL
		}
	}
	return ""
}

func downloadAndReplace(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("다운로드 실패: %w", err)
	}
	defer resp.Body.Close()

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("실행 파일 경로 확인 실패: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("심볼릭 링크 확인 실패: %w", err)
	}

	tmpFile, err := os.CreateTemp(filepath.Dir(execPath), "plab-app-update-*")
	if err != nil {
		return fmt.Errorf("임시 파일 생성 실패: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		return fmt.Errorf("파일 쓰기 실패: %w", err)
	}
	tmpFile.Close()

	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		return fmt.Errorf("권한 설정 실패: %w", err)
	}

	if err := os.Rename(tmpFile.Name(), execPath); err != nil {
		return fmt.Errorf("바이너리 교체 실패: %w", err)
	}

	return nil
}
