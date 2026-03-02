import { UnauthorizedError } from '@workspace/shared-http'
import { redirect } from 'next/navigation'

export function handleServerAuthError(error: unknown): never {
  if (error instanceof UnauthorizedError) {
    redirect('/login')
  }

  throw error
}