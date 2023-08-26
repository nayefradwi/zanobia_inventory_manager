DROP TABLE IF EXISTS users;
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

DROP INDEX IF EXISTS idx_email;
CREATE UNIQUE INDEX idx_email ON users(email);

DROP TABLE IF EXISTS permissions;
CREATE TABLE permissions (
    id SERIAL PRIMARY KEY,
    handle VARCHAR(50) NOT NULL,
    name VARCHAR(50) NOT NULL,
    description VARCHAR(255) NOT NULL,
    is_secret BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

DROP INDEX IF EXISTS idx_handle;
CREATE UNIQUE INDEX idx_handle ON permissions(handle);

DROP TABLE IF EXISTS user_permissions;
CREATE TABLE user_permissions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    permission_handle VARCHAR(50) NOT NULL REFERENCES permissions(handle),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

DROP INDEX IF EXISTS idx_user_permission;
CREATE UNIQUE INDEX idx_user_permission ON user_permissions(user_id, permission_handle);

DROP TABLE IF EXISTS roles;
CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    description VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

DROP INDEX IF EXISTS idx_name;
CREATE UNIQUE INDEX idx_name ON roles(name);

DROP TABLE IF EXISTS role_permissions;
CREATE TABLE role_permissions (
    id SERIAL PRIMARY KEY,
    role_id INTEGER NOT NULL REFERENCES roles(id),
    permission_handle VARCHAR(50) NOT NULL REFERENCES permissions(handle),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

DROP INDEX IF EXISTS idx_role_permission;
CREATE UNIQUE INDEX idx_role_permission ON role_permissions(role_id, permission_handle);

CREATE TABLE units (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    symbol VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

DROP INDEX IF EXISTS idx_name;
CREATE UNIQUE INDEX idx_name ON units(name);

DROP TABLE IF EXISTS unit_conversions;
CREATE TABLE unit_conversions (
    id SERIAL PRIMARY KEY,
    unit_id INTEGER NOT NULL REFERENCES units(id),
    conversion_unit_id INTEGER NOT NULL REFERENCES units(id),
    conversion_factor NUMERIC(10, 2) NOT NULL
);

DROP INDEX IF EXISTS idx_unit_conversion;
CREATE UNIQUE INDEX idx_unit_conversion ON unit_conversions(unit_id, conversion_unit_id);

DROP TABLE IF EXISTS translations;
CREATE TABLE translations (
    id SERIAL PRIMARY KEY,
    entity_id INTEGER NOT NULL,
    translated_entity_id INTEGER NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    locale VARCHAR(2) NOT NULL DEFAULT 'ar',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

DROP INDEX IF EXISTS idx_entity;
CREATE UNIQUE INDEX idx_entity ON translations(entity_id, entity_type, locale);