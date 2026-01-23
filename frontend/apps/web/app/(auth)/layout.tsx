import { redirect } from 'next/navigation'
import { auth } from '@workspace/auth'
import { AuthProvider } from '@/components/providers/auth-provider'

export default async function AuthLayout({ children }: { children: React.ReactNode }) {
  const session = await auth()

  // If user is already authenticated, redirect to home
  if (session) {
    redirect('/')
  }

  return (
    <AuthProvider>
      {children}
    </AuthProvider>
  )
}
