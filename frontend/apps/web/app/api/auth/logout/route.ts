import { signOut } from '@workspace/auth'
import { NextResponse } from 'next/server'

export async function POST() {
  try {
    await signOut({
      redirect: false,
    })
    return NextResponse.json({
      success: true,
    })
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
