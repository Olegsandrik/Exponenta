-- +goose Up
-- +goose StatementBegin

CREATE TABLE generated_recipes (
    id BIGSERIAL PRIMARY KEY,
    user_id INT,
    name TEXT,
    description TEXT,
    dish_types JSON,
    servings INT,
    diets JSON,
    ingredients JSON,
    ready_in_minutes INT,
    steps JSON,
    total_steps INT,
    UNIQUE (user_id, id)
);

CREATE TABLE generated_recipes_versions (
    id INT,
    user_id INT,
    version INT,
    name TEXT,
    description TEXT,
    dish_types JSON,
    servings INT,
    diets JSON,
    ingredients JSON,
    ready_in_minutes INT,
    steps JSON,
    total_steps INT,
    FOREIGN KEY (id) REFERENCES generated_recipes(id) ON DELETE CASCADE,
    UNIQUE (id, user_id, version)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE generated_recipes_versions;
DROP TABLE  generated_recipes;

-- +goose StatementEnd
