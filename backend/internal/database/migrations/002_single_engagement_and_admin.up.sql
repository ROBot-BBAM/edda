-- Add is_admin field to users table
ALTER TABLE users ADD COLUMN is_admin BOOLEAN DEFAULT FALSE NOT NULL;

CREATE INDEX idx_users_is_admin ON users(is_admin);

-- Make project_id nullable in scan_files (we'll remove it later, but need to handle existing data)
ALTER TABLE scan_files ALTER COLUMN project_id DROP NOT NULL;

-- Make project_id nullable in hosts and update unique constraint
ALTER TABLE hosts ALTER COLUMN project_id DROP NOT NULL;
ALTER TABLE hosts DROP CONSTRAINT IF EXISTS hosts_project_id_ip_address_key;
ALTER TABLE hosts ADD CONSTRAINT hosts_ip_address_key UNIQUE (ip_address);

-- Make project_id nullable in urls
ALTER TABLE urls ALTER COLUMN project_id DROP NOT NULL;

-- Note: We keep the project_id columns for now to avoid breaking existing data
-- In a fresh deployment, these columns can be removed entirely
