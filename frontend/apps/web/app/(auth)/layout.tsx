import { redirect } from 'next/navigation'
import { auth } from '@workspace/auth'

export default async function AuthLayout({ children }: { children: React.ReactNode }) {
  const session = await auth()

  // If user is already authenticated, redirect to home
  if (session) {
    redirect('/')
  }

  return children
}
