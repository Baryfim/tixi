INSERT INTO users (email, code, code_expiration)
VALUES ($1, $2, NOW() + INTERVAL '10 minutes')
ON CONFLICT (email) 
DO UPDATE SET code = EXCLUDED.code, code_expiration = EXCLUDED.code_expiration;
