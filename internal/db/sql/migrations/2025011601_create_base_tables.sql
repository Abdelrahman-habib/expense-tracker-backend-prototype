-- +goose Up
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TYPE "projects_status" AS ENUM (
  'ongoing',
  'completed',
  'canceled'
);

CREATE TABLE "users" (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clerk_ex_user_id VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    address_line1 VARCHAR(255),
    address_line2 VARCHAR(255),
    country VARCHAR(100),
    city VARCHAR(100),
    state_province VARCHAR(100),
    zip_postal_code VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE "users_settings" (
  user_settings_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  default_currency CHAR(3) NOT NULL DEFAULT 'USD',
  default_country VARCHAR(100),
  timezone VARCHAR(50),
  date_format VARCHAR(20),
  number_format VARCHAR(20),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX "users_settings_user_id_idx" ON users_settings (user_id);

CREATE TABLE "projects" (
    project_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    status projects_status NOT NULL DEFAULT 'ongoing',
    start_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    end_date TIMESTAMP,
    budget DECIMAL(10,2),
    actual_cost DECIMAL(10,2) DEFAULT 0,
    address_line1 VARCHAR(255),
    address_line2 VARCHAR(255),
    country VARCHAR(100),
    city VARCHAR(100),
    state_province VARCHAR(100),
    zip_postal_code VARCHAR(20),
    website VARCHAR(255),
    tags UUID[],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX projects_user_id_name_idx ON projects(user_id, name);
CREATE INDEX projects_status_idx ON projects(status);
CREATE INDEX projects_tags_gin_idx ON projects USING gin(tags);


CREATE TABLE "wallets" (
    wallet_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    project_id UUID,
    name VARCHAR(100) NOT NULL,
    balance DECIMAL(10,2) DEFAULT 0,
    currency CHAR(3) NOT NULL DEFAULT 'USD',
    tags UUID[],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (project_id) REFERENCES projects(project_id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX wallets_user_id_name_idx ON wallets(user_id, name);
CREATE INDEX wallets_project_id_idx ON wallets(project_id);
CREATE INDEX wallets_tags_gin_idx ON wallets USING gin(tags);


CREATE TABLE "contacts" (
    contact_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    name VARCHAR(100) NOT NULL,
    phone VARCHAR(20),
    email VARCHAR(100),
    address_line1 VARCHAR(255),
    address_line2 VARCHAR(255),
    country VARCHAR(100),
    city VARCHAR(100),
    state_province VARCHAR(100),
    zip_postal_code VARCHAR(20),
    tags UUID[],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);
CREATE INDEX contacts_user_id_idx ON contacts(user_id);
CREATE INDEX contacts_tags_gin_idx ON contacts USING gin(tags);

CREATE TABLE "tags" (
  tag_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  name VARCHAR(100) NOT NULL,
  color VARCHAR(7),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);
CREATE INDEX tags_user_id_idx ON tags(user_id);

ALTER TABLE tags
ADD CONSTRAINT color_hex_constraint
CHECK (color is null or color ~* '^#[a-f0-9]{6}$');


-- +goose Down
ALTER TABLE tags DROP CONSTRAINT color_hex_constraint;
DROP INDEX IF EXISTS tags_user_id_idx;
DROP TABLE IF EXISTS tags;
DROP INDEX IF EXISTS contacts_tags_gin_idx;
DROP INDEX IF EXISTS contacts_user_id_idx;
DROP TABLE IF EXISTS contacts;
DROP INDEX IF EXISTS wallets_tags_gin_idx;
DROP INDEX IF EXISTS wallets_project_id_idx;
DROP INDEX IF EXISTS wallets_user_id_name_idx;
DROP TABLE IF EXISTS wallets;
DROP INDEX IF EXISTS projects_tags_gin_idx;
DROP INDEX IF EXISTS projects_status_idx;
DROP INDEX IF EXISTS projects_user_id_name_idx;
DROP TABLE IF EXISTS projects;
DROP INDEX IF EXISTS users_settings_user_id_idx;
DROP TABLE IF EXISTS users_settings;
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS projects_status;
DROP EXTENSION IF EXISTS pg_trgm; 






