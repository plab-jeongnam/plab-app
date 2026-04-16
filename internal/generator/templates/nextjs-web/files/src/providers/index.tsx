"use client"

import type { ReactNode } from "react"

import { QueryProvider } from "./query-provider"
import { SsgoiProvider } from "./ssgoi-provider"

interface ProvidersProps {
  children: ReactNode
}

export function Providers({ children }: ProvidersProps) {
  return (
    <QueryProvider>
      <SsgoiProvider>{children}</SsgoiProvider>
    </QueryProvider>
  )
}
