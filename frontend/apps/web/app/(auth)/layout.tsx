import { redirect } from 'next/navigation'
import { headers } from 'next/headers'
import { AuthProvider } from '@/components/providers/auth-provider'
import { AuthApi } from '@workspace/services'
import { UnauthorizedError, extractTenantFromHostname } from '@workspace/shared-http'

export default async function AuthLayout({ children }: { children: React.ReactNode }) {
  const incomingHeaders = await headers()
  const host = incomingHeaders.get('host') || ''
  const tenant = extractTenantFromHostname(host)

  if (!tenant) {
    return (
      <AuthProvider>
        {children}
      </AuthProvider>
    )
  }

  try {
    await AuthApi.getCurrentUser()
    redirect('/')
  } catch (error: unknown) {
    if (error instanceof UnauthorizedError) {
      // unauthenticated: allow auth pages
    }
  }

  return (
    <AuthProvider>
      {children}
    </AuthProvider>
  )
}
