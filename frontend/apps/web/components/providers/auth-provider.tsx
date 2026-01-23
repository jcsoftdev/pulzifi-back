'use client'

import { ValidatedSessionProvider } from '@/components/providers/validated-session-provider'
import { ClientInit } from '@/components/client-init'
import type { ReactNode } from 'react'

export function AuthProvider({
  children,
}: Readonly<{
  children: ReactNode
}>) {
  return (
    <ValidatedSessionProvider refetchInterval={0} refetchOnWindowFocus={false}>
      <ClientInit />
      {children}
    </ValidatedSessionProvider>
  )
}
