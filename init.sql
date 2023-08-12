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
