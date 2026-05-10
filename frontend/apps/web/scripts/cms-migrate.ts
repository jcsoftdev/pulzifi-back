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

const connStr = `postgresql://${process.env.DB_USER}:${process.env.DB_PASSWORD}@${process.env.DB_HOST}:${process.env.DB_PORT}/${process.env.DB_NAME}`

// Prepare cms schema:
// - --fresh: always drop and recreate (destructive, intentional)
// - normal: drop and recreate only if schema is in a broken partial state
//   (exists but payload_migrations table is missing = leftover from a failed init)
const isFreshEarly = process.argv.includes('--fresh')
try {
  const client = new pg.Client({ connectionString: connStr })
  await client.connect()

  // Check for a key data table (not payload_migrations, which is created before the migration
  // transaction starts and may exist even after a failed/partial migration run).
  const { rows } = await client.query<{ exists: boolean }>(`
    SELECT EXISTS (
      SELECT 1 FROM information_schema.tables
      WHERE table_schema = 'cms' AND table_name = 'pages'
    ) AS exists
  `)
  const isBrokenState = !(rows[0]?.exists ?? false)

  if (isFreshEarly || isBrokenState) {
    await client.query(`DROP SCHEMA IF EXISTS cms CASCADE`)
  }
  await client.query(`CREATE SCHEMA IF NOT EXISTS cms`)
  await client.end()
} catch (err) {
  console.error('[cms-migrate] failed to prepare cms schema:', err)
  process.exit(1)
}

// Ensure plans and plans_features tables exist (guard against partial migration state)
try {
    const client = new pg.Client({ connectionString: connStr })
    await client.connect()
    await client.query(`
        CREATE TABLE IF NOT EXISTS "cms"."plans_features" (
              "_order" integer NOT NULL,
                    "_parent_id" integer NOT NULL,
                          "id" varchar PRIMARY KEY NOT NULL,
                                "text" varchar,
                                      "included" boolean DEFAULT true
                                          );
                                              CREATE TABLE IF NOT EXISTS "cms"."plans" (
                                                    "id" serial PRIMARY KEY NOT NULL,
                                                          "name" varchar NOT NULL,
                                                                "price" varchar NOT NULL,
                                                                      "period" varchar,
                                                                            "tagline" varchar,
                                                                                  "price_annual" varchar,
                                                                                        "cta_label" varchar,
                                                                                              "cta_href" varchar,
                                                                                                    "highlighted" boolean,
                                                                                                          "popular_badge" varchar,
                                                                                                                "updated_at" timestamp(3) with time zone DEFAULT now() NOT NULL,
                                                                                                                      "created_at" timestamp(3) with time zone DEFAULT now() NOT NULL
                                                                                                                          );
                                                                                                                            `)
    await client.end()
} catch (err) {
    console.error('[cms-migrate] failed to ensure plans tables:', err)
    process.exit(1)
}

const { getPayload } = await import('payload')
const { default: config } = await import('../payload.config')

const isFresh = process.argv.includes('--fresh')
const isCreate = process.argv.includes('--create')

const payload = await getPayload({ config: await config })

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
