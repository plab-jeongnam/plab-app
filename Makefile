VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# GCP OAuth secrets - 로컬 빌드 시 환경변수로 전달
# export PLAB_GCP_PROJECT_NUMBER=996753315925
# export PLAB_OAUTH_CLIENT_ID=996753315925-xxx.apps.googleusercontent.com
# export PLAB_OAUTH_CLIENT_SECRET=GOCSPX-xxx
GCP_PKG = github.com/plab/plab-app/internal/gcp

LDFLAGS = -ldflags "\
  -X main.version=$(VERSION) \
  -X $(GCP_PKG).GCPProjectNumber=$(PLAB_GCP_PROJECT_NUMBER) \
  -X $(GCP_PKG).OAuthClientID=$(PLAB_OAUTH_CLIENT_ID) \
  -X $(GCP_PKG).OAuthClientSecret=$(PLAB_OAUTH_CLIENT_SECRET)"

.PHONY: build build-all clean test

build:
	go build $(LDFLAGS) -o plab-app .

build-all: clean
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/plab-app-darwin-arm64 .
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/plab-app-darwin-amd64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/plab-app-windows-amd64.exe .
	GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o dist/plab-app-windows-arm64.exe .

clean:
	rm -rf dist/ plab-app plab-app.exe

test:
	go vet ./...
	go build ./...
