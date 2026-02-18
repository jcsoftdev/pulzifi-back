'use client'

import type { ReactNode } from 'react'

export function AuthProvider({
  children,
}: Readonly<{
  children: ReactNode
}>) {
  return <>{children}</>
}
