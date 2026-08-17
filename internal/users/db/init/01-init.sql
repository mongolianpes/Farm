CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE users (
    user_id serial PRIMARY KEY,
    login varchar(20) NOT NULL,
    name varchar(15) NOT NULL,
    password char(192) NOT NULL,
    avatar_path text DEFAULT '',
    embedding vector(768),
    UNIQUE(login)
);