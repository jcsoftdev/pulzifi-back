'use client'

import { createContext, type ReactNode, useContext } from 'react'

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
