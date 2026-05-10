/**
 * Run Payload CMS migrations programmatically via bun.
 *
 * Usage:
 *   bun scripts/cms-migrate.ts            — apply pending migrations (safe, keeps data)
 *   bun scripts/cms-migrate.ts --create   — generate migration from schema diff
 *   bun scripts/cms-migrate.ts --fresh    — drop schema and re-run all migrations (DESTROYS DATA)
 */
import { fileURLToPath } from 'url'
import path from 'path'
import pg from 'pg'
import { seedCMSIfEmpty } from '../features/cms/seed'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

const payloadBin = path.resolve(__dirname, '../../../node_modules/payload/dist/bin/migrate.js')
const { migrate } = await import(payloadBin)

// Ensure the cms schema exists BEFORE Payload initializes (onInit hook runs during getPayload)
const connStr = `postgresql://${process.env.DB_USER}:${process.env.DB_PASSWORD}@${process.env.DB_HOST}:${process.env.DB_PORT}/${process.env.DB_NAME}`
try {
  const client = new pg.Client({ connectionString: connStr })
  await client.connect()
  await client.query(`CREATE SCHEMA IF NOT EXISTS cms`)
  await client.end()
} catch (err) {
  console.error('[cms-migrate] failed to create cms schema:', err)
  process.exit(1)
}

const { getPayload } = await import('payload')
const { default: config } = await import('../payload.config')

const isFresh = process.argv.includes('--fresh')
const isCreate = process.argv.includes('--create')

if (!isFresh && !isCreate) {
  // Delete the dev-mode marker row so migrate doesn't prompt.
  // This row (batch = -1) is written when Payload runs in push mode.
  // Removing it is safe — it just tells Payload "no pending dev-mode schema drift".
  try {
    const client = new pg.Client({ connectionString: connStr })
    await client.connect()
    await client.query(`DELETE FROM cms.payload_migrations WHERE batch = -1`)
    await client.end()
  } catch {
    // Table may not exist yet on first run — that's fine
  }
}

const parsedArgs = {
  _: [isCreate ? 'migrate:create' : isFresh ? 'migrate:fresh' : 'migrate'],
  forceAcceptWarning: true,
}

await migrate({ config: payload.config, parsedArgs })

// Seed default content after migrations
const result = await seedCMSIfEmpty(payload)
if (result.seeded) {
  console.log('[cms] seeded default content (block-library, pages, globals)')
} else {
  console.log('[cms] seed skipped:', result.reason)
}

await payload.db.destroy?.()
process.exit(0)
