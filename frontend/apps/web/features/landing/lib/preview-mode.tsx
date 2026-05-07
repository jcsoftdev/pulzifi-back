'use client'

import { createContext, useContext, type ReactNode } from 'react'

const PreviewModeContext = createContext<boolean>(false)

export function PreviewModeProvider({
  children,
  value = true,
}: {
  children: ReactNode
  value?: boolean
}) {
  return <PreviewModeContext.Provider value={value}>{children}</PreviewModeContext.Provider>
}

export function usePreviewMode(): boolean {
  return useContext(PreviewModeContext)
}
