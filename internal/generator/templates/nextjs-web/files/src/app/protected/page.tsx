"use client"

import { useSession } from "next-auth/react"

import { AuthGuard } from "@/components/auth/auth-guard"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"

export default function ProtectedPage() {
  return (
    <AuthGuard>
      <ProtectedContent />
    </AuthGuard>
  )
}

function ProtectedContent() {
  const { data: session } = useSession()

  return (
    <main className="container mx-auto py-10 px-4 max-w-3xl">
      <h1 className="mb-6 text-2xl font-bold">리서처 전용 페이지</h1>
      <Card>
        <CardHeader>
          <CardTitle>로그인 정보</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2 text-sm">
          <p>이름: {session?.user?.name}</p>
          <p>이메일: {session?.user?.email}</p>
          <p className="text-muted-foreground mt-4">
            이 페이지는 로그인한 사용자만 볼 수 있어요.
            AuthGuard 컴포넌트로 감싸면 어떤 페이지든 보호할 수 있어요.
          </p>
        </CardContent>
      </Card>
    </main>
  )
}
