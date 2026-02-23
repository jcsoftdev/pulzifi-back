import { getBackendOrigin } from '@/lib/server-config'
import type { NextRequest } from 'next/server'
import { NextResponse } from 'next/server'

export async function POST(request: NextRequest) {
  try {
    const body = await request.json()
    const response = await fetch(`${getBackendOrigin()}/api/v1/auth/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
      cache: 'no-store',
    })
    const data = await response.json()
    return NextResponse.json(data, { status: response.status })
  } catch (error) {
    console.error('[Register] Error:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}
