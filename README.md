# plab-app

비개발자도 쉽게 사용할 수 있는 웹 프로젝트 스캐폴딩 도구입니다.

## 설치

### 바이너리 다운로드 (Go 불필요)

GitHub Releases에서 OS에 맞는 바이너리를 다운로드합니다.

**macOS (Apple Silicon / M1~)**
```bash
curl -L https://github.com/plab-jeongnam/plab-app/releases/latest/download/plab-app-darwin-arm64 -o /usr/local/bin/plab-app
chmod +x /usr/local/bin/plab-app
```

**macOS (Intel)**
```bash
curl -L https://github.com/plab-jeongnam/plab-app/releases/latest/download/plab-app-darwin-amd64 -o /usr/local/bin/plab-app
chmod +x /usr/local/bin/plab-app
```

**Windows (PowerShell)**
```powershell
Invoke-WebRequest -Uri https://github.com/plab-jeongnam/plab-app/releases/latest/download/plab-app-windows-amd64.exe -OutFile plab-app.exe
Move-Item plab-app.exe "$env:LOCALAPPDATA\Microsoft\WindowsApps\plab-app.exe"
```

### Go가 설치되어 있는 경우

```bash
go install github.com/plab-jeongnam/plab-app@latest
```

### 설치 확인

```bash
plab-app version
```

## LLM을 위한 설치 가이드

LLM이 사용자 환경에 plab-app을 설치할 때 아래 순서를 따르세요.

```
1. OS 감지
   - uname -m (macOS/Linux) 또는 $env:PROCESSOR_ARCHITECTURE (Windows)

2. 바이너리 다운로드
   - macOS arm64: plab-app-darwin-arm64
   - macOS amd64: plab-app-darwin-amd64
   - Windows amd64: plab-app-windows-amd64.exe
   - Windows arm64: plab-app-windows-arm64.exe
   - Base URL: https://github.com/plab-jeongnam/plab-app/releases/latest/download/

3. PATH에 배치
   - macOS: /usr/local/bin/plab-app (chmod +x 필수)
   - Windows: $env:LOCALAPPDATA\Microsoft\WindowsApps\plab-app.exe

4. 환경 세팅 (원스텝)
   - plab-app setup
   - 필수 도구 자동 설치 + GitHub 로그인까지 안내

5. 프로젝트 생성 (비대화형 CLI 모드)
   - plab-app create --name <name> [--plab-data] [--researchers-only] [--api-key KEY] --json
   - --json 플래그로 결과를 파싱 가능한 JSON으로 받기

6. 에러 처리
   - 모든 --json 에러 응답에 code, fix, command 필드 포함
   - code별 자동 복구 가능 (예: dir_exists → rm -rf 후 재시도)
```

**macOS 원라이너 (LLM 복붙용):**
```bash
curl -L https://github.com/plab-jeongnam/plab-app/releases/latest/download/plab-app-darwin-arm64 -o /usr/local/bin/plab-app && chmod +x /usr/local/bin/plab-app && plab-app setup
```

**Windows 원라이너 (LLM 복붙용):**
```powershell
Invoke-WebRequest -Uri https://github.com/plab-jeongnam/plab-app/releases/latest/download/plab-app-windows-amd64.exe -OutFile "$env:LOCALAPPDATA\Microsoft\WindowsApps\plab-app.exe"; plab-app setup
```

## 시작하기

```bash
# 1. 환경 점검
plab-app doctor

# 2. 프로젝트 생성 (대화형)
plab-app create

# 3. 프로젝트 생성 (CLI)
plab-app create --name landing
plab-app create --name dashboard --plab-data
plab-app create --name internal-tool --researchers-only
plab-app create --name full --plab-data --researchers-only
```

## 명령어

| 명령어 | 설명 |
|--------|------|
| `create` | 새 프로젝트 생성 (TUI 대화형 또는 CLI 플래그) |
| `doctor` | 개발 환경 점검 (Git, Node.js, npm, gh, Vercel CLI, Claude Code) |
| `deploy` | Vercel 배포 (로그인 체크 + 빌드 검증 + 배포) |
| `dev` | 개발 서버 실행 + 브라우저 자동 오픈 |
| `open` | GitHub / Vercel / localhost 브라우저 열기 |
| `list` | 내 plab- 프로젝트 목록 (로컬 + GitHub) |
| `status` | 현재 프로젝트 상태 확인 (빌드/Git/배포) |
| `reset` | 프로젝트 복구 (node_modules 재설치 + 빌드 검증) |
| `update-template` | 공통 설정을 최신 템플릿으로 업데이트 |
| `version` | 버전 확인 |

## 생성되는 프로젝트

기본 포함:
- Next.js 16 + React 19 (App Router)
- TypeScript
- Tailwind CSS v4 + shadcn/ui
- Zustand (클라이언트 상태) + React Query (서버 상태)
- react-hook-form + zod (폼 + 검증)
- Vitest + Testing Library (테스트)
- ESLint + Prettier (코드 품질)
- Vercel 배포 설정
- robots.txt (Disallow 기본)
- CLAUDE.md + .skills (AI 코드 에이전트 가이드)

### 플랩 데이터 연동 (`--plab-data`)

| 파일 | 역할 |
|------|------|
| `src/lib/plab.ts` | 플랩 API 클라이언트 (query, tables) |
| `src/app/api/query/route.ts` | 서버사이드 API Route (키 노출 방지) |
| `src/app/plab-demo/page.tsx` | SQL 쿼리 데모 페이지 |
| `.env.local` | `PLAB_API_KEY`, `PLAB_API_URL` 환경변수 |

### 리서처 전용 모드 (`--researchers-only`)

모든 페이지에 Google 로그인이 필수가 됩니다.

| 파일 | 역할 |
|------|------|
| `src/middleware.ts` | 전체 페이지 로그인 강제 (/login, /api/auth만 예외) |
| `src/lib/auth/config.ts` | NextAuth + Google Provider 설정 |
| `src/lib/auth/types.ts` | Session 타입 확장 |
| `src/components/auth/nav-bar.tsx` | 상단 사용자 이름 + 로그아웃 버튼 |
| `src/components/auth/session-provider.tsx` | SessionProvider 래퍼 |
| `src/components/auth/auth-guard.tsx` | 페이지별 보호 컴포넌트 |
| `src/components/auth/sign-in-button.tsx` | 로그인/로그아웃 버튼 |
| `src/app/login/page.tsx` | 로그인 페이지 |
| `src/app/protected/page.tsx` | 보호 페이지 예제 |
| `.env.local` | Google OAuth + NEXTAUTH 환경변수 |
| `layout.tsx` | SessionProvider + NavBar 자동 래핑 |

흐름:
```
비로그인 → 아무 페이지 접근 → /login 리다이렉트
→ "Google로 로그인" 클릭 → Google OAuth
→ 로그인 성공 → 원래 페이지로 이동
→ 상단에 이름 + 로그아웃 버튼 항상 표시
→ 로그아웃 클릭 → /login으로 돌아감
```

## LLM / 자동화 가이드

모든 명령어에 `--json` 플래그를 지원합니다. LLM이 파싱하기 쉬운 구조화된 JSON을 반환합니다.

### 환경 점검

```bash
plab-app doctor --json
```
```json
{
  "ok": true,
  "platform": "darwin",
  "results": [
    {"name": "Git", "ok": true, "required": true, "version": "git version 2.46.0"},
    {"name": "Node.js", "ok": true, "required": true, "version": "v25.2.1"},
    {"name": "Vercel CLI", "ok": false, "required": false, "error": "미설치", "fix": "npm install -g vercel"}
  ]
}
```

### 프로젝트 생성

```bash
# 기본
plab-app create --name landing --json

# 플랩 데이터 + API 키
plab-app create --name api --plab-data --api-key plb_xxx --json

# 리서처 전용
plab-app create --name dash --researchers-only --json

# 전체 옵션
plab-app create --name full --plab-data --api-key plb_xxx --researchers-only --json
```
```json
{"success": true, "project": "plab-landing", "path": "/path/to/plab-landing", "plab_data": false}
```

### 에러 응답 형식

```json
{
  "error": "plab-landing 폴더가 이미 있어요.",
  "code": "dir_exists",
  "fix": "다른 이름을 사용하거나 기존 폴더를 삭제해 주세요.",
  "command": "rm -rf plab-landing"
}
```

에러 코드:
| code | 의미 |
|------|------|
| `required_tools_missing` | 필수 도구 미설치 |
| `dir_exists` | 폴더 중복 |
| `generate_failed` | 템플릿 생성 실패 |
| `build_failed` | 빌드 실패 |
| `not_project` | plab 프로젝트 디렉토리가 아님 |
| `vercel_not_found` | Vercel CLI 미설치 |
| `vercel_not_logged_in` | Vercel 미로그인 |
| `deploy_failed` | 배포 실패 |

### 프로젝트 상태 확인

```bash
cd plab-landing && plab-app status --json
```

### 프로젝트 복구

```bash
cd plab-landing && plab-app reset --force --json
```

### 배포

```bash
cd plab-landing && plab-app deploy --json         # 프리뷰
cd plab-landing && plab-app deploy --prod --json   # 프로덕션
```

### 프로젝트 목록

```bash
plab-app list --json
```
