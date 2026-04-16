"use client"

import { useSession } from "next-auth/react"
import { useRouter } from "next/navigation"
import { useEffect } from "react"

interface AuthGuardProps {
  children: React.ReactNode
}

export function AuthGuard({ children }: AuthGuardProps) {
  const { status } = useSession()
  const router = useRouter()

  useEffect(() => {
    if (status === "unauthenticated") {
      router.push("/login")
    }
  }, [status, router])

  if (status === "loading") {
    return (
      <div className="grid min-h-screen place-items-center">
        <p className="text-muted-foreground">로딩 중...</p>
      </div>
    )
  }

  if (status === "unauthenticated") {
    return null
  }

  return <>{children}</>
}
