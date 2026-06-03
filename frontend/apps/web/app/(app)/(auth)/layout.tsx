export const dynamic = 'force-dynamic'

import { extractTenantFromHostname } from '@workspace/shared-http'
import { headers } from 'next/headers'
import { AuthProvider } from '@/components/providers/auth-provider'
import { redirectAuthenticatedUser } from '@/features/auth/application/redirect-authenticated.server'

export default async function AuthLayout({ children }: { children: React.ReactNode }) {
  const incomingHeaders = await headers()
  // Prefer x-forwarded-host (set by Railway/proxies to the public domain) over
  // the raw host header (which may be the internal service address, e.g. localhost:8080).
  const host = incomingHeaders.get('x-forwarded-host') || incomingHeaders.get('host') || ''
  const tenant = extractTenantFromHostname(host)

  const protocol = (() => {
    const p = incomingHeaders.get('x-forwarded-proto')
    return p ? `${p}:` : 'http:'
  })()

  await redirectAuthenticatedUser({ host, protocol, tenant })

  return <AuthProvider>{children}</AuthProvider>
}
