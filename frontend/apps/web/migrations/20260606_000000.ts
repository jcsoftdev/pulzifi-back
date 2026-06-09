// biome-ignore-all lint/correctness/noUnusedFunctionParameters: generated migration file — payload/req required by interface
import { type MigrateDownArgs, type MigrateUpArgs, sql } from '@payloadcms/db-postgres'

export async function up({ db, payload, req }: MigrateUpArgs): Promise<void> {
  // The `planCode` select field was added to the plans collection but no
  // migration was generated, so prod `cms.plans` lacks the column and Payload
  // crashes on boot ("column plans.plan_code does not exist"). Create the enum
  // (including `free`) and add the column. Also relax the legacy `price`
  // NOT NULL — price now comes from Stripe via public.plans, and Payload no
  // longer writes it, so inserts of new plans would otherwise fail.
  await db.execute(sql`
    DO $$ BEGIN CREATE TYPE "cms"."enum_plans_plan_code" AS ENUM('free', 'starter', 'pro', 'enterprise'); EXCEPTION WHEN duplicate_object THEN null; END $$;
    ALTER TABLE "cms"."plans" ADD COLUMN IF NOT EXISTS "plan_code" "cms"."enum_plans_plan_code";
    ALTER TABLE "cms"."plans" ALTER COLUMN "price" DROP NOT NULL;
  `)
}

export async function down({ db, payload, req }: MigrateDownArgs): Promise<void> {
  await db.execute(sql`
    ALTER TABLE "cms"."plans" DROP COLUMN IF EXISTS "plan_code";
    DROP TYPE IF EXISTS "cms"."enum_plans_plan_code";
  `)
}
