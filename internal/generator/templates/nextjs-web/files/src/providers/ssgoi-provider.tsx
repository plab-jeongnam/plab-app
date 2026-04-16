"use client"

import { Ssgoi } from "@ssgoi/react"
import { usePathname } from "next/navigation"
import { type ReactNode } from "react"

import { ssgoiConfig } from "./ssgoi-config"

interface SsgoiProviderProps {
  children: ReactNode
}

export function SsgoiProvider({ children }: SsgoiProviderProps) {
  return (
    <Ssgoi config={ssgoiConfig} usePathname={usePathname}>
      {children}
    </Ssgoi>
  )
}
