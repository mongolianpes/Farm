CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE users (
    user_id serial PRIMARY KEY,
    login varchar(20) NOT NULL,
    name varchar(15) NOT NULL,
    password char(192) NOT NULL,
    avatar_path text,
    embedding vector(768),
    UNIQUE(login)
);

CREATE TABLE sessions (
    user_id int references users(user_id) ON DELETE CASCADE ON UPDATE CASCADE,
    session_id char(64) NOT NULL,
    create_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(session_id)
);

CREATE TYPE announcement_category AS ENUM('Участки', 'Животноводство', 'Растеневодство', 'Другое');

CREATE TABLE announcements (
    announcement_id serial PRIMARY KEY,
    title varchar(10) NOT NULL,
    description varchar(100) NOT NULL,
    category announcement_category NOT NULL,
    announcement_author_id int references users(user_id) ON DELETE CASCADE ON UPDATE CASCADE,
    create_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    embedding vector(768),
    images_path text[] CHECK (cardinality(images_path) <= 10) DEFAULT '{}'::text[]
);

CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_announcements_title_trgm ON announcements USING gin (title gin_trgm_ops);
CREATE INDEX idx_announcements_description_trgm ON announcements USING gin (description gin_trgm_ops);

CREATE TABLE messages (
    message_id serial PRIMARY KEY,
    sender_id int references users(user_id) ON DELETE CASCADE ON UPDATE CASCADE,
    received_id int references users(user_id) ON DELETE CASCADE ON UPDATE CASCADE,
    create_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    message text NOT NULL
);