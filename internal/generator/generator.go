package generator

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/plab/plab-app/internal/gcp"
	"github.com/plab/plab-app/internal/model"
)

//go:embed all:templates
var templateFS embed.FS

type Generator struct {
	Project   model.Project
	OutputDir string
}

func New(project model.Project, outputDir string) *Generator {
	return &Generator{
		Project:   project,
		OutputDir: outputDir,
	}
}

func (g *Generator) Generate() error {
	templateDir := "templates/nextjs-web/files"
	if err := g.walkAndRender(templateDir, g.OutputDir); err != nil {
		return err
	}

	if g.Project.ResearchersOnly {
		if err := g.injectAuthDependency(); err != nil {
			return fmt.Errorf("next-auth 의존성 추가 실패: %w", err)
		}
		if err := g.injectOAuthSecrets(); err != nil {
			return fmt.Errorf("OAuth 시크릿 주입 실패: %w", err)
		}
	}

	return nil
}

func (g *Generator) injectOAuthSecrets() error {
	envPath := filepath.Join(g.OutputDir, ".env.local")
	data, err := os.ReadFile(envPath)
	if err != nil {
		return nil // .env.local이 없으면 건너뜀
	}

	content := string(data)
	content = strings.ReplaceAll(content, "__PLAB_OAUTH_CLIENT_ID__", gcp.OAuthClientID)
	content = strings.ReplaceAll(content, "__PLAB_OAUTH_CLIENT_SECRET__", gcp.OAuthClientSecret)

	return os.WriteFile(envPath, []byte(content), 0644)
}

func (g *Generator) injectAuthDependency() error {
	pkgPath := filepath.Join(g.OutputDir, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return err
	}

	var pkg map[string]interface{}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return err
	}

	deps, ok := pkg["dependencies"].(map[string]interface{})
	if !ok {
		deps = make(map[string]interface{})
	}
	deps["next-auth"] = "^4.24.0"
	pkg["dependencies"] = deps

	updated, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(pkgPath, append(updated, '\n'), 0644)
}

func (g *Generator) walkAndRender(srcDir, destDir string) error {
	entries, err := templateFS.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("템플릿 디렉토리 읽기 실패: %w", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())
		destName := g.replaceVars(entry.Name())
		destPath := filepath.Join(destDir, destName)

		if entry.IsDir() {
			if g.shouldSkipDir(entry.Name()) {
				continue
			}
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return fmt.Errorf("디렉토리 생성 실패: %w", err)
			}
			if err := g.walkAndRender(srcPath, destPath); err != nil {
				return err
			}
			continue
		}

		if g.shouldSkipFile(srcPath) {
			continue
		}

		content, err := templateFS.ReadFile(srcPath)
		if err != nil {
			return fmt.Errorf("파일 읽기 실패 (%s): %w", srcPath, err)
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("디렉토리 생성 실패: %w", err)
		}

		var output string
		if strings.HasSuffix(entry.Name(), ".tmpl") {
			output, err = g.renderTemplate(string(content), destName)
			if err != nil {
				return fmt.Errorf("템플릿 렌더링 실패 (%s): %w", srcPath, err)
			}
		} else {
			output = string(content)
		}

		if err := os.WriteFile(destPath, []byte(output), 0644); err != nil {
			return fmt.Errorf("파일 쓰기 실패 (%s): %w", destPath, err)
		}
	}

	return nil
}

func (g *Generator) renderTemplate(content, filename string) (string, error) {
	tmpl, err := template.New(filename).
		Delims("<%", "%>").
		Parse(content)
	if err != nil {
		return "", fmt.Errorf("템플릿 파싱 실패: %w", err)
	}

	data := map[string]interface{}{
		"ProjectName":     g.Project.Name,
		"DisplayName":     g.Project.DisplayName,
		"UsePlabData":     g.Project.UsePlabData,
		"ResearchersOnly": g.Project.ResearchersOnly,
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("템플릿 실행 실패: %w", err)
	}

	return buf.String(), nil
}

func (g *Generator) replaceVars(name string) string {
	name = strings.TrimSuffix(name, ".tmpl")
	return name
}

func (g *Generator) shouldSkipDir(name string) bool {
	if !g.Project.ResearchersOnly {
		authDirs := []string{"[...nextauth]", "auth", "login"}
		for _, d := range authDirs {
			if name == d {
				return true
			}
		}
	}
	return false
}

func (g *Generator) shouldSkipFile(path string) bool {
	if !g.Project.UsePlabData {
		// plab.ts는 유지할 수도 있지만 사용처가 없으면 제거
		if strings.Contains(path, "plab.ts") {
			return true
		}
	}
	if !g.Project.ResearchersOnly {
		authFiles := []string{"/auth/", "/login/", "session-provider", "sign-in-button", "auth-guard", "nav-bar", "middleware.ts"}
		for _, f := range authFiles {
			if strings.Contains(path, f) {
				return true
			}
		}
	}
	return false
}
