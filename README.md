# plab-app

비개발자도 쉽게 사용할 수 있는 웹 프로젝트 스캐폴딩 도구입니다.

<!-- TODO: 데모 GIF 추가 (터미널 녹화 후 교체) -->
<!-- ![plab-app demo](./demo.gif) -->

> 👋 **코드를 몰라도 괜찮아요.** AI에게 "이 레포로 프로젝트 만들어줘"라고 말하기만 하면 돼요. → **[비개발자 가이드 보기](./docs/비개발자-가이드.md)**

## 설치

### macOS / Linux (원라이너)

```bash
curl -fsSL https://raw.githubusercontent.com/plab-jeongnam/plab-app/main/install.sh | bash
```

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/plab-jeongnam/plab-app/main/install.ps1 | iex
```

### 설치 확인

```bash
plab-app version
```

> OS/아키텍처를 자동 감지하고 최신 버전을 `/usr/local/bin`(macOS) 또는 `WindowsApps`(Windows)에 설치합니다.

## LLM을 위한 설치 가이드

LLM은 비개발자 유저의 자연어 요청(예: "이 레포로 프로젝트 만들어줘")을 받아 이 CLI를 `--json --yes` 모드로 자동 실행합니다. 유저에게는 아래 "최소 질문"만 물어보세요.

### 최소 질문 (유저에게 물어볼 것)

1. **프로젝트 이름이 뭐예요?** → `--name`
2. **플랩 데이터를 쓸 건가요?** (예/아니오) → `--plab-data`
   - 예인 경우에만: **플랩 API 키를 알려주세요** → `--api-key`
3. **리서처만 사용하나요?** (예/아니오) → `--researchers-only`

### 표준 플로우 (무인 실행)

```
1. OS 감지
   - uname -m (macOS/Linux) 또는 $env:PROCESSOR_ARCHITECTURE (Windows)

2. 바이너리 다운로드 + PATH 배치
   - Base URL: https://github.com/plab-jeongnam/plab-app/releases/latest/download/
   - macOS arm64/amd64: plab-app-darwin-{arch}      → /usr/local/bin/plab-app (chmod +x)
   - Windows amd64/arm64: plab-app-windows-{arch}.exe → $env:LOCALAPPDATA\Microsoft\WindowsApps\plab-app.exe

3. 환경 세팅
   - plab-app setup --json
   - 도구 자동 설치 + 상태 JSON 반환
   - gh_auth: false 이면 requires_user_action: true → 유저에게 'gh auth login --web' 안내

4. 프로젝트 생성
   - plab-app create --name <name> [--plab-data] [--researchers-only] [--api-key KEY] --json
   - 응답의 next_steps[] 를 그대로 실행하면 다음 단계 진행 가능

5. 배포 (선택)
   - cd <project> && plab-app deploy --prod --json --yes
   - oauth.requires_user_action: true 이면 유저에게 console_url 안내
```

### 모든 커맨드 공통 플래그

- `--json` : 구조화된 JSON 응답 (`--yes` 자동 포함)
- `--yes`  : 모든 확인 질문을 자동 승인 (JSON 없이도 사용 가능)

### 에러 코드 → LLM 대응

| code | 의미 | LLM 대응 |
|------|------|---------|
| `brew_required` | macOS에 Homebrew 없음 | `command` 필드의 설치 명령을 **유저에게 복붙 안내** (curl\|bash는 무인 실행 X) |
| `required_tools_missing` | 필수 도구 부족 | `plab-app setup --json --yes` 로 재시도 |
| `apikey_required` | `--plab-data` 인데 API 키 없음 | **유저에게 플랩 API 키를 물어보고** 재시도 |
| `dir_exists` | 프로젝트 폴더 중복 | 다른 이름 제안 또는 `rm -rf` 후 재시도 (유저 동의 필수) |
| `generate_failed` | 템플릿 생성 실패 | `stderr_head` 확인 후 보고 |
| `build_failed` | `npm run build` 실패 | `cd <project> && npm run build` 로그 확인 |
| `vercel_not_found` | Vercel CLI 없음 | `npm install -g vercel` 자동 실행 또는 안내 |
| `vercel_not_logged_in` | Vercel 미로그인 | `requires_user_action: true` — 유저에게 `vercel login` 안내 |
| `deploy_failed` | 배포 실패 | `stderr` 분석 후 보고 |
| `not_project` | plab 프로젝트 폴더 아님 | `cd <project>` 한 뒤 재시도 |

### `requires_user_action: true` 필드

브라우저·계정 상호작용이 필수인 단계 (GitHub 로그인, Vercel 로그인, Google OAuth 등록)에서 이 필드가 `true`로 붙습니다. LLM은 무인 실행을 중단하고, `user_action_reason` 메시지를 유저에게 그대로 전달하세요.

**macOS / Linux (LLM 복붙용):**
```bash
curl -fsSL https://raw.githubusercontent.com/plab-jeongnam/plab-app/main/install.sh | bash && plab-app setup --json
```

**Windows (LLM 복붙용):**
```powershell
irm https://raw.githubusercontent.com/plab-jeongnam/plab-app/main/install.ps1 | iex; plab-app setup --json
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
| `setup` | 개발 환경 원스텝 세팅 (도구 설치 + GitHub 로그인) |
| `create` | 새 프로젝트 생성 (TUI 대화형 또는 CLI 플래그) |
| `doctor` | 개발 환경 점검 (`--fix`로 자동 설치) |
| `deploy` | Vercel 배포 (환경변수 동기화 + 빌드 검증 + 배포) |
| `dev` | 개발 서버 실행 + 브라우저 자동 오픈 |
| `open` | GitHub / Vercel / localhost 브라우저 열기 |
| `list` | 내 plab- 프로젝트 목록 (로컬 + GitHub) |
| `status` | 현재 프로젝트 상태 확인 (빌드/Git/배포) |
| `reset` | 프로젝트 복구 (node_modules 재설치 + 빌드 검증) |
| `update-template` | 공통 설정을 최신 템플릿으로 업데이트 |
| `upgrade` | plab-app 자체를 최신 버전으로 업데이트 |
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
