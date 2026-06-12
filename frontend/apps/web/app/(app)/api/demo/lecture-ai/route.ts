import { type NextRequest, NextResponse } from 'next/server'
import { getDemoState, setDemoState } from '@/lib/demo-store'

export function GET() {
  return NextResponse.json(getDemoState())
}

export async function POST(req: NextRequest) {
  const body = await req.json()
  const version = body?.version
  if (version !== '1' && version !== '2') {
    return NextResponse.json({ error: 'version must be "1" or "2"' }, { status: 400 })
  }
  const state = setDemoState({ lectureAiVersion: version })
  return NextResponse.json(state)
}
