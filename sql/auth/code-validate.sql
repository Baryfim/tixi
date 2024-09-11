SELECT code, code_expiration
FROM users
WHERE email = $1;
