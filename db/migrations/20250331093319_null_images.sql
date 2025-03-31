-- +goose Up
-- +goose StatementBegin
UPDATE ingredients SET image='null.jpg'  WHERE image IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
