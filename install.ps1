# plab-app Windows 설치 스크립트
# PowerShell: irm https://raw.githubusercontent.com/plab-jeongnam/plab-app/main/install.ps1 | iex

$ErrorActionPreference = "Stop"

$Repo = "plab-jeongnam/plab-app"
$BinaryName = "plab-app.exe"

# 아키텍처 감지
$Arch = if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }
$Asset = "plab-app-windows-${Arch}.exe"

# 최신 버전 확인
Write-Host "최신 버전을 확인하고 있어요..."
$Release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
$Version = $Release.tag_name

Write-Host "plab-app $Version 설치 중..."

# 다운로드
$DownloadUrl = "https://github.com/$Repo/releases/download/$Version/$Asset"
$InstallDir = "$env:LOCALAPPDATA\Microsoft\WindowsApps"
$InstallPath = Join-Path $InstallDir $BinaryName

Invoke-WebRequest -Uri $DownloadUrl -OutFile $InstallPath

Write-Host ""
Write-Host "✓ plab-app $Version 설치 완료!" -ForegroundColor Green
Write-Host ""
Write-Host "  시작하려면:"
Write-Host "  plab-app setup"
Write-Host ""
