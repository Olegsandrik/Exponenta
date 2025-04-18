-- +goose Up
-- +goose StatementBegin
UPDATE ingredients SET image='null.jpg'  WHERE image IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE ingredients SET image=null WHERE image='null.jpg';
-- +goose StatementEnd
