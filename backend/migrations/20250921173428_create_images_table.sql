-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE INDEX idx_images_original_id ON images (original_id);
CREATE TABLE IF NOT EXISTS images
(
    id          UUID PRIMARY KEY   DEFAULT gen_random_uuid(),
    original_id UUID      REFERENCES images (id) ON DELETE SET NULL,
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
DROP INDEX IF EXISTS idx_images_original_id;
DROP TABLE IF EXISTS images;
-- +goose StatementEnd
