-- +goose Up
-- +goose StatementBegin

CREATE TABLE generated_recipes (
    id BIGSERIAL PRIMARY KEY,
    user_id INT,
    name TEXT,
    description TEXT,
    healthscore DOUBLE PRECISION,
    dish_types JSON,
    diets JSON,
    ready_in_minutes INT,
    steps JSON,
    servings INT,
    total_steps INT,
    UNIQUE (user_id, id)
);

CREATE TABLE generated_recipes_versions (
    id INT,
    user_id INT,
    version INT,
    name TEXT,
    description TEXT,
    healthscore DOUBLE PRECISION,
    dish_types JSON,
    diets JSON,
    ready_in_minutes INT,
    steps JSON,
    servings INT,
    total_steps INT,
    FOREIGN KEY (id) REFERENCES generated_recipes(id) ON DELETE CASCADE,
    UNIQUE (id, user_id, version)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE  generated_recipes;
DROP TABLE generated_recipes_versions;
-- +goose StatementEnd
