UPDATE public.plans
   SET checks_allowed_monthly = 500,
       updated_at             = NOW()
 WHERE code = 'trial';
