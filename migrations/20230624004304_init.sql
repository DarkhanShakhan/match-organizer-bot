-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    "name" TEXT NOT NULL,
    "username" TEXT NOT NULL UNIQUE,
    phone_number INT NOT NULL UNIQUE
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TYPE sport_type AS ENUM('football', 'volleyball', 'basketball');
-- +goose StatementEnd


-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS matches (
    id SERIAL PRIMARY KEY,
    organizer_id INT NOT NULL,
    location TEXT NOT NULL,
    sport sport_type NOT NULL,
    team_size INT NOT NULL,
    team_count INT NOT NULL,
    rent INT NOT NULL,
    start_at timestamp WITHOUT TIME ZONE NOT NULL,
    finish_at timestamp WITHOUT TIME ZONE NOT NULL,
    cancelled BOOLEAN DEFAULT false, 
    cancel_message TEXT,
    "private" BOOLEAN DEFAULT false,
    CONSTRAINT fk_user FOREIGN KEY(organizer_id) REFERENCES users(id) ON DELETE CASCADE 
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS teams (
    id SERIAL PRIMARY KEY,
    "name" TEXT,
    match_id INT NOT NULL,
    CONSTRAINT fk_match FOREIGN KEY(match_id) REFERENCES matches(id) ON DELETE CASCADE
); 
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS team_members (
    team_id INT NOT NULL,
    member_id INT NOT NULL,
    paid BOOLEAN DEFAULT false,
    confirmed BOOLEAN DEFAULT false,
    cancelled BOOLEAN DEFAULT false
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS team_members;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS matches;
DROP TYPE IF EXISTS sport_type;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
