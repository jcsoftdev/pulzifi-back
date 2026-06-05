// biome-ignore-all lint/correctness/noUnusedFunctionParameters: generated migration file — payload/req required by interface
import { type MigrateDownArgs, type MigrateUpArgs, sql } from '@payloadcms/db-postgres'

export async function up({ db, payload, req }: MigrateUpArgs): Promise<void> {
  await db.execute(sql`
    ALTER TABLE "cms"."posts" ADD COLUMN IF NOT EXISTS "reading_time" numeric;
    ALTER TABLE "cms"."_posts_v" ADD COLUMN IF NOT EXISTS "version_reading_time" numeric;
  `)
}

export async function down({ db, payload, req }: MigrateDownArgs): Promise<void> {
  await db.execute(sql`
    ALTER TABLE "cms"."posts" DROP COLUMN IF EXISTS "reading_time";
    ALTER TABLE "cms"."_posts_v" DROP COLUMN IF EXISTS "version_reading_time";
  `)
}
