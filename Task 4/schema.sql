/*
CREATE TABLE friend_requests (
    id SERIAL PRIMARY KEY,
    sender_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    recipient_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE friends (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    friend_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, friend_id)
);

-- DROP TABLE IF EXISTS friends CASCADE;
-- DROP TABLE IF EXISTS friend_requests CASCADE;

*/
/*CREATE DATABASE webchatdb;

\c webchatdb;
\c webchatdb;
*/








DROP TABLE IF EXISTS message_readers;
DROP TABLE IF EXISTS message_files;
DROP TABLE IF EXISTS chat_join_requests;
DROP TABLE IF EXISTS chat_invites;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS chat_members;
DROP TABLE IF EXISTS chats;
DROP TABLE IF EXISTS temp_passwords;
DROP TABLE IF EXISTS registration_codes;
DROP TABLE IF EXISTS user_sessions;
DROP TABLE IF EXISTS users;

-- Создание таблиц
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    phone VARCHAR(20) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    username VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_seen_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(500) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL
);

CREATE TABLE registration_codes (
    id SERIAL PRIMARY KEY,
    phone VARCHAR(20) NOT NULL,
    code VARCHAR(10) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL
);

CREATE TABLE temp_passwords (
    id SERIAL PRIMARY KEY,
    phone VARCHAR(20) NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE chats (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100),
    type VARCHAR(20) NOT NULL,
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_searchable BOOLEAN DEFAULT false
);

CREATE TABLE chat_members (
    id SERIAL PRIMARY KEY,
    chat_id INTEGER REFERENCES chats(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    is_admin BOOLEAN DEFAULT false,
    UNIQUE(chat_id, user_id)
);

CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    chat_id INTEGER REFERENCES chats(id) ON DELETE CASCADE,
    sender_id INTEGER REFERENCES users(id),
    content TEXT,
    type VARCHAR(50) DEFAULT 'regular',
    is_deleted BOOLEAN DEFAULT false,
    is_edited BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE chat_invites (
    id SERIAL PRIMARY KEY,
    chat_id INTEGER REFERENCES chats(id) ON DELETE CASCADE,
    inviter_id INTEGER REFERENCES users(id),
    user_id INTEGER REFERENCES users(id),
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE chat_join_requests (
    id SERIAL PRIMARY KEY,
    chat_id INTEGER REFERENCES chats(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id),
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE message_files (
    id SERIAL PRIMARY KEY,
    message_id INTEGER REFERENCES messages(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    filepath VARCHAR(500) NOT NULL,
    filesize BIGINT NOT NULL,
    mime_type VARCHAR(100),
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE message_readers (
    id SERIAL PRIMARY KEY,
    message_id INTEGER REFERENCES messages(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    read_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(message_id, user_id)
);

/*CREATE UNIQUE INDEX friend_requests_unique_pending 
ON friend_requests (sender_id, recipient_id) 
WHERE status = 'pending';
*/

CREATE UNIQUE INDEX chat_invites_unique_pending 
ON chat_invites (chat_id, user_id) 
WHERE status = 'pending';
/*
CREATE UNIQUE INDEX chat_join_requests_unique_pending 
ON chat_join_requests (chat_id, user_id) 
WHERE status = 'pending';
*/
INSERT INTO users (phone, password_hash, username) VALUES
('79082796394', '$2a$12$v.lWrhGZs3RauWuWucevPuLoTXi.hf5PzESxNTsvR0mEC5kd0KtkO', 'Алексей'),
('79083795623', '$2a$12$AC4jySF1j4OgGDKRSR4.8uVazYc8NG.iR6mQ7vRcCONbdQOSmC/Ee', 'Мария'),
('79022383848', '$2a$12$W.yJaYcGKE6cHojol/kMduK4yvtWV/it9gqE1ia3PQLYTETwxYhqy', 'Дмитрий'),
('79996782365', '$2a$12$6zGHML.3c1GI8zPvoWro1OkPyLes/cjOxqlq3LWB5nCd9zXiY1TzO', 'Екатерина'),
('72390239038', '$2a$12$hf93LEWmpJ7kcCtVmvQDcukNjRwwyiQtCWY8Gz/gH7vz6zhQsamfG', 'Иван'),
('72342383826', '$2a$12$.8nFaYtiXxo2n24pj4V11OU04CkyVZCKOKJmPap.dqIV2rWCVa1Hy', 'Ольга');

-- 79082796394
-- aaaaaa9

-- 79083795623
-- barsik23

-- 79022383848
-- adka_lf_ewf

-- 79996782365
-- @afkjфаыфва

-- 72390239038
-- ___@fanslkf

-- 72342383826
-- _2asdf323
