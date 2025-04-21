-- +goose Up
-- +goose StatementBegin

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    vk_id INT DEFAULT NULL, -- для OAuth
    name text NOT NULL CHECK (length(name) <= 25 and length(name) >= 3),
    sur_name text NOT NULL CHECK (length(name) <= 25 and length(name) >= 3),
    login text UNIQUE CHECK ((vk_id IS NULL AND login IS NOT NULL)
                              OR
                             (vk_id IS NOT NULL AND login IS NULL)),
    password_hash TEXT CHECK ((vk_id IS NULL AND password_hash IS NOT NULL)
                              OR
                              (vk_id IS NOT NULL AND password_hash IS NULL)),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE favorite_recipes (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    recipe_id INT NOT NULL REFERENCES recipes(id),
    UNIQUE (user_id, recipe_id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS favorite_recipes, vk_users, users;

-- +goose StatementEnd
