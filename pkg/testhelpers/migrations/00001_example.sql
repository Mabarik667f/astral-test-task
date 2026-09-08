-- +goose Up
-- +goose StatementBegin
CREATE TABLE example (
    id UUID PRIMARY KEY
);
-- +goose StatementEnd

-- +goose Down
DROP TABLE example;
-- +goose StatementBegin
-- +goose StatementEnd
