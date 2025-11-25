ALTER TABLE conversations DROP CONSTRAINT IF EXISTS fk_last_message;

DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS user_conversations;
DROP TABLE IF EXISTS conversations;
DROP TABLE IF EXISTS contacts;
DROP TABLE IF EXISTS users;