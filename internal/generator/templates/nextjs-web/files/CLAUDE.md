# CLAUDE.md

이 파일은 AI 코드 에이전트가 프로젝트 규칙을 이해하고 따르도록 돕습니다.

> **Skills**: React Best Practices, Web Design Guidelines가 `.skills/` 디렉토리에 설치되어 있습니다.

## 프로젝트 개요

Next.js 16 + React 19 기반 보일러플레이트 템플릿입니다.

## 커밋 규칙

### 커밋 메시지 형식

```
<type>: <description>

<optional body>
```

### Type 종류

| Type | 설명 |
|------|------|
| `feat` | 새로운 기능 추가 |
| `fix` | 버그 수정 |
| `refactor` | 리팩토링 (기능 변경 없음) |
| `docs` | 문서 수정 |
| `test` | 테스트 추가/수정 |
| `chore` | 빌드, 설정 파일 수정 |
| `perf` | 성능 개선 |
| `style` | 코드 포맷팅 (기능 변경 없음) |

### 커밋 메시지 예시

```bash
# Good
feat: add user authentication with NextAuth
fix: resolve hydration mismatch in ThemeProvider
refactor: extract common hooks to shared module

# Bad
update code
fix bug
WIP
```

### 커밋 규칙

- 한 커밋에 하나의 논리적 변경만 포함
- 제목은 50자 이내, 명령형으로 작성
- 본문은 72자에서 줄바꿈
- "왜" 변경했는지 설명 (무엇을 변경했는지는 코드로 확인 가능)

---

## React 성능 최적화 (CRITICAL)

> 출처: [Vercel Engineering React Best Practices](https://github.com/vercel-labs/agent-skills)

### 1. Waterfall 제거 (2-10배 성능 향상)

```typescript
// ❌ Bad: 순차 실행 (3번의 네트워크 왕복)
const user = await fetchUser()
const posts = await fetchPosts()
const comments = await fetchComments()

// ✅ Good: 병렬 실행 (1번의 네트워크 왕복)
const [user, posts, comments] = await Promise.all([
  fetchUser(),
  fetchPosts(),
  fetchComments()
])
```

### 2. await 지연 (필요할 때만 await)

```typescript
// ❌ Bad: 항상 await
async function handleRequest(userId: string, skip: boolean) {
  const userData = await fetchUserData(userId)
  if (skip) return { skipped: true }  // userData 불필요했음
  return processUserData(userData)
}

// ✅ Good: 필요할 때만 await
async function handleRequest(userId: string, skip: boolean) {
  if (skip) return { skipped: true }
  const userData = await fetchUserData(userId)
  return processUserData(userData)
}
```

### 3. Barrel File Import 피하기

```typescript
// ❌ Bad: 전체 라이브러리 로드 (200-800ms 지연)
import { Check, X, Menu } from 'lucide-react'

// ✅ Good: 직접 import
import Check from 'lucide-react/dist/esm/icons/check'
import X from 'lucide-react/dist/esm/icons/x'
import Menu from 'lucide-react/dist/esm/icons/menu'

// ✅ Alternative: next.config.js 설정
// experimental: { optimizePackageImports: ['lucide-react'] }
```

### 4. Strategic Suspense Boundaries

```tsx
// ❌ Bad: 전체 페이지가 데이터 로딩 대기
async function Page() {
  const data = await fetchData()
  return (
    <div>
      <Sidebar />
      <DataDisplay data={data} />
      <Footer />
    </div>
  )
}

// ✅ Good: Sidebar, Footer 즉시 렌더링
function Page() {
  return (
    <div>
      <Sidebar />
      <Suspense fallback={<Skeleton />}>
        <DataDisplay />
      </Suspense>
      <Footer />
    </div>
  )
}
```

### 5. Server Action 인증

```typescript
// ❌ Bad: 미들웨어에만 의존 (Server Action은 공개 엔드포인트!)
"use server"
export async function deleteUser(id: string) {
  await db.user.delete({ where: { id } })
}

// ✅ Good: Server Action 내부에서 명시적 인증
"use server"
export async function deleteUser(id: string) {
  const session = await auth()
  if (!session?.user?.isAdmin) {
    throw new Error("Unauthorized")
  }
  await db.user.delete({ where: { id } })
}
```

### 6. React.cache()로 중복 요청 제거

```typescript
// ✅ 같은 요청에서 여러 컴포넌트가 호출해도 1번만 실행
import { cache } from 'react'

export const getUser = cache(async (id: string) => {
  return await db.user.findUnique({ where: { id } })
})
```

### 7. 비차단 작업은 after() 사용

```typescript
import { after } from 'next/server'

export async function POST(request: Request) {
  const data = await processRequest(request)

  // 응답 후 실행 (응답 지연 없음)
  after(async () => {
    await logAnalytics(data)
    await sendNotification(data)
  })

  return Response.json(data)
}
```

---

## 코딩 스타일

### 불변성 (Immutability) - 필수

```typescript
// ❌ Bad: 객체 변경
function updateUser(user: User, name: string) {
  user.name = name
  return user
}

// ✅ Good: 새 객체 생성
function updateUser(user: User, name: string) {
  return { ...user, name }
}
```

### 파일 구조

- 파일당 200-400줄 권장, 최대 800줄
- 하나의 파일에 하나의 책임
- 관련 파일은 같은 폴더에 배치

```
src/
├── app/                 # 라우트 (App Router)
├── components/
│   ├── ui/              # shadcn/ui 기본 컴포넌트
│   ├── common/          # 공통 컴포넌트
│   └── [feature]/       # 기능별 컴포넌트
├── hooks/               # 커스텀 훅
├── lib/                 # 유틸리티, API 클라이언트
├── providers/           # React Context/Provider
├── stores/              # Zustand 스토어
└── types/               # TypeScript 타입 정의
```

### 네이밍 규칙

| 대상 | 규칙 | 예시 |
|------|------|------|
| 컴포넌트 | PascalCase | `UserProfile.tsx` |
| 훅 | camelCase, use 접두사 | `useDebounce.ts` |
| 유틸리티 | camelCase | `formatDate.ts` |
| 타입/인터페이스 | PascalCase | `UserProfile` |
| 상수 | SCREAMING_SNAKE_CASE | `MAX_RETRY_COUNT` |
| CSS 클래스 | kebab-case (Tailwind) | `text-primary` |

### TypeScript

```typescript
// ✅ 타입 추론 활용
const [count, setCount] = useState(0)

// ✅ 명시적 타입 필요한 경우
const [user, setUser] = useState<User | null>(null)

// ✅ 인터페이스로 Props 정의
interface ButtonProps {
  variant?: "default" | "destructive" | "outline"
  size?: "sm" | "md" | "lg"
  children: React.ReactNode
}

// ❌ any 사용 금지
function process(data: any) { }

// ✅ unknown 또는 적절한 타입 사용
function process(data: unknown) { }
```

### 에러 처리

```typescript
// ✅ try-catch로 에러 처리
try {
  const result = await fetchData()
  return result
} catch (error) {
  console.error("Failed to fetch:", error)
  throw new Error("데이터를 불러올 수 없습니다")
}
```

### 입력 검증

```typescript
// ✅ zod로 입력 검증
import { z } from "zod"

const userSchema = z.object({
  email: z.string().email(),
  age: z.number().int().min(0).max(150),
})

const validated = userSchema.parse(input)
```

---

## 컴포넌트 작성 규칙

### 클라이언트 컴포넌트

```typescript
"use client"

import { useState } from "react"

export function Counter() {
  const [count, setCount] = useState(0)
  return <button onClick={() => setCount(count + 1)}>{count}</button>
}
```

### 서버 컴포넌트 (기본)

```typescript
// "use client" 없으면 서버 컴포넌트
export async function UserList() {
  const users = await fetchUsers()
  return <ul>{users.map(u => <li key={u.id}>{u.name}</li>)}</ul>
}
```

### shadcn/ui 컴포넌트 추가

```bash
npx shadcn@latest add button
npx shadcn@latest add card
```

---

## 상태 관리

### Zustand (클라이언트 상태)

```typescript
// src/stores/useCounterStore.ts
import { create } from "zustand"

interface CounterState {
  count: number
  increment: () => void
  decrement: () => void
}

export const useCounterStore = create<CounterState>((set) => ({
  count: 0,
  increment: () => set((state) => ({ count: state.count + 1 })),
  decrement: () => set((state) => ({ count: state.count - 1 })),
}))
```

### React Query (서버 상태)

```typescript
// src/hooks/useUsers.ts
import { useQuery } from "@tanstack/react-query"
import { api } from "@/lib/api"

export function useUsers() {
  return useQuery({
    queryKey: ["users"],
    queryFn: () => api.get<User[]>("/api/users"),
  })
}
```

---

## 테스트

### 테스트 파일 위치

- `tests/` 폴더에 테스트 파일 배치
- 파일명: `*.test.ts` 또는 `*.test.tsx`

### 테스트 작성

```typescript
import { render, screen } from "@testing-library/react"
import { describe, it, expect } from "vitest"
import { Button } from "@/components/ui/button"

describe("Button", () => {
  it("renders correctly", () => {
    render(<Button>Click me</Button>)
    expect(screen.getByRole("button")).toHaveTextContent("Click me")
  })
})
```

### 테스트 실행

```bash
npm run test           # 테스트 실행
npm run test:coverage  # 커버리지 포함
```

---

## 금지 사항

- ❌ `console.log` 프로덕션 코드에 남기지 않기
- ❌ `any` 타입 사용 금지
- ❌ 하드코딩된 시크릿/API 키
- ❌ 주석 처리된 코드 커밋
- ❌ 사용하지 않는 import/변수
- ❌ Barrel file에서 import (lucide-react, @mui/material 등)
- ❌ 순차적 await (병렬 가능할 때)

## 권장 사항

- ✅ Early return 패턴 사용
- ✅ 함수는 50줄 이내로 유지
- ✅ 의미 있는 변수/함수 이름 사용
- ✅ 복잡한 로직에는 주석 추가
- ✅ 재사용 가능한 컴포넌트 추출
- ✅ `Promise.all()` for 독립적인 비동기 작업
- ✅ `React.cache()` for 중복 요청 제거
- ✅ Strategic Suspense boundaries

---

## 스크립트

```bash
npm run setup          # 프로젝트 초기 설정
npm run dev            # 개발 서버
npm run build          # 프로덕션 빌드
npm run test           # 테스트
npm run lint           # 린트
npm run format         # 코드 포맷팅
```

---

## 설치된 Skills

| Skill | 경로 | 설명 |
|-------|------|------|
| React Best Practices | `.skills/react-best-practices/` | Vercel Engineering 40+ 최적화 규칙 |
| Web Design Guidelines | `.skills/web-design-guidelines/` | UI/UX, 접근성, 성능 100+ 규칙 |

> Skills는 [Agent Skills](https://agentskills.io/) 형식을 따릅니다.
