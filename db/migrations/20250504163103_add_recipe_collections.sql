-- +goose Up
-- +goose StatementBegin
CREATE TABLE recipes_collection (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE recipes_collection_recipes (
    collection_id INT,
    recipe_id INT,
    UNIQUE (collection_id, recipe_id),
    FOREIGN KEY (collection_id) REFERENCES recipes_collection(id) ON DELETE CASCADE,
    FOREIGN KEY (recipe_id) REFERENCES recipes(id) ON DELETE CASCADE
);

INSERT INTO recipes_collection(name) VALUES
('русская кухня'),
('прохладительные напитки'),
('фруктовый микс'),
('десерты'),
('холодные супы'),
('свежие салаты');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE recipes_collection, recipes_collection_recipes;
-- +goose StatementEnd
