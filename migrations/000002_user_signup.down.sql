-- User Name
ALTER TABLE users DROP COLUMN name;

-- User EMail
ALTER TABLE users DROP COLUMN email;

-- User EMail Confirmation
DROP TABLE user_confirmations;