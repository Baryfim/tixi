INSERT INTO users (email, code)
VALUES ($1, $2)
ON CONFLICT (email) 
DO UPDATE SET code = EXCLUDED.code;
