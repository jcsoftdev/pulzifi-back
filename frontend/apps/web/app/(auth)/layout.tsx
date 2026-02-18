import { AuthApi } from '@workspace/services'
import { extractTenantFromHostname, UnauthorizedError } from '@workspace/shared-http'
import { headers } from 'next/headers'
import { redirect } from 'next/navigation'
import { AuthProvider } from '@/components/providers/auth-provider'

export default async function AuthLayout({ children }: { children: React.ReactNode }) {
  const incomingHeaders = await headers()
  const host = incomingHeaders.get('host') || ''
  const tenant = extractTenantFromHostname(host)

  if (!tenant) {
    return <AuthProvider>{children}</AuthProvider>
  }

  try {
    await AuthApi.getCurrentUser()
    redirect('/')
  } catch (error: unknown) {
    if (error instanceof UnauthorizedError) {
      // unauthenticated: allow auth pages
    }
  }

  return <AuthProvider>{children}</AuthProvider>
}
