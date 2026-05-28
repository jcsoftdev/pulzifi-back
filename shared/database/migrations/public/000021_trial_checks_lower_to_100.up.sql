-- Lower the trial plan's monthly check quota from 500 to 100 so the trial
-- gives a real "preview" of value without competing with the paid Starter
-- tier (Starter retains its 500 checks/month). Existing trial usage_tracking
-- rows in tenant schemas are NOT touched — the next refill / new period will
-- pull this value via the PlanAssigner sync. Active trials keep their
-- already-granted quota until the period rolls over.
UPDATE public.plans
   SET checks_allowed_monthly = 100,
       updated_at             = NOW()
 WHERE code      = 'trial'
   AND is_active = TRUE;
