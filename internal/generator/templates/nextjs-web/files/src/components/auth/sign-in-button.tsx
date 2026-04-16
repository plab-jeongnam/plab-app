"use client"

import { signIn, signOut, useSession } from "next-auth/react"

import { Button } from "@/components/ui/button"

export function SignInButton() {
  const { data: session, status } = useSession()

  if (status === "loading") {
    return <Button variant="outline" disabled>로딩 중...</Button>
  }

  if (session) {
    return (
      <div className="flex items-center gap-3">
        <span className="text-sm text-muted-foreground">
          {session.user?.name ?? session.user?.email}
        </span>
        <Button variant="outline" size="sm" onClick={() => signOut()}>
          로그아웃
        </Button>
      </div>
    )
  }

  return (
    <Button onClick={() => signIn("google")}>
      Google로 로그인
    </Button>
  )
}
