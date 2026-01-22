'use client'

import type * as React from 'react'
import { ThemeProvider as NextThemesProvider } from 'next-themes'
import { ValidatedSessionProvider } from '@/components/providers/validated-session-provider'

export function Providers({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <ValidatedSessionProvider refetchInterval={0} refetchOnWindowFocus={false}>
      <NextThemesProvider
        attribute="class"
        defaultTheme="system"
        enableSystem
        disableTransitionOnChange
        enableColorScheme
      >
        {children}
      </NextThemesProvider>
    </ValidatedSessionProvider>
  )
}
