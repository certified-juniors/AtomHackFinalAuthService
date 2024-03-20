CREATE
EXTENSION IF NOT EXISTS moddatetime
    WITH SCHEMA public
    CASCADE;

CREATE TABLE "user"
(
    id          SERIAL PRIMARY KEY,
    email       TEXT  NOT NULL UNIQUE,
    password    BYTEA NOT NULL UNIQUE,
    name        TEXT  NOT NULL,
    surname     TEXT  NOT NULL,
    middle_name TEXT  NOT NULL,
    role        TEXT  NOT NULL,
    confirmed   BOOL        DEFAULT FALSE,
    created_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER modify_user_updated_at
    BEFORE UPDATE
    ON "user"
    FOR EACH ROW
    EXECUTE PROCEDURE public.moddatetime(updated_at);