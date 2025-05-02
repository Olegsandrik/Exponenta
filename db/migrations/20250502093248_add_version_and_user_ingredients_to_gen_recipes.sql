-- +goose Up
-- +goose StatementBegin
ALTER TABLE generated_recipes
    ADD COLUMN version INT DEFAULT 1;
ALTER TABLE generated_recipes
    ADD COLUMN user_ingredients JSON DEFAULT '[]';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE generated_recipes
    DROP COLUMN version;
ALTER TABLE generated_recipes
    DROP COLUMN user_ingredients;
-- +goose StatementEnd
