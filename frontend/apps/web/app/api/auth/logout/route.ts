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

/**
 * GET /api/auth/logout?redirectTo=/login
 *
 * Used as a cross-subdomain cleanup step: the tenant subdomain redirects the
 * browser here so the main-domain cookies (set during login) are also cleared.
 */
export async function GET(request: NextRequest) {
  const redirectTo = request.nextUrl.searchParams.get('redirectTo') || '/login'
  const response = NextResponse.redirect(new URL(redirectTo, request.url))
  response.cookies.set('access_token', '', { path: '/', httpOnly: true, maxAge: 0 })
  response.cookies.set('refresh_token', '', { path: '/', httpOnly: true, maxAge: 0 })
  response.cookies.set('tenant_hint', '', { path: '/', httpOnly: true, maxAge: 0 })

  return response
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

    nextResponse.cookies.set('access_token', '', { path: '/', httpOnly: true, maxAge: 0 })
    nextResponse.cookies.set('refresh_token', '', { path: '/', httpOnly: true, maxAge: 0 })
    nextResponse.cookies.set('tenant_hint', '', { path: '/', httpOnly: true, maxAge: 0 })

    return nextResponse
  } catch (error) {
    console.error('[Logout] Error:', error)
    return NextResponse.json({ success: false }, { status: 500 })
  }
}
