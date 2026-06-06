import { type NextRequest, NextResponse } from 'next/server'
import { seedAll, seedCMSIfEmpty, validateRefs } from '@/features/cms/seed'
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

  // Guard #1: force re-seed (overwrite:true) is a reset-to-defaults — it
  // rebuilds page block-refs and overwrites navbar/footer/theme globals from
  // the seed constants, clobbering any CMS edits. Disabled in production unless
  // ALLOW_PROD_RESEED is explicitly set (and a DB backup was taken first).
  const isProduction = process.env.NODE_ENV === 'production'
  if (force && isProduction && process.env.ALLOW_PROD_RESEED !== 'true') {
    return NextResponse.json(
      {
        error:
          'Force re-seed is disabled in production. It resets navbar/footer/theme and rebuilds page block-refs from seed defaults, which clobbers CMS edits. Edit content in /cms-admin instead. To override deliberately, back up the cms schema and set ALLOW_PROD_RESEED=true.',
      },
      {
        status: 403,
      }
    )
  }

  try {
    const payload = await getPayloadClient()

    if (force) {
      await seedAll(payload, {
        overwrite: true,
      })
    } else {
      const result = await seedCMSIfEmpty(payload)
      if (!result.seeded) {
        return NextResponse.json({
          message: 'Already seeded, skipping. Pass ?force=1 to overwrite.',
        })
      }
    }

    // Guard #4: after any seed, verify every page block-ref still resolves to a
    // block-library entry. Broken refs render as an empty page body (the exact
    // failure that silently wiped the landing), so surface them loudly.
    const broken = await validateRefs(payload)
    if (broken.length > 0) {
      console.error('[seed-cms] broken block-refs detected after seed:', broken)
      return NextResponse.json(
        {
          message:
            'Seeded, but broken block-refs were detected — affected pages will render empty. Re-link the blocks in /cms-admin.',
          brokenRefs: broken,
        },
        {
          status: 207,
        }
      )
    }

    return NextResponse.json({
      message: force ? 'Force re-seeded landing/navbar/footer.' : 'Seeded successfully.',
    })
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
