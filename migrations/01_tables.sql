-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    login VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS docs (
    id UUID PRIMARY KEY,
    owner_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    file BOOLEAN NOT NULL,
    public BOOLEAN NOT NULL,
    mime VARCHAR(255) NOT NULL,
    json JSONB,
    file_path TEXT,
    created TIMESTAMP NOT NULL,

    FOREIGN KEY (owner_id) REFERENCES users (id) ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS doc_grants (
    user_id UUID NOT NULL,
    doc_id UUID NOT NULL,

    PRIMARY KEY (user_id, doc_id),

    FOREIGN KEY (doc_id) REFERENCES docs (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin 
DROP TABLE IF EXISTS doc_grants;
DROP TABLE IF EXISTS docs;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
