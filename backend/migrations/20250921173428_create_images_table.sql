-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS images
(
    id          UUID PRIMARY KEY   DEFAULT gen_random_uuid(),
    filename    TEXT      NOT NULL,
    path        TEXT      NOT NULL,
    action      TEXT      NOT NULL,
    params      JSONB,
    status      TEXT      NOT NULL DEFAULT 'pending',
    created_at  TIMESTAMP NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP EXTENSION IF EXISTS "uuid-ossp";
DROP TABLE IF EXISTS images;
-- +goose StatementEnd
