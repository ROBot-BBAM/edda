DROP TRIGGER IF EXISTS update_urls_updated_at ON urls;
DROP TRIGGER IF EXISTS update_ports_updated_at ON ports;
DROP TRIGGER IF EXISTS update_hosts_updated_at ON hosts;
DROP TRIGGER IF EXISTS update_projects_updated_at ON projects;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

DROP FUNCTION IF EXISTS update_updated_at_column();

DROP TABLE IF EXISTS urls;
DROP TABLE IF EXISTS ports;
DROP TABLE IF EXISTS hosts;
DROP TABLE IF EXISTS scan_files;
DROP TABLE IF EXISTS project_members;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS users;
