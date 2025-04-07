-- +goose Up
-- +goose StatementBegin
ALTER TABLE recipe_ingredients
    ADD COLUMN old_name TEXT DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE recipe_ingredients
    DROP COLUMN old_name;
-- +goose StatementEnd
