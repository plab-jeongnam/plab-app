# CLAUDE.md - plab-app

Go TUI 프로젝트 스캐폴딩 도구. 비개발자가 표준화된 Next.js 웹 프로젝트를 생성할 수 있게 해줌.

## 프로젝트 구조

```
cmd/
  setup.go              # 원스텝 온보딩 (도구 설치 + GitHub 로그인)
  create.go             # 프로젝트 생성 (TUI + CLI, --name --plab-data --researchers-only --api-key)
  doctor.go             # 환경 진단 (--fix로 자동 설치, --simulate-os)
  deploy.go             # Vercel 배포 (--prod, 환경변수 동기화 + OAuth redirect URI)
  dev.go                # npm run dev + 브라우저 오픈
  open.go               # GitHub/Vercel/localhost 브라우저 열기
  list.go               # plab- 프로젝트 목록 (로컬 + GitHub)
  status.go             # 프로젝트 상태 (빌드/Git/배포)
  reset.go              # 프로젝트 복구 (--force)
  update_template.go    # 공통 설정 최신화 (--force)
  upgrade.go            # 자체 업데이트 (GitHub Releases)
  version.go            # 버전 확인
  root.go               # cobra root, --json 글로벌 플래그

internal/
  config/               # 중앙 설정 (PlabAPIURL, GitHubOwner 등)
  doctor/               # 환경 진단 (Git, Node.js, npm, gh, Vercel CLI, Claude Code)
  gcp/                  # Google OAuth redirect URI 자동 등록
  generator/            # 템플릿 렌더링 + 파일 생성 + post-create + next-auth 주입
    templates/          # Go embed 템플릿 (Next.js 16 + shadcn/ui + Tailwind v4)
  model/                # Project 모델 (Name, UsePlabData, ResearchersOnly)
  platform/             # OS별 로직 (macOS: brew / Windows: winget)
  tui/                  # huh 폼 + lipgloss 스타일 + 배너 + API 키 입력
  updater/              # 자동 업데이트 (GitHub Releases API)
main.go
```

## 기술 스택

- Go 1.22+ / cobra / bubbletea / huh / lipgloss
- 템플릿 엔진: `text/template` (delimiters: `<%` `%>`)
- 템플릿 파일: `.tmpl` 확장자만 렌더링, 나머지는 그대로 복사
- 바이너리에 embed (`//go:embed all:templates`)

## 핵심 규칙

- 모든 생성 프로젝트는 `plab-` prefix 강제
- npm 사용 (bun 아님) — 비개발자 대상
- `.npmrc`에 `legacy-peer-deps=true` 포함 (React 19 peer dep 호환)
- doctor 필수 항목: Git, Node.js, npm, gh CLI, GitHub 인증
- doctor 권장 항목: Vercel CLI, Claude Code
- `--json` 플래그로 모든 명령의 구조화된 JSON 출력 지원
- 크로스 플랫폼: macOS (brew) + Windows (winget)

## 조건부 기능

| 플래그 | 포함되는 파일 | 추가 의존성 |
|--------|--------------|------------|
| `--plab-data` | `lib/plab.ts`, `api/query/route.ts`, `plab-demo/page.tsx`, `.env.local`(API키) | 없음 |
| `--researchers-only` | `middleware.ts`, `lib/auth/*`, `components/auth/*`, `login/page.tsx`, `protected/page.tsx`, `.env.local`(OAuth) | `next-auth` |

ResearchersOnly=true 시:
- `middleware.ts`가 모든 페이지를 로그인 필수로 강제 (/login, /api/auth 예외)
- `layout.tsx`에 SessionProvider + NavBar 자동 래핑
- `next-auth` 의존성이 package.json에 후처리로 주입됨

## 빌드 & 테스트

```bash
go build -o plab-app .                               # 빌드
go vet ./...                                          # 정적 분석
./plab-app doctor                                     # 환경 점검
./plab-app create --name test                         # E2E: 기본
./plab-app create --name test --plab-data             # E2E: 플랩 연동
./plab-app create --name test --researchers-only      # E2E: 리서처 전용
./plab-app create --name test --plab-data --researchers-only  # E2E: 전체 옵션
```

모든 E2E 테스트는 생성된 프로젝트에서 `npm run build`가 통과해야 함.

## 템플릿 수정 시 주의사항

- `.tmpl` 파일만 Go template 렌더링됨
- delimiters는 `<%` `%>` (JSX의 `{}` 충돌 방지)
- 조건부 파일은 `generator.go`의 `shouldSkipDir`/`shouldSkipFile`에서 관리
- `next-auth` 의존성은 `injectAuthDependency()`에서 package.json에 후처리 주입
- 템플릿 변경 후 반드시 4가지 조합 모두 `go build` + `npm run build` 검증

## 에러 코드 (--json 모드)

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
