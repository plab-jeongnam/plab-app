#!/bin/bash
# plab-app 설치 스크립트
# curl -fsSL https://raw.githubusercontent.com/plab-jeongnam/plab-app/main/install.sh | bash

set -e

REPO="plab-jeongnam/plab-app"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="plab-app"

# OS/아키텍처 감지
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "지원하지 않는 아키텍처: $ARCH"; exit 1 ;;
esac

case "$OS" in
  darwin) ASSET="${BINARY_NAME}-darwin-${ARCH}" ;;
  linux) ASSET="${BINARY_NAME}-linux-${ARCH}" ;;
  *) echo "지원하지 않는 OS: $OS"; exit 1 ;;
esac

# 최신 버전 확인
echo "최신 버전을 확인하고 있어요..."
LATEST=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST" ]; then
  echo "버전 확인에 실패했어요."
  exit 1
fi

echo "plab-app ${LATEST} 설치 중..."

# 다운로드
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST}/${ASSET}"
TMP_FILE=$(mktemp)

curl -fsSL "$DOWNLOAD_URL" -o "$TMP_FILE"
chmod +x "$TMP_FILE"

# 설치 (권한 필요 시 sudo)
if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
else
  echo "설치에 관리자 권한이 필요해요. 비밀번호를 입력해 주세요."
  sudo mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
fi

echo ""
echo "✓ plab-app ${LATEST} 설치 완료!"
echo ""
echo "  시작하려면:"
echo "  plab-app setup"
echo ""
