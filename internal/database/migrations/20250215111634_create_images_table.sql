-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS images (
    "id" SERIAL PRIMARY KEY NOT NULL,
    "user_id" UUID NOT NULL,
    "image_url" TEXT NOT NULL,
    "filename" VARCHAR(255) NOT NULL,
    "format" VARCHAR(10) NOT NULL,
    "alt" VARCHAR(255) NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE ON UPDATE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS images;
-- +goose StatementEnd
