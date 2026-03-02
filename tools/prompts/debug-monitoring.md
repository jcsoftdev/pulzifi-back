# Prompt: Debug Monitoring Pipeline

## When to Use

When monitoring checks are not running, changes are not detected, or insights are not generating.

## Prompt

```
Debug the monitoring pipeline issue: {describe the symptom}

Check each stage of the pipeline in order:

1. **Scheduler**: Is the scheduler running?
   - Check `ENABLE_WORKERS=true` in worker env
   - Check worker logs: `make logs service=worker`
   - Query due configs:
     ```sql
     SET search_path TO {tenant}, public;
     SELECT id, page_id, check_frequency, is_active, next_check_at
     FROM monitoring_configs WHERE is_active = true ORDER BY next_check_at;
     ```

2. **Worker Pool**: Are checks being dispatched?
   - Look for "dispatching check" or similar log entries
   - Check for goroutine errors or panics in worker logs

3. **Extractor**: Is Playwright capturing correctly?
   - Health check: `curl http://localhost:3005/health`
   - Logs: `make logs service=extractor`
   - Test URL manually: target URL reachable from Docker network?

4. **Change Detection**: Is the hash comparison working?
   - Query recent checks:
     ```sql
     SELECT id, status, content_hash, change_detected, extractor_failed, checked_at
     FROM checks WHERE page_id = '{pageId}' ORDER BY checked_at DESC LIMIT 5;
     ```
   - If `extractor_failed = true`, the issue is in step 3

5. **Insight Generation**: Is the LLM responding?
   - Check `OPENROUTER_API_KEY` is set
   - Look for API errors in worker logs
   - Query insights: `SELECT * FROM insights WHERE check_id = '{checkId}';`

6. **SSE Delivery**: Are events reaching the client?
   - Test: `curl -N http://tenant.localhost:3000/api/v1/monitoring/checks/page/{pageId}/stream`
   - CheckBroker drops slow subscribers (buffered channel size 1)
```
