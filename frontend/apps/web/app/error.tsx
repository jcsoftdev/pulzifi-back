'use client'

import { useRouter } from 'next/navigation'
import { useEffect } from 'react'

export default function ErrorPage({
  error,
  reset,
}: Readonly<{
  error: Error & {
    digest?: string
  }
  reset: () => void
}>) {
  const router = useRouter()

  useEffect(() => {
    // Check if it's an unauthorized error
    if (error.message === 'Unauthorized' || error.name === 'UnauthorizedError') {
      // Call logout API to clear cookies
      fetch('/api/auth/logout', {
        method: 'POST',
      }).finally(() => {
        // Redirect to login
        router.push('/login')
      })
      return
    }

    // Log other errors to console
    console.error(error.message)
  }, [
    error,
    router,
  ])

  // Don't show UI for unauthorized errors, just redirect
  if (error.message === 'Unauthorized' || error.name === 'UnauthorizedError') {
    return null
  }

  return (
    <div>
      <h2>Something went wrong!</h2>
      <button type="button" onClick={() => reset()}>
        Try again
      </button>
    </div>
  )
}
