-- migrate:up tx=true
-- Add categories table for organizing posts
CREATE TABLE categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create junction table for many-to-many relationship
CREATE TABLE post_categories (
    post_id INTEGER NOT NULL,
    category_id INTEGER NOT NULL,
    PRIMARY KEY (post_id, category_id),
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
);

CREATE INDEX idx_post_categories_category_id ON post_categories(category_id);

-- migrate:down tx=true
DROP INDEX IF EXISTS idx_post_categories_category_id;
DROP TABLE IF EXISTS post_categories;
DROP TABLE IF EXISTS categories;
