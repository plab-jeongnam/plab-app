package generator

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type EnvVar struct {
	Key   string
	Value string
}

// ReadEnvLocal reads .env.local and returns non-empty, non-comment variables.
func ReadEnvLocal(dir string) ([]EnvVar, error) {
	path := filepath.Join(dir, ".env.local")
	f, err := os.Open(path)
	if err != nil {
		return nil, nil // no .env.local is fine
	}
	defer f.Close()

	var vars []EnvVar
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if value == "" || strings.Contains(value, "placeholder") {
			continue // skip empty or placeholder values
		}
		vars = append(vars, EnvVar{Key: key, Value: value})
	}
	return vars, nil
}

// SyncEnvToVercel checks which env vars are missing on Vercel and adds them.
// Returns (synced count, error).
func SyncEnvToVercel(dir string, vars []EnvVar) (int, error) {
	if len(vars) == 0 {
		return 0, nil
	}

	// Get existing Vercel env vars
	existing := getVercelEnvVars(dir)

	synced := 0
	for _, v := range vars {
		if _, ok := existing[v.Key]; ok {
			continue // already set
		}
		// Add to Vercel
		cmd := exec.Command("vercel", "env", "add", v.Key, "production", "preview", "development")
		cmd.Dir = dir
		cmd.Stdin = strings.NewReader(v.Value)
		cmd.Stdout = nil
		cmd.Stderr = nil
		if err := cmd.Run(); err == nil {
			synced++
		}
	}

	return synced, nil
}

func getVercelEnvVars(dir string) map[string]bool {
	cmd := exec.Command("vercel", "env", "ls")
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return nil
	}

	result := make(map[string]bool)
	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// vercel env ls output has variable names as first word
		fields := strings.Fields(line)
		if len(fields) > 0 {
			result[fields[0]] = true
		}
	}
	return result
}

// PrintEnvSyncReport prints the env sync result for non-JSON mode.
func PrintEnvSyncReport(vars []EnvVar, synced int) {
	okStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11"))

	if len(vars) == 0 {
		return
	}

	fmt.Println()
	fmt.Println(dimStyle.Render("  환경변수 동기화:"))
	if synced > 0 {
		fmt.Printf("  %s %d개 환경변수를 Vercel에 설정했어요\n", okStyle.Render("✓"), synced)
	} else {
		fmt.Printf("  %s 모든 환경변수가 이미 설정되어 있어요\n", okStyle.Render("✓"))
	}

	fmt.Println()
	fmt.Println(dimStyle.Render("  동기화된 변수:"))
	for _, v := range vars {
		masked := v.Value
		if len(masked) > 6 {
			masked = masked[:3] + "***" + masked[len(masked)-3:]
		} else {
			masked = "***"
		}
		fmt.Printf("    %s = %s\n", warnStyle.Render(v.Key), dimStyle.Render(masked))
	}
	fmt.Println()
}
