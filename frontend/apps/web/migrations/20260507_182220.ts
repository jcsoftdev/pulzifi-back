import { type MigrateUpArgs, type MigrateDownArgs, sql } from '@payloadcms/db-postgres'

export async function up({ db }: MigrateUpArgs): Promise<void> {
  // Change ai-intelligence tab item image from upload (image_id FK) to text URL
  await db.execute(sql`
    ALTER TABLE cms.block_library_blocks_ai_intelligence_tabs_items
      DROP COLUMN IF EXISTS image_id,
      ADD COLUMN IF NOT EXISTS image varchar;
  `)
  await db.execute(sql`
    ALTER TABLE cms.pages_blocks_ai_intelligence_tabs_items
      DROP COLUMN IF EXISTS image_id,
      ADD COLUMN IF NOT EXISTS image varchar;
  `)
  await db.execute(sql`
    ALTER TABLE cms._pages_v_blocks_ai_intelligence_tabs_items
      DROP COLUMN IF EXISTS image_id,
      ADD COLUMN IF NOT EXISTS image varchar;
  `)
}

export async function down({ db }: MigrateDownArgs): Promise<void> {
  await db.execute(sql`
    ALTER TABLE cms.block_library_blocks_ai_intelligence_tabs_items
      DROP COLUMN IF EXISTS image,
      ADD COLUMN IF NOT EXISTS image_id integer;
  `)
  await db.execute(sql`
    ALTER TABLE cms.pages_blocks_ai_intelligence_tabs_items
      DROP COLUMN IF EXISTS image,
      ADD COLUMN IF NOT EXISTS image_id integer;
  `)
  await db.execute(sql`
    ALTER TABLE cms._pages_v_blocks_ai_intelligence_tabs_items
      DROP COLUMN IF EXISTS image,
      ADD COLUMN IF NOT EXISTS image_id integer;
  `)
}
