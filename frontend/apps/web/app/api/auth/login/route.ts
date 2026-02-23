import { saveNonce } from '@/lib/auth-nonce-store'
import { authCookieOptions } from '@/lib/cookie-options'
import { getBackendOrigin } from '@/lib/server-config'
import type { NextRequest } from 'next/server'
import { NextResponse } from 'next/server'
import { randomUUID } from 'crypto'

export async function POST(request: NextRequest) {
  try {
    const body = await request.json()

    const response = await fetch(`${getBackendOrigin()}/api/v1/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
      cache: 'no-store',
    })

    const data = await response.json()

    if (!response.ok) {
      return NextResponse.json(data, { status: response.status })
    }

    const { access_token, refresh_token, expires_in, tenant } = data

    console.log('[Login] Go backend response keys:', Object.keys(data))
    console.log('[Login] Has tokens:', { access_token: !!access_token, refresh_token: !!refresh_token, tenant })

    if (!access_token || !refresh_token) {
      console.error('[Login] Go backend did not return tokens in response body. Restart the Go server.')
      console.error('[Login] Full response data:', JSON.stringify(data))
      return NextResponse.json(
        { error: 'Server configuration error — tokens missing from backend response' },
        { status: 500 }
      )
    }

    // Store tokens under a one-time nonce for cross-subdomain redirect.
    // E.g. login at localhost:3000 → redirect to tenant.localhost:3000/api/auth/callback
    const nonce = randomUUID()
    saveNonce(nonce, {
      accessToken: access_token,
      refreshToken: refresh_token,
      expiresIn: expires_in,
    })

    const nextResponse = NextResponse.json(
      { expires_in, tenant, nonce },
      { status: 200 }
    )

    const { isSecure, cookieDomain, sameSite } = authCookieOptions(request)

    nextResponse.cookies.set('access_token', access_token, {
      path: '/',
      httpOnly: true,
      secure: isSecure,
      sameSite,
      maxAge: expires_in,
      ...(cookieDomain ? { domain: cookieDomain } : {}),
    })

    nextResponse.cookies.set('refresh_token', refresh_token, {
      path: '/',
      httpOnly: true,
      secure: isSecure,
      sameSite,
      maxAge: 7 * 24 * 60 * 60,
      ...(cookieDomain ? { domain: cookieDomain } : {}),
    })

    return nextResponse
  } catch (error) {
    console.error('[Login] Error:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}
