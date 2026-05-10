# CMS Production Deploy — feat/cms-payload-v3

## Deploy chain (runs on every container start)

```
bun scripts/cms-migrate.ts
  → Payload applies pending migrations (creates/updates cms schema)
  → seedCMSIfEmpty()
      → collections empty? → seed all default content (first deploy only)
      → collections have data? → skip (editor changes preserved)
→ next start
```

If `cms-migrate.ts` fails → `&&` short-circuits → `next start` never runs →
container crashes → **Dokploy keeps the old container alive automatically**.

---

## Step 1 — Add env vars in Dokploy (before deploying)

| Var | Status | Notes |
|-----|--------|-------|
| `PAYLOAD_SECRET` | **NEW — required** | Strong random string (32+ chars). CMS admin JWT signing key. |
| `SEED_SECRET` | **NEW — recommended** | Guards the force-seed endpoint. |
| `DB_USER`, `DB_PASSWORD`, `DB_HOST`, `DB_PORT`, `DB_NAME` | Already set | Same DB as Go backend. Verify present. |
| `MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY`, `MINIO_ENDPOINT`, `MINIO_BUCKET` | Already set | Payload uses them for media uploads. Verify present. |
| `NEXT_PUBLIC_SERVER_URL` | Already set | Should be `https://pulzifi.com`. Used by Payload for live preview links. |

Generate `PAYLOAD_SECRET`:
```bash
openssl rand -base64 32
```

---

## Step 2 — Merge and deploy

```
feat/cms-payload-v3 → PR → merge to main → Dokploy auto-deploys
```

**First deploy behavior:**
- `cms` schema doesn't exist yet → Payload migrations create it
- Collections are empty → seed fires → landing + pricing data populated
- `/cms-admin` is live

---

## Step 3 — Subsequent deploys (zero risk to editor data)

Guard in `frontend/apps/web/features/cms/seed.ts` line ~620:

```typescript
if (plansEmpty || pagesEmpty) {
  // only seeds when collections are empty
}
return { seeded: false, reason: 'data-exists' } // normal path on every redeploy
```

Editor changes are **never touched** on redeploys. Only schema migrations run (additive, safe).

---

## Escape hatches

```bash
# Force re-seed without dropping schema (overwrites CMS content, keeps schema)
curl -X POST https://pulzifi.com/api/seed-cms?force=1 \
  -H "x-seed-secret: $SEED_SECRET"

# Nuclear: drop cms schema entirely and re-seed (DESTROYS all editor data)
bun scripts/cms-migrate.ts --fresh
```

---

## Risk summary

| Risk | Mitigation |
|------|-----------|
| Missing env var | Container crashes, Dokploy keeps old container |
| Migration fails | Same — old container stays alive |
| Seed fires on redeploy | Impossible — guard checks for existing data |
| Editor content overwritten | Impossible — `seedAll` only called when collections are empty |
