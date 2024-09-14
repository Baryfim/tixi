INSERT INTO users (email, phone, code, code_expiration)
VALUES ($1, $2, $3, NOW() + INTERVAL '10 minutes')
ON CONFLICT (email, phone) 
DO UPDATE SET code = EXCLUDED.code, code_expiration = EXCLUDED.code_expiration;
