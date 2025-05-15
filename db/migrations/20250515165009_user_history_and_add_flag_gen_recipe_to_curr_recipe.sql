-- +goose Up
-- +goose StatementBegin
CREATE TABLE user_cooking_history (
    user_id int,
    recipe_id int,
    is_generated bool DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

ALTER TABLE current_recipe
    ADD COLUMN is_generated bool DEFAULT false;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE current_recipe
    DROP COLUMN is_generated;
DROP TABLE user_cooking_history;

-- +goose StatementEnd
