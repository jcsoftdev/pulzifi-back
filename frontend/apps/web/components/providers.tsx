'use client'

import type * as React from 'react'
import { ThemeProvider as NextThemesProvider } from 'next-themes'
import { SessionProvider } from 'next-auth/react'

export function Providers({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <SessionProvider refetchInterval={0} refetchOnWindowFocus={false}>
      <NextThemesProvider
        attribute="class"
        defaultTheme="system"
        enableSystem
        disableTransitionOnChange
        enableColorScheme
      >
        {children}
      </NextThemesProvider>
    </SessionProvider>
  )
}
