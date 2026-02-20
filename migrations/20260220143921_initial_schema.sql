-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS departments (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    parent_id INTEGER REFERENCES departments(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS employees (
    id SERIAL PRIMARY KEY,
    department_id INTEGER NOT NULL REFERENCES departments(id) ON DELETE CASCADE,
    full_name VARCHAR(200) NOT NULL,
    position VARCHAR(200) NOT NULL,
    hired_at DATE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for faster parent lookups
CREATE INDEX IF NOT EXISTS idx_departments_parent_id ON departments(parent_id);

-- Index for faster department lookups for employees
CREATE INDEX IF NOT EXISTS idx_employees_department_id ON employees(department_id);

-- Unique constraint: department names must be unique within the same parent
CREATE UNIQUE INDEX IF NOT EXISTS idx_departments_name_parent ON departments(name, COALESCE(parent_id, -1));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_departments_name_parent;
DROP INDEX IF EXISTS idx_employees_department_id;
DROP INDEX IF EXISTS idx_departments_parent_id;
DROP TABLE IF EXISTS employees;
DROP TABLE IF EXISTS departments;

-- +goose StatementEnd
