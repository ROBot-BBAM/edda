-- Clear all data from tables (keeps schema intact)
-- Run with: psql postgres://edda:edda_dev_password@localhost:5432/edda -f clear_db.sql

-- Disable foreign key checks temporarily by truncating in order
-- Note: schema_migrations is kept so migrations don't re-run

TRUNCATE TABLE urls CASCADE;
TRUNCATE TABLE ports CASCADE;
TRUNCATE TABLE hosts CASCADE;
TRUNCATE TABLE scan_files CASCADE;
TRUNCATE TABLE project_members CASCADE;
TRUNCATE TABLE projects CASCADE;
TRUNCATE TABLE users CASCADE;

-- Reset sequences if any (though UUIDs don't use sequences)
-- This script clears all data but keeps the schema and migrations history
