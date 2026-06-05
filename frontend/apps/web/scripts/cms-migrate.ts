/**
 * Run Payload CMS migrations programmatically via bun.
 *
 * Usage:
 *   bun scripts/cms-migrate.ts            — apply pending migrations (safe, keeps data)
 *   bun scripts/cms-migrate.ts --create   — generate migration from schema diff
 *   bun scripts/cms-migrate.ts --fresh    — drop schema and re-run all migrations (DESTROYS DATA)
 *
 * SAFETY: any operation that drops the `cms` schema (--fresh, or auto-recovery
 * from a broken/partial schema) is REFUSED unless CMS_ALLOW_DESTRUCTIVE=1 is set.
 * This holds for every DB host — local or prod — so an accidental run can never
 * wipe landing/CMS content. First-time init (schema absent) is non-destructive
 * and runs normally.
 */

import path from 'node:path'
import { fileURLToPath } from 'node:url'
import pg from 'pg'
import { seedCMSIfEmpty } from '../features/cms/seed'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

const payloadBin = path.resolve(__dirname, '../../../node_modules/payload/dist/bin/migrate.js')
const { migrate } = await import(payloadBin)

const connStr = `postgresql://${process.env.DB_USER}:${process.env.DB_PASSWORD}@${process.env.DB_HOST}:${process.env.DB_PORT}/${process.env.DB_NAME}`

// Prepare cms schema. Any DROP (--fresh, or recovering a broken partial schema)
// requires CMS_ALLOW_DESTRUCTIVE=1 and is refused otherwise (see SAFETY note above).
const isFreshEarly = process.argv.includes('--fresh')
const allowDestructive = process.env.CMS_ALLOW_DESTRUCTIVE === '1'
try {
  const client = new pg.Client({
    connectionString: connStr,
  })
  await client.connect()

  // Does the cms schema exist at all? Absent = first-time init (non-destructive).
  const { rows: schemaRows } = await client.query<{
    exists: boolean
  }>(`
    SELECT EXISTS (
      SELECT 1 FROM information_schema.schemata WHERE schema_name = 'cms'
    ) AS exists
  `)
  const schemaExists = schemaRows[0]?.exists ?? false

  // Check for a key data table (not payload_migrations, which is created before the migration
  // transaction starts and may exist even after a failed/partial migration run).
  const { rows } = await client.query<{
    exists: boolean
  }>(`
    SELECT EXISTS (
      SELECT 1 FROM information_schema.tables
      WHERE table_schema = 'cms' AND table_name = 'pages'
    ) AS exists
  `)
  const pagesExists = rows[0]?.exists ?? false

  // A drop is destructive only when the schema actually exists. First init (schema
  // absent) just needs CREATE — no drop, no gate.
  //   --fresh: explicit destructive intent → always gated.
  //   broken state (schema present but cms.pages missing) → gated too.
  const wantsDrop = schemaExists && (isFreshEarly || !pagesExists)

  if (wantsDrop && !allowDestructive) {
    const reason = isFreshEarly ? '--fresh flag' : 'broken/partial cms schema (cms.pages missing)'
    console.error(
      [
        '',
        '[cms-migrate] REFUSING destructive operation.',
        `  Trigger:   ${reason}`,
        '  Would run: DROP SCHEMA cms CASCADE  → destroys ALL CMS / landing content',
        `  Target DB: ${process.env.DB_HOST}:${process.env.DB_PORT}/${process.env.DB_NAME}`,
        '  To proceed intentionally, re-run with CMS_ALLOW_DESTRUCTIVE=1',
        '',
      ].join('\n')
    )
    await client.end()
    process.exit(1)
  }

  if (wantsDrop) {
    console.warn(
      `[cms-migrate] DROP SCHEMA cms CASCADE on ${process.env.DB_HOST}/${process.env.DB_NAME} (CMS_ALLOW_DESTRUCTIVE=1)`
    )
    await client.query(`DROP SCHEMA IF EXISTS cms CASCADE`)
  }
  await client.query(`CREATE SCHEMA IF NOT EXISTS cms`)
  await client.end()
} catch (err) {
  console.error('[cms-migrate] failed to prepare cms schema:', err)
  process.exit(1)
}

const { default: rawConfig } = await import('../payload.config')
const builtConfig = await rawConfig

const isFresh = process.argv.includes('--fresh')
const isCreate = process.argv.includes('--create')

if (!isFresh && !isCreate) {
  // Delete the dev-mode marker row so migrate doesn't prompt.
  try {
    const client = new pg.Client({
      connectionString: connStr,
    })
    await client.connect()
    await client.query(`DELETE FROM cms.payload_migrations WHERE batch = -1`)
    await client.end()
  } catch {
    // Table may not exist yet on first run — that's fine
  }
}

const parsedArgs = {
  _: [
    isCreate ? 'migrate:create' : isFresh ? 'migrate:fresh' : 'migrate',
  ],
  forceAcceptWarning: true,
}

// Run migrations first — tables must exist before Payload initializes (onInit fires on getPayload)
await migrate({
  config: builtConfig,
  parsedArgs,
})

// Initialize Payload after migrations so onInit sees the complete schema
const { getPayload } = await import('payload')
const payload = await getPayload({
  config: builtConfig,
})

// Seed default content after migrations
const result = await seedCMSIfEmpty(payload)
if (result.seeded) {
  console.log('[cms] seeded default content (block-library, pages, globals)')
} else {
  console.log('[cms] seed skipped:', result.reason)
}

await payload.db.destroy?.()
process.exit(0)
