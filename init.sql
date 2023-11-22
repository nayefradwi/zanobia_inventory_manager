-- AUTHENTICATION TABLES --
DROP TABLE IF EXISTS user_permissions CASCADE;
DROP TABLE IF EXISTS role_permissions CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS permissions CASCADE;
DROP TABLE IF EXISTS roles CASCADE;

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    password VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE permissions (
    id SERIAL PRIMARY KEY,
    handle VARCHAR(50) NOT NULL UNIQUE,
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
    name VARCHAR(50) NOT NULL UNIQUE,
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

DROP INDEX IF EXISTS idx_user_permission CASCADE;
DROP INDEX IF EXISTS idx_role_permission CASCADE;

CREATE UNIQUE INDEX idx_user_permission ON user_permissions(user_id, permission_handle);
CREATE UNIQUE INDEX idx_role_permission ON role_permissions(role_id, permission_handle);
-- END AUTHENTICATION TABLES --

-- UNIT TABLES --
DROP TABLE IF EXISTS units CASCADE;
DROP TABLE IF EXISTS unit_translations CASCADE;
DROP TABLE IF EXISTS unit_conversions CASCADE;

CREATE TABLE units (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE unit_translations (
    id SERIAL PRIMARY KEY,
    unit_id INTEGER NOT NULL REFERENCES units(id),
    language_code VARCHAR(2) NOT NULL DEFAULT 'en',
    name VARCHAR(50) UNIQUE NOT NULL,
    symbol VARCHAR(10) NOT NULL
);

CREATE TABLE unit_conversions (
    id SERIAL PRIMARY KEY,
    to_unit_id INTEGER NOT NULL REFERENCES units(id),
    from_unit_id INTEGER NOT NULL REFERENCES units(id),
    conversion_factor NUMERIC(12, 6) NOT NULL
);

DROP INDEX IF EXISTS idx_unit_name CASCADE;
DROP INDEX IF EXISTS idx_unit_conversion CASCADE;

CREATE UNIQUE INDEX idx_unit_conversion ON unit_conversions(to_unit_id, from_unit_id);
CREATE UNIQUE INDEX idx_unit_translations ON unit_translations(unit_id, language_code);

-- END UNIT TABLES --

-- WAREHOUSE TABLES --
DROP TABLE IF EXISTS warehouses CASCADE;
DROP TABLE IF EXISTS user_warehouses CASCADE;

CREATE TABLE warehouses (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    lat  DOUBLE PRECISION NOT NULL,
    lng  DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_warehouses (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    warehouse_id INTEGER NOT NULL REFERENCES warehouses(id)
);

DROP INDEX IF EXISTS idx_warehouse_name CASCADE;
DROP INDEX IF EXISTS idx_user_warehouse CASCADE;

CREATE UNIQUE INDEX idx_warehouse_name ON warehouses(name);
CREATE UNIQUE INDEX idx_user_warehouse ON user_warehouses(user_id, warehouse_id);
-- END WAREHOUSE TABLES --


-- PRODUCT TABLES --
DROP TABLE IF EXISTS categories CASCADE;
DROP TABLE IF EXISTS category_translations CASCADE;
DROP TABLE IF EXISTS product_translations CASCADE;
DROP TABLE IF EXISTS products CASCADE;

CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE category_translations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    category_id INTEGER NOT NULL REFERENCES categories(id),
    language_code VARCHAR(2) NOT NULL DEFAULT 'en'
);

CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    image VARCHAR(255),
    category_id INTEGER REFERENCES categories(id),
    is_archived BOOLEAN NOT NULL DEFAULT FALSE,
    is_ingredient BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE product_translations (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id),
    name VARCHAR(50) UNIQUE NOT NULL,
    description VARCHAR(255),
    language_code VARCHAR(2) NOT NULL DEFAULT 'en'
);

DROP INDEX IF EXISTS idx_category_translation CASCADE;
DROP INDEX IF EXISTS idx_product_translation_name CASCADE;
DROP INDEX IF EXISTS idx_product_translation CASCADE;
DROP INDEX IF EXISTS idx_product_is_archived CASCADE;
DROP INDEX IF EXISTS idx_product_category CASCADE;
DROP INDEX IF EXISTS idx_product_created_at CASCADE;
DROP INDEX IF EXISTS idx_product_is_ingredient CASCADE;

CREATE UNIQUE INDEX idx_category_translation ON category_translations(category_id, language_code);
CREATE UNIQUE INDEX idx_product_translation ON product_translations(product_id, language_code);
CREATE INDEX idx_product_is_archived ON products(is_archived);
CREATE INDEX idx_product_category ON products(category_id);
CREATE INDEX idx_product_created_at ON products(created_at);
CREATE INDEX idx_product_is_ingredient ON products(is_ingredient);

-- END PRODUCT TABLES --

-- VARIANT TABLES --
DROP TABLE IF EXISTS product_options CASCADE;
DROP TABLE IF EXISTS product_option_translations CASCADE;
DROP TABLE IF EXISTS product_option_values CASCADE;
DROP TABLE IF EXISTS product_option_values_translations CASCADE;
DROP TABLE IF EXISTS product_variants CASCADE;
DROP TABLE IF EXISTS product_variant_translations CASCADE;
DROP TABLE IF EXISTS product_variant_values CASCADE;

CREATE TABLE product_options (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id),
    name VARCHAR(50) NOT NULL,
    language_code VARCHAR(2) NOT NULL DEFAULT 'en'
);


create table product_option_values(
    id SERIAL PRIMARY KEY,
    product_option_id INTEGER NOT NULL REFERENCES product_options(id),
    value VARCHAR(50) NOT NULL,
    language_code VARCHAR(2) NOT NULL DEFAULT 'en'
);

CREATE TABLE product_variants (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id),
    sku VARCHAR(36) UNIQUE NOT NULL,
    image VARCHAR(255),
    price DECIMAL(12,2) NOT NULL,
    width_in_cm DECIMAL(12, 2),
    height_in_cm DECIMAL(12, 2),
    depth_in_cm DECIMAL(12, 2),
    weight_in_g DECIMAL(12, 2),
    standard_unit_id INTEGER NOT NULL REFERENCES units(id),
    is_archived BOOLEAN NOT NULL DEFAULT FALSE,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    expires_in_days INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE product_variant_translations (
    id SERIAL PRIMARY KEY,
    product_variant_id INTEGER NOT NULL REFERENCES product_variants(id),
    product_id INTEGER NOT NULL REFERENCES products(id),
    name VARCHAR(50) NOT NULL,
    language_code VARCHAR(2) NOT NULL DEFAULT 'en'
);

CREATE TABLE product_variant_values (
    id SERIAL PRIMARY KEY,
    product_option_value_id INTEGER NOT NULL REFERENCES product_option_values(id),
    product_variant_id INTEGER NOT NULL REFERENCES product_variants(id)
);


DROP INDEX IF EXISTS idx_product_options CASCADE;
DROP INDEX IF EXISTS idx_product_variant_sku CASCADE;
DROP INDEX IF EXISTS idx_product_variant_created_at CASCADE;
DROP INDEX IF EXISTS idx_product_variant_is_archived CASCADE;
DROP INDEX IF EXISTS idx_product_variant_is_default CASCADE;
DROP INDEX IF EXISTS idx_product_variant_price CASCADE;
DROP INDEX IF EXISTS idx_product_variant_translation CASCADE;
DROP INDEX IF EXISTS idx_product_variant_translation_id CASCADE;
DROP INDEX IF EXISTS idx_product_variant_values CASCADE;
DROP INDEX IF EXISTS idx_product_option_values CASCADE;

CREATE UNIQUE INDEX idx_product_options ON product_options(name, language_code, product_id);
CREATE UNIQUE INDEX idx_product_variant_sku ON product_variants(sku);
CREATE INDEX idx_product_variant_created_at ON product_variants(created_at);
CREATE INDEX idx_product_variant_is_archived ON product_variants(is_archived);
CREATE INDEX idx_product_variant_is_default ON product_variants(is_default, product_id);
CREATE INDEX idx_product_variant_price ON product_variants(price);
CREATE UNIQUE INDEX idx_product_variant_translation ON product_variant_translations(product_id, language_code, name);
CREATE INDEX idx_product_variant_translation_id on product_variant_translations(product_variant_id, language_code);
CREATE UNIQUE INDEX idx_product_variant_values on product_variant_values(product_option_value_id, product_variant_id);
CREATE UNIQUE INDEX idx_product_option_values on product_option_values(value, language_code, product_option_id);
-- END VARIANT TABLES --

-- RECIPE AND BATCHES TABLES --
DROP TABLE IF EXISTS recipes CASCADE;
DROP TABLE IF EXISTS batches CASCADE;

CREATE TABLE recipes (
    id SERIAL PRIMARY KEY,
    result_variant_id INTEGER NOT NULL REFERENCES product_variants(id),
    recipe_variant_id INTEGER NOT NULL REFERENCES product_variants(id),
    unit_id INTEGER NOT NULL REFERENCES units(id),
    quantity NUMERIC(12, 4) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE batches (
    id SERIAL PRIMARY KEY,
    sku VARCHAR(36) NOT NULL REFERENCES product_variants(sku),
    warehouse_id INTEGER NOT NULL REFERENCES warehouses(id),
    quantity NUMERIC(12, 4) NOT NULL,
    unit_id INTEGER NOT NULL REFERENCES units(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL
);

DROP INDEX IF EXISTS idx_recipe CASCADE;
DROP INDEX IF EXISTS idx_batch CASCADE;

CREATE UNIQUE INDEX idx_recipe ON recipes(result_variant_id, recipe_variant_id);
CREATE UNIQUE INDEX idx_batch ON batches(sku, warehouse_id, expires_at);

-- END RECIPE AND BATCHES TABLES --