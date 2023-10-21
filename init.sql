





DROP INDEX IF EXISTS idx_recipe CASCADE;
DROP INDEX IF EXISTS idx_batch CASCADE;

-- AUTHENTICATION TABLES --
DROP TABLE IF EXISTS user_permissions CASCADE;
DROP TABLE IF EXISTS role_permissions CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS permissions CASCADE;
DROP TABLE IF EXISTS roles CASCADE;

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

DROP INDEX IF EXISTS idx_email CASCADE;
DROP INDEX IF EXISTS idx_handle CASCADE;
DROP INDEX IF EXISTS idx_role_name CASCADE;
DROP INDEX IF EXISTS idx_user_permission CASCADE;
DROP INDEX IF EXISTS idx_role_permission CASCADE;

CREATE UNIQUE INDEX idx_email ON users(email);
CREATE UNIQUE INDEX idx_handle ON permissions(handle);
CREATE UNIQUE INDEX idx_user_permission ON user_permissions(user_id, permission_handle);
CREATE UNIQUE INDEX idx_role_name ON roles(name);
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

CREATE TABLE warehouses (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    lat  DOUBLE PRECISION NOT NULL,
    lng  DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

DROP INDEX IF EXISTS idx_warehouse_name CASCADE;

CREATE UNIQUE INDEX idx_warehouse_name ON warehouses(name);
-- END WAREHOUSE TABLES --

-- INGREDIENT AND INVENTORY TABLES --

DROP TABLE IF EXISTS ingredients CASCADE;
DROP TABLE IF EXISTS ingredient_translations CASCADE;
DROP TABLE IF EXISTS inventories CASCADE;

CREATE TABLE ingredients (
    id SERIAL PRIMARY KEY,
    price DECIMAL(12,2) NOT NULL,
    standard_quantity NUMERIC(10, 3) NOT NULL,
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

CREATE TABLE inventories (
    id SERIAL PRIMARY KEY,
    ingredient_id INTEGER NOT NULL REFERENCES ingredients(id),
    warehouse_id INTEGER NOT NULL REFERENCES warehouses(id),
    quantity NUMERIC(12, 4) NOT NULL,
    unit_id INTEGER NOT NULL REFERENCES units(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

DROP INDEX IF EXISTS idx_ingredient_translation_name_and_brand CASCADE;
DROP INDEX IF EXISTS idx_ingredient_translation CASCADE;
DROP INDEX IF EXISTS ingredients_unit_quantity_idx CASCADE;
DROP INDEX IF EXISTS inventory_warehouse_ingredient_idx CASCADE;
DROP INDEX IF EXISTS idx_inventory_updated_at CASCADE;

CREATE UNIQUE INDEX idx_ingredient_translation_name_and_brand ON ingredient_translations(name, brand);
CREATE UNIQUE INDEX idx_ingredient_translation ON ingredient_translations(ingredient_id, language_code);
CREATE INDEX ingredients_unit_quantity_idx ON ingredients (standard_unit_id, standard_quantity);
CREATE UNIQUE INDEX inventory_warehouse_ingredient_idx ON inventories (warehouse_id, ingredient_id);
CREATE INDEX idx_inventory_updated_at ON inventories (updated_at);
-- END INGREDIENT AND INVENTORY TABLES --

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

CREATE UNIQUE INDEX idx_category_translation ON category_translations(category_id, language_code);
CREATE UNIQUE INDEX idx_product_translation ON product_translations(product_id, language_code);
CREATE INDEX idx_product_is_archived ON products(is_archived);
CREATE INDEX idx_product_category ON products(category_id);
CREATE INDEX idx_product_created_at ON products(created_at);

-- END PRODUCT TABLES --

-- VARIANT TABLES --
DROP TABLE IF EXISTS product_variant_selected_values CASCADE;
DROP TABLE IF EXISTS product_variant_translations CASCADE;
DROP TABLE IF EXISTS product_variants CASCADE;
DROP TABLE IF EXISTS variant_values CASCADE;
DROP TABLE IF EXISTS variants CASCADE;
DROP TABLE IF EXISTS variant_translations CASCADE;
DROP TABLE IF EXISTS product_options CASCADE;
DROP TABLE IF EXISTS product_selected_values CASCADE;

CREATE TABLE variants (
    id SERIAL PRIMARY KEY,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE variant_translations (
    id SERIAL PRIMARY KEY,
    variant_id INTEGER NOT NULL REFERENCES variants(id),
    name VARCHAR(50) NOT NULL,
    language_code VARCHAR(2) NOT NULL DEFAULT 'en'
);

CREATE TABLE variant_values (
    id SERIAL PRIMARY KEY,
    variant_id INTEGER NOT NULL REFERENCES variants(id),
    value VARCHAR(50) NOT NULL,
    language_code VARCHAR(2) NOT NULL DEFAULT 'en'
);

CREATE TABLE product_options (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id),
    variant_id INTEGER NOT NULL REFERENCES variants(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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
    name VARCHAR(50) NOT NULL,
    language_code VARCHAR(2) NOT NULL DEFAULT 'en'
);

CREATE TABLE product_variant_selected_values (
    id SERIAL PRIMARY KEY,
    product_variant_id INTEGER NOT NULL REFERENCES product_variants(id),
    variant_value_id INTEGER NOT NULL REFERENCES variant_values(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE product_selected_values (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id),
    variant_value_id INTEGER NOT NULL REFERENCES variant_values(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

DROP INDEX IF EXISTS idx_variant_translation_name CASCADE;
DROP INDEX IF EXISTS idx_variant_translation CASCADE;
DROP INDEX IF EXISTS idx_variant_value_translation CASCADE;
DROP INDEX IF EXISTS idx_variant_translation CASCADE;
DROP INDEX IF EXISTS idx_product_variant_option CASCADE;
DROP INDEX IF EXISTS idx_product_variant_sku CASCADE;
DROP INDEX IF EXISTS idx_product_variant_created_at CASCADE;
DROP INDEX IF EXISTS idx_product_variant_is_archived CASCADE;
DROP INDEX IF EXISTS idx_product_variant_is_default CASCADE;
DROP INDEX IF EXISTS idx_product_variant_price CASCADE;
DROP INDEX IF EXISTS idx_product_variant_translation CASCADE;
DROP INDEX IF EXISTS idx_product_variant_selected_value CASCADE;
DROP INDEX IF EXISTS idx_product_selected_value CASCADE;


CREATE UNIQUE INDEX idx_variant_value_translation ON variant_values(value, language_code);
CREATE UNIQUE INDEX idx_variant_translation_name ON variant_translations(name);
CREATE UNIQUE INDEX idx_variant_translation on variant_translations(variant_id, language_code);
CREATE UNIQUE INDEX idx_product_variant_option ON product_options(product_id, variant_id);
CREATE UNIQUE INDEX idx_product_variant_sku ON product_variants(sku);
CREATE INDEX idx_product_variant_created_at ON product_variants(created_at);
CREATE INDEX idx_product_variant_is_archived ON product_variants(is_archived);
CREATE INDEX idx_product_variant_is_default ON product_variants(is_default, product_id);
CREATE INDEX idx_product_variant_price ON product_variants(price);
CREATE UNIQUE INDEX idx_product_variant_translation ON product_variant_translations(product_variant_id, language_code);
CREATE UNIQUE INDEX idx_product_variant_selected_value ON product_variant_selected_values(product_variant_id, variant_value_id);
CREATE UNIQUE INDEX idx_product_selected_value ON product_selected_values(product_id, variant_value_id);

-- END VARIANT TABLES --

-- RECIPE AND BATCHES TABLES --
DROP TABLE IF EXISTS recipes CASCADE;
DROP TABLE IF EXISTS batches CASCADE;

CREATE TABLE recipes (
    id SERIAL PRIMARY KEY,
    product_variant_id INTEGER NOT NULL REFERENCES product_variants(id),
    ingredient_id INTEGER NOT NULL REFERENCES ingredients(id),
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

CREATE UNIQUE INDEX idx_recipe ON recipes(product_variant_id, ingredient_id);
CREATE UNIQUE INDEX idx_batch ON batches(sku, warehouse_id, expires_at);

-- END RECIPE AND BATCHES TABLES --