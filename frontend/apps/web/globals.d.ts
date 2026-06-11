declare module '*.css' {
  const classes: {
    [key: string]: string
  }
  export default classes
}

declare module '*.svg' {
  const content: string
  export default content
}

type FbqStandardEvent =
  | 'PageView'
  | 'CompleteRegistration'
  | 'Lead'
  | 'Purchase'
  | 'InitiateCheckout'

interface Window {
  fbq?: (
    method: 'init' | 'track' | 'trackCustom',
    event: FbqStandardEvent | string,
    params?: Record<string, unknown>
  ) => void
  _fbq?: unknown
}
