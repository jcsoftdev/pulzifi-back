import type { NextRequest } from 'next/server'
import { NextResponse } from 'next/server'
import { env } from '@/lib/env'

function getBackendOrigin(): string {
  const apiBase = env.SERVER_API_URL ?? 'http://localhost:9090'
  try {
    return new URL(apiBase).origin
  } catch {
    return 'http://localhost:9090'
  }
}

export async function POST(request: NextRequest) {
  try {
    const cookie = request.headers.get('cookie') || ''

    await fetch(`${getBackendOrigin()}/api/v1/auth/logout`, {
      method: 'POST',
      headers: { Cookie: cookie },
      cache: 'no-store',
    })

    const nextResponse = NextResponse.json({ success: true })

    // Clear cookies without Domain so they match the current origin
    nextResponse.cookies.set('access_token', '', {
      path: '/',
      httpOnly: true,
      maxAge: 0,
    })
    nextResponse.cookies.set('refresh_token', '', {
      path: '/',
      httpOnly: true,
      maxAge: 0,
    })

    return nextResponse
  } catch (error) {
    console.error('[Logout] Error:', error)
    return NextResponse.json({ success: false }, { status: 500 })
  }
}
