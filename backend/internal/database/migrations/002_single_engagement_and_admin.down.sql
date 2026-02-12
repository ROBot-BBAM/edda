-- Revert unique constraint on hosts
ALTER TABLE hosts DROP CONSTRAINT IF EXISTS hosts_ip_address_key;
ALTER TABLE hosts ADD CONSTRAINT hosts_project_id_ip_address_key UNIQUE (project_id, ip_address);

-- Make project_id NOT NULL again
ALTER TABLE hosts ALTER COLUMN project_id SET NOT NULL;
ALTER TABLE urls ALTER COLUMN project_id SET NOT NULL;
ALTER TABLE scan_files ALTER COLUMN project_id SET NOT NULL;

-- Remove is_admin field
DROP INDEX IF EXISTS idx_users_is_admin;
ALTER TABLE users DROP COLUMN IF EXISTS is_admin;
