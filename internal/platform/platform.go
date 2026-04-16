package platform

import "runtime"

type PackageManager string

const (
	Brew   PackageManager = "brew"
	Winget PackageManager = "winget"
)

type Platform struct {
	OS             string
	PackageManager PackageManager
}

func Detect() Platform {
	return ForOS(runtime.GOOS)
}

func ForOS(goos string) Platform {
	p := Platform{OS: goos}
	switch goos {
	case "darwin":
		p.PackageManager = Brew
	case "windows":
		p.PackageManager = Winget
	}
	return p
}

func (p Platform) InstallCommand(tool string) string {
	commands := map[PackageManager]map[string]string{
		Brew: {
			"git":    "brew install git",
			"node":   "brew install node",
			"gh":     "brew install gh",
			"vercel": "npm install -g vercel",
			"claude": "https://claude.ai/download 에서 데스크톱 앱 설치",
		},
		Winget: {
			"git":    "winget install --id Git.Git",
			"node":   "winget install --id OpenJS.NodeJS.LTS",
			"gh":     "winget install --id GitHub.cli",
			"vercel": "npm install -g vercel",
			"claude": "https://claude.ai/download 에서 데스크톱 앱 설치",
		},
	}

	if cmds, ok := commands[p.PackageManager]; ok {
		if cmd, ok := cmds[tool]; ok {
			return cmd
		}
	}
	return ""
}

func (p Platform) BinaryName() string {
	if p.OS == "windows" {
		return "plab-app.exe"
	}
	return "plab-app"
}
