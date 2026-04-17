"use client"

import Image from "next/image"
import { signOut, useSession } from "next-auth/react"

import { Button } from "@/components/ui/button"

export function NavBar() {
  const { data: session } = useSession()

  if (!session) return null

  const user = session.user
  const displayName = user?.name ?? user?.email ?? "사용자"
  const initial = displayName.slice(0, 1).toUpperCase()

  return (
    <header className="border-b">
      <div className="container mx-auto flex items-center justify-between px-4 py-3">
        <div className="flex items-center gap-3">
          {user?.image ? (
            <Image
              src={user.image}
              alt={displayName}
              width={32}
              height={32}
              className="rounded-full"
            />
          ) : (
            <div className="flex h-8 w-8 items-center justify-center rounded-full bg-muted text-sm font-medium text-muted-foreground">
              {initial}
            </div>
          )}
          <span className="text-sm font-semibold">{displayName}</span>
        </div>
        <Button variant="ghost" size="sm" onClick={() => signOut({ callbackUrl: "/login" })}>
          로그아웃
        </Button>
      </div>
    </header>
  )
}
