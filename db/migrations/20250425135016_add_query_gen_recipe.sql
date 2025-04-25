-- +goose Up
-- +goose StatementBegin
ALTER TABLE generated_recipes
    ADD COLUMN query TEXT DEFAULT '';
ALTER TABLE generated_recipes_versions
    ADD COLUMN query TEXT DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE generated_recipes
    DROP COLUMN query;
ALTER TABLE generated_recipes_versions
    DROP COLUMN query;
-- +goose StatementEnd
