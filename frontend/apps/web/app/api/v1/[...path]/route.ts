import { getBackendOrigin } from '@/lib/server-config'
import type { NextRequest } from 'next/server'
import { NextResponse } from 'next/server'

/**
 * Catch-all proxy for /api/v1/* → Go backend.
 *
 * Replaces the build-time next.config.mjs rewrite so the backend URL is
 * resolved at request time from SERVER_API_URL. This means Railway runtime
 * env vars work without needing the URL baked into the build.
 */
async function forward(
  request: NextRequest,
  context: { params: Promise<{ path: string[] }> }
): Promise<Response> {
  const { path } = await context.params
  const backend = getBackendOrigin()
  const search = request.nextUrl.search
  const url = `${backend}/api/v1/${path.join('/')}${search}`

  const requestHeaders = new Headers(request.headers)
  // Don't forward the host — the backend resolves its own host
  requestHeaders.delete('host')

  const hasBody = request.method !== 'GET' && request.method !== 'HEAD'

  let backendResponse: Response
  try {
    backendResponse = await fetch(url, {
      method: request.method,
      headers: requestHeaders,
      body: hasBody ? request.body : undefined,
      // @ts-expect-error — duplex required for streaming request bodies
      duplex: 'half',
      cache: 'no-store',
    })
  } catch (err) {
    console.error(`[api/v1] Failed to proxy ${request.method} ${url}`, err)
    return NextResponse.json({ error: 'Backend unavailable' }, { status: 502 })
  }

  return new Response(backendResponse.body, {
    status: backendResponse.status,
    headers: backendResponse.headers,
  })
}

export const GET = forward
export const POST = forward
export const PUT = forward
export const PATCH = forward
export const DELETE = forward
export const OPTIONS = forward
