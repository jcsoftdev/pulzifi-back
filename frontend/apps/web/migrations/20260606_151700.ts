// biome-ignore-all lint/correctness/noUnusedFunctionParameters: generated migration file — payload/req required by interface
import { type MigrateDownArgs, type MigrateUpArgs } from '@payloadcms/db-postgres'

/**
 * Snapshot-resync migration (no-op).
 *
 * The committed Payload snapshots (`*.json`) had drifted from reality: the latest
 * snapshot still described the old `landing_blocks` tables/enums even though the
 * `20260506`/`20260507` migrations already moved everything to `block_library`,
 * and later hand-written migrations (`20260604` reading_time, `20260606` plan_code)
 * never regenerated a snapshot. That stale snapshot made `migrate:create` emit a
 * phantom full-schema diff (the migration drift guard kept failing / prompting for
 * enum renames).
 *
 * This entry carries a FRESH, correct snapshot (`20260606_151700.json`) that matches
 * the current collections. The actual schema (enums, tables, columns — including
 * `*_how_it_works_steps_mock_type`, `enum_plans_plan_code`, etc.) already exists in
 * every environment from the earlier migrations, so there is nothing to apply here.
 *
 * up/down are intentional no-ops: their only job is to register the corrected
 * snapshot so future `migrate:create` runs diff against reality and the drift guard
 * stays green.
 */
export async function up({ db, payload, req }: MigrateUpArgs): Promise<void> {
  // no-op — schema already present from prior migrations; this entry only resyncs the snapshot.
}

export async function down({ db, payload, req }: MigrateDownArgs): Promise<void> {
  // no-op — see up().
}
