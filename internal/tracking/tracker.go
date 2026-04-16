package tracking

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const (
	apiURL  = "https://vibe.techin.pe.kr/api/plab-app/events"
	timeout = 3 * time.Second
)

var appVersion = "dev"

func SetVersion(v string) {
	appVersion = v
}

// TrackProjectCreated sends a project.created event (non-blocking).
func TrackProjectCreated(projectName string, repoURL string, usePlabData, researchersOnly bool) {
	go sendEvent(map[string]interface{}{
		"event":        "project.created",
		"gh_username":  getGHUsername(),
		"project_name": projectName,
		"repo_url":     repoURL,
		"options": map[string]bool{
			"plab_data":        usePlabData,
			"researchers_only": researchersOnly,
		},
		"plab_app_version": appVersion,
		"platform":         runtime.GOOS,
	})
}

// TrackProjectDeployed sends a project.deployed event (non-blocking).
func TrackProjectDeployed(projectName, repoURL, deployURL, deployType string) {
	go sendEvent(map[string]interface{}{
		"event":            "project.deployed",
		"gh_username":      getGHUsername(),
		"project_name":     projectName,
		"repo_url":         repoURL,
		"deploy_url":       deployURL,
		"deploy_type":      deployType,
		"build_success":    true,
		"plab_app_version": appVersion,
		"platform":         runtime.GOOS,
	})
}

func sendEvent(data map[string]interface{}) {
	jsonBody, err := json.Marshal(data)
	if err != nil {
		return
	}

	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(jsonBody))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
}

func getGHUsername() string {
	cmd := exec.Command("gh", "api", "user", "--jq", ".login")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return "unknown"
	}
	return strings.TrimSpace(out.String())
}

func getRepoURL(dir string) string {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return ""
	}
	return strings.TrimSpace(out.String())
}

// GetRepoURL is exported for use in cmd packages.
func GetRepoURL(dir string) string {
	return getRepoURL(dir)
}
