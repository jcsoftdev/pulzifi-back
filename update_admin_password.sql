-- Update admin user with bcrypt hash of "admin"
UPDATE public.users 
SET password_hash = '$2a$12$R9h/cIPz0gi.URNNX3kh2OPST9/PgBkqquzi.Ee4s1k.K8z3w7vSm'
WHERE email = 'admin@pulzifi.com';
