-- +goose Up
-- +goose StatementBegin
ALTER TABLE recipes
    ADD COLUMN lang Text DEFAULT 'eng';

CREATE TABLE IF NOT EXISTS current_recipe (
     user_id INT UNIQUE,
     recipe_id INT,
     name TEXT,
     current_step_num INT DEFAULT 1 CHECK (current_step_num >=1),
     PRIMARY KEY (user_id)
);

CREATE TABLE timers (
    timer_id SERIAL PRIMARY KEY,
    user_id INT,
    step_num INT,
    description TEXT,
    end_time TIMESTAMP,
    UNIQUE (user_id, step_num),
    FOREIGN KEY (user_id) REFERENCES current_recipe(user_id) ON DELETE CASCADE
);

CREATE TABLE current_recipe_step (
    user_id INT,
    recipe_id INT,
    step_num INT,
    step TEXT,
    ingredients JSON,
    equipment JSON,
    length JSON,
    FOREIGN KEY (user_id) REFERENCES current_recipe(user_id) ON DELETE CASCADE
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS timers;
DROP TABLE IF EXISTS current_recipe_step;
DROP TABLE IF EXISTS current_recipe;
ALTER TABLE recipes DROP COLUMN lang;

-- +goose StatementEnd
