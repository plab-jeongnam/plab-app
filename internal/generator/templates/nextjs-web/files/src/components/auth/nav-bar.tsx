"use client"

import { signOut, useSession } from "next-auth/react"

import { Button } from "@/components/ui/button"

export function NavBar() {
  const { data: session } = useSession()

  if (!session) return null

  return (
    <header className="border-b">
      <div className="container mx-auto flex items-center justify-between px-4 py-3">
        <span className="font-semibold text-sm">
          {session.user?.name ?? session.user?.email}
        </span>
        <Button variant="ghost" size="sm" onClick={() => signOut({ callbackUrl: "/login" })}>
          로그아웃
        </Button>
      </div>
    </header>
  )
}
