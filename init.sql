DROP TABLE IF EXISTS user_permissions;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS unit_conversions;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS units;
DROP TABLE IF EXISTS unit_translations;
DROP TABLE IF EXISTS warehouses;
DROP TABLE IF EXISTS ingredients;
DROP TABLE IF EXISTS ingredient_translations;
DROP TABLE IF EXISTS inventory;


DROP INDEX IF EXISTS idx_email;
DROP INDEX IF EXISTS idx_handle;
DROP INDEX IF EXISTS idx_role_name;
DROP INDEX IF EXISTS idx_user_permission;
DROP INDEX IF EXISTS idx_role_permission;
DROP INDEX IF EXISTS idx_unit_name;
DROP INDEX IF EXISTS idx_unit_conversion;
DROP INDEX IF EXISTS idx_unit_translations;
DROP INDEX IF EXISTS idx_warehouse_name;
DROP INDEX IF EXISTS idx_ingredient_translation;
DROP INDEX IF EXISTS ingredients_unit_quantity_idx;
DROP INDEX IF EXISTS inventory_warehouse_ingredient_idx;

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    password VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE permissions (
    id SERIAL PRIMARY KEY,
    handle VARCHAR(50) NOT NULL,
    name VARCHAR(50) NOT NULL,
    description VARCHAR(255) NOT NULL,
    is_secret BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_permissions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    permission_handle VARCHAR(50) NOT NULL REFERENCES permissions(handle),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    description VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE role_permissions (
    id SERIAL PRIMARY KEY,
    role_id INTEGER NOT NULL REFERENCES roles(id),
    permission_handle VARCHAR(50) NOT NULL REFERENCES permissions(handle),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE units (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE unit_translations (
    id SERIAL PRIMARY KEY,
    unit_id INTEGER NOT NULL REFERENCES units(id),
    language_code VARCHAR(2) NOT NULL DEFAULT 'en',
    name VARCHAR(50) NOT NULL,
    symbol VARCHAR(10) NOT NULL
);

CREATE TABLE unit_conversions (
    id SERIAL PRIMARY KEY,
    unit_id INTEGER NOT NULL REFERENCES units(id),
    conversion_unit_id INTEGER NOT NULL REFERENCES units(id),
    conversion_factor NUMERIC(10, 5) NOT NULL
);

CREATE TABLE warehouses (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    lat  DOUBLE PRECISION NOT NULL,
    lng  DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE ingredients (
    id SERIAL PRIMARY KEY,
    price NUMERIC(10, 2) NOT NULL,
    standard_quantity NUMERIC(10, 5) NOT NULL,
    standard_unit_id INTEGER NOT NULL REFERENCES units(id),
    expires_in_days INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE ingredient_translations(
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    brand VARCHAR(50),
    language_code VARCHAR(2) NOT NULL DEFAULT 'en',
    ingredient_id INTEGER NOT NULL REFERENCES ingredients(id)
);

CREATE TABLE inventory (
    id SERIAL PRIMARY KEY,
    ingredient_id INTEGER NOT NULL REFERENCES ingredients(id),
    warehouse_id INTEGER NOT NULL REFERENCES warehouses(id),
    quantity NUMERIC(10, 5) NOT NULL,
    unit_id INTEGER NOT NULL REFERENCES units(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_email ON users(email);
CREATE UNIQUE INDEX idx_handle ON permissions(handle);
CREATE UNIQUE INDEX idx_user_permission ON user_permissions(user_id, permission_handle);
CREATE UNIQUE INDEX idx_role_name ON roles(name);
CREATE UNIQUE INDEX idx_role_permission ON role_permissions(role_id, permission_handle);
CREATE UNIQUE INDEX idx_unit_name ON unit_translations(name);
CREATE UNIQUE INDEX idx_unit_conversion ON unit_conversions(unit_id, conversion_unit_id);
CREATE UNIQUE INDEX idx_unit_translations ON unit_translations(unit_id, language_code);
CREATE UNIQUE INDEX idx_warehouse_name ON warehouses(name);
CREATE UNIQUE INDEX idx_ingredient_translation ON ingredient_translations(name, brand);
CREATE UNIQUE INDEX ingredients_unit_quantity_idx ON ingredients (standard_unit_id, standard_quantity);
CREATE UNIQUE INDEX inventory_warehouse_ingredient_idx ON inventory (warehouse_id, ingredient_id);


