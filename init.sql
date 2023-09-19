DROP TABLE IF EXISTS user_permissions;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS unit_conversions;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS units;
DROP TABLE IF EXISTS unit_translations;
DROP TABLE IF EXISTS warehouses;


DROP INDEX IF EXISTS idx_email;
DROP INDEX IF EXISTS idx_handle;
DROP INDEX IF EXISTS idx_role_name;
DROP INDEX IF EXISTS idx_user_permission;
DROP INDEX IF EXISTS idx_role_permission;
DROP INDEX IF EXISTS idx_unit_name;
DROP INDEX IF EXISTS idx_unit_conversion;
DROP INDEX IF EXISTS idx_unit_translations;
DROP INDEX IF EXISTS idx_warehouse_name;

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


CREATE UNIQUE INDEX idx_email ON users(email);
CREATE UNIQUE INDEX idx_handle ON permissions(handle);
CREATE UNIQUE INDEX idx_user_permission ON user_permissions(user_id, permission_handle);
CREATE UNIQUE INDEX idx_role_name ON roles(name);
CREATE UNIQUE INDEX idx_role_permission ON role_permissions(role_id, permission_handle);
CREATE UNIQUE INDEX idx_unit_name ON unit_translations(name);
CREATE UNIQUE INDEX idx_unit_conversion ON unit_conversions(unit_id, conversion_unit_id);
CREATE UNIQUE INDEX idx_unit_translations ON unit_translations(unit_id, language_code);
CREATE UNIQUE INDEX idx_warehouse_name ON warehouses(name);
