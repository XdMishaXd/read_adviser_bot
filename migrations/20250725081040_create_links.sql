-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS links (
  id       BIGSERIAL PRIMARY KEY,
  link     TEXT NOT NULL,
  username TEXT NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE links
-- +goose StatementEnd
