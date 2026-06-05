import { type NextRequest, NextResponse } from 'next/server'
import { seedAll, seedCMSIfEmpty } from '@/features/cms/seed'
import { getPayloadClient } from '@/lib/payload'

export async function POST(req: NextRequest) {
  const secret = req.headers.get('x-seed-secret')
  if (!process.env.SEED_SECRET || secret !== process.env.SEED_SECRET) {
    return NextResponse.json(
      {
        error: 'Unauthorized',
      },
      {
        status: 401,
      }
    )
  }

  const url = new URL(req.url)
  const force = url.searchParams.get('force') === '1'

  try {
    const payload = await getPayloadClient()

    if (force) {
      await seedAll(payload, {
        overwrite: true,
      })
      return NextResponse.json({
        message: 'Force re-seeded landing/navbar/footer.',
      })
    }

    const result = await seedCMSIfEmpty(payload)
    return NextResponse.json(
      result.seeded
        ? {
            message: 'Seeded successfully.',
          }
        : {
            message: 'Already seeded, skipping. Pass ?force=1 to overwrite.',
          }
    )
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err)
    const stack = err instanceof Error ? err.stack : undefined
    console.error('[seed-cms] failed:', err)
    return NextResponse.json(
      {
        error: message,
        stack,
      },
      {
        status: 500,
      }
    )
  }
}
