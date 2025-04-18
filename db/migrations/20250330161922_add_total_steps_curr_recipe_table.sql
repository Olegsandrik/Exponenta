-- +goose Up
-- +goose StatementBegin
ALTER TABLE current_recipe
    ADD COLUMN total_steps int DEFAULT 0;

ALTER TABLE recipes
    ADD COLUMN total_steps int DEFAULT 0;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE current_recipe
    DROP COLUMN total_steps;

ALTER TABLE recipes
    DROP COLUMN total_steps;
-- +goose StatementEnd
