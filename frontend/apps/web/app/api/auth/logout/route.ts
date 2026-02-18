import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

function getBackendOrigin(): string {
  const apiBase =
    process.env.SERVER_API_URL ?? process.env.API_URL ?? process.env.NEXT_PUBLIC_API_URL ?? ''

  if (!apiBase) {
    return 'http://localhost:9090'
  }

  try {
    return new URL(apiBase).origin
  } catch {
    return new URL(`http://${apiBase}`).origin
  }
}

export async function POST(request: NextRequest) {
  try {
    const cookie = request.headers.get('cookie') || ''
    const response = await fetch(`${getBackendOrigin()}/api/v1/auth/logout`, {
      method: 'POST',
      headers: {
        Cookie: cookie,
      },
      cache: 'no-store',
    })

    const nextResponse = NextResponse.json({
      success: true,
    })

    const setCookie = response.headers.get('set-cookie')
    if (setCookie) {
      nextResponse.headers.set('set-cookie', setCookie)
    }

    return nextResponse
  } catch (error) {
    console.error('[Logout] Error:', error)
    return NextResponse.json(
      {
        success: false,
      },
      {
        status: 500,
      }
    )
  }
}
