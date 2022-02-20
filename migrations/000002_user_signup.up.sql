-- User Name
ALTER TABLE users ADD name VARCHAR(100);

UPDATE users SET name = 'Ed Admin' where username = 'admin';
UPDATE users SET name = 'Ulani User' where username = 'user1';
UPDATE users SET name = 'Uriah User' where username = 'user2';

UPDATE users SET name = username WHERE name is null;


-- User EMail
ALTER TABLE users ADD email VARCHAR(100);

UPDATE users SET email = 'admin@baralga.com' where username = 'admin';
UPDATE users SET email = 'user1@baralga.com' where username = 'user1';
UPDATE users SET email = 'user2@baralga.com' where username = 'user2';

UPDATE users SET username = 'admin@baralga.com' where username = 'admin';
UPDATE users SET username = 'user1@baralga.com' where username = 'user1';
UPDATE users SET username = 'user2@baralga.com' where username = 'user2';

UPDATE users SET email = username WHERE email is null;


-- User EMail Confirmation
CREATE TABLE user_confirmations (
  user_confirmation_id  UUID NOT NULL,
  user_id               UUID NOT NULL
);

ALTER TABLE user_confirmations
ADD CONSTRAINT fk_user_confirmations_users
FOREIGN KEY (user_id) REFERENCES users (user_id);

ALTER TABLE user_confirmations
ADD CONSTRAINT pk_user_confirmations PRIMARY KEY (user_confirmation_id);
