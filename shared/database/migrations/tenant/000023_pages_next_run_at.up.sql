-- Add next_run_at to pages for queue-mode scheduler (S5).
-- Column is dormant until SCHEDULER_MODE=queue is activated (S6).
-- Back-fills from last_checked_at + frequency interval so queue mode
-- can take over without re-checking every page immediately.
ALTER TABLE pages ADD COLUMN IF NOT EXISTS next_run_at TIMESTAMPTZ;

-- Back-fill: pages with a monitoring config and a known last_checked_at get
-- next_run_at = last_checked_at + interval.  Pages never checked (NULL) or
-- having no active config get next_run_at = NOW() so they are due immediately.
-- Pages whose config is Off stay NULL (excluded from queue-mode scheduling).
UPDATE pages p
   SET next_run_at = CASE
           WHEN p.last_checked_at IS NULL THEN NOW()
           ELSE p.last_checked_at + CASE mc.check_frequency
               WHEN '5m'            THEN INTERVAL '5 minutes'
               WHEN 'Every 5 minutes' THEN INTERVAL '5 minutes'
               WHEN '10m'           THEN INTERVAL '10 minutes'
               WHEN 'Every 10 minutes' THEN INTERVAL '10 minutes'
               WHEN '15m'           THEN INTERVAL '15 minutes'
               WHEN 'Every 15 minutes' THEN INTERVAL '15 minutes'
               WHEN '30m'           THEN INTERVAL '30 minutes'
               WHEN 'Every 30 minutes' THEN INTERVAL '30 minutes'
               WHEN '1h'            THEN INTERVAL '1 hour'
               WHEN 'Every hour'    THEN INTERVAL '1 hour'
               WHEN 'Every 1 hour'  THEN INTERVAL '1 hour'
               WHEN '1 hr'          THEN INTERVAL '1 hour'
               WHEN '2h'            THEN INTERVAL '2 hours'
               WHEN 'Every 2 hours' THEN INTERVAL '2 hours'
               WHEN '2 hr'          THEN INTERVAL '2 hours'
               WHEN '4h'            THEN INTERVAL '4 hours'
               WHEN 'Every 4 hours' THEN INTERVAL '4 hours'
               WHEN '4 hr'          THEN INTERVAL '4 hours'
               WHEN '6h'            THEN INTERVAL '6 hours'
               WHEN 'Every 6 hours' THEN INTERVAL '6 hours'
               WHEN '6 hr'          THEN INTERVAL '6 hours'
               WHEN '12h'           THEN INTERVAL '12 hours'
               WHEN 'Every 12 hours' THEN INTERVAL '12 hours'
               WHEN '12 hr'         THEN INTERVAL '12 hours'
               WHEN '24h'           THEN INTERVAL '1 day'
               WHEN 'Every day'     THEN INTERVAL '1 day'
               WHEN '1d'            THEN INTERVAL '1 day'
               WHEN '168h'          THEN INTERVAL '7 days'
               WHEN 'Every week'    THEN INTERVAL '7 days'
               WHEN '7d'            THEN INTERVAL '7 days'
               ELSE INTERVAL '1 hour'
           END
       END
  FROM monitoring_configs mc
 WHERE mc.page_id = p.id
   AND mc.deleted_at IS NULL
   AND mc.check_frequency <> 'Off';

-- Partial index: only non-deleted pages with a scheduled next_run_at.
-- Queue-mode due query WHERE next_run_at <= NOW() uses this index.
CREATE INDEX IF NOT EXISTS idx_pages_next_run_at
    ON pages (next_run_at)
 WHERE deleted_at IS NULL AND next_run_at IS NOT NULL;
