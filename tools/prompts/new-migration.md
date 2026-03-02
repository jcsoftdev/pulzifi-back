# Prompt: New Migration

## When to Use

When adding a new database table or modifying an existing schema.

## Prompt

```
Create a database migration for: {description}

Scope: {public | tenant}

Follow the project's migration conventions:

1. Use `./tools/scripts/new-migration.sh {scope} {description}` to scaffold the files
   - Files go in `shared/database/migrations/{scope}/`
   - Format: `{NNNNNN}_{description}.up.sql` / `{NNNNNN}_{description}.down.sql`

2. Up migration:
   - CREATE TABLE with all columns, constraints, and indexes
   - Use UUID for primary keys: `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`
   - Add `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
   - Add `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
   - Add foreign key constraints where appropriate
   - Add indexes for frequently queried columns

3. Down migration:
   - DROP TABLE IF EXISTS (or reverse the up migration exactly)

4. Tenant migrations apply to ALL tenant schemas automatically

Schema changes:
- {describe the tables/columns to add or modify}

Relationships:
- {describe foreign keys and references}
```
