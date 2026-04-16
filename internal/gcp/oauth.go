package gcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
)

// 플랩 공용 GCP OAuth 설정
// 빌드 시 ldflags로 주입됩니다. GitHub Actions secrets에서 관리.
var (
	GCPProjectNumber    = ""  // -X github.com/plab/plab-app/internal/gcp.GCPProjectNumber=...
	OAuthClientID       = ""  // -X github.com/plab/plab-app/internal/gcp.OAuthClientID=...
	OAuthClientSecret   = ""  // -X github.com/plab/plab-app/internal/gcp.OAuthClientSecret=...
)

const callbackPath = "/api/auth/callback/google"

type oauthClientResponse struct {
	RedirectURIs []string `json:"redirectUris"`
}

// AddRedirectURI adds a Vercel deployment URL as an OAuth redirect URI.
// Returns nil if successful, error with fallback instructions if failed.
func AddRedirectURI(deployURL string) error {
	redirectURI := strings.TrimSuffix(deployURL, "/") + callbackPath

	token, err := getGcloudToken()
	if err != nil {
		return fmt.Errorf("gcloud_not_available")
	}

	// Get current redirect URIs
	current, err := getCurrentRedirectURIs(token)
	if err != nil {
		return fmt.Errorf("api_failed: %w", err)
	}

	// Check if already exists
	for _, uri := range current {
		if uri == redirectURI {
			return nil // already registered
		}
	}

	// Add new URI
	updated := append(current, redirectURI)
	if err := updateRedirectURIs(token, updated); err != nil {
		return fmt.Errorf("update_failed: %w", err)
	}

	return nil
}

// RedirectURIForURL returns the callback URL for a given deployment URL.
func RedirectURIForURL(deployURL string) string {
	return strings.TrimSuffix(deployURL, "/") + callbackPath
}

func getGcloudToken() (string, error) {
	cmd := exec.Command("gcloud", "auth", "print-access-token")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

func getCurrentRedirectURIs(token string) ([]string, error) {
	url := fmt.Sprintf(
		"https://oauth2.googleapis.com/v1/projects/%s/oauthClients/%s",
		GCPProjectNumber, OAuthClientID,
	)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API %d: %s", resp.StatusCode, string(body))
	}

	var result oauthClientResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.RedirectURIs, nil
}

func updateRedirectURIs(token string, uris []string) error {
	url := fmt.Sprintf(
		"https://oauth2.googleapis.com/v1/projects/%s/oauthClients/%s",
		GCPProjectNumber, OAuthClientID,
	)

	body := map[string]interface{}{
		"redirectUris": uris,
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("PATCH", url, bytes.NewReader(jsonBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
