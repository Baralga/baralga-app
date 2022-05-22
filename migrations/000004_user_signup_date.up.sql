-- User signup date
ALTER TABLE user_confirmations ADD created_at timestamp not null DEFAULT CURRENT_TIMESTAMP;

