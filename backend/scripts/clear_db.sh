#!/bin/bash
# Clear database script for demo purposes
# Usage: ./clear_db.sh

set -e

DB_URL="${DATABASE_URL:-postgres://edda:edda_dev_password@localhost:5432/edda?sslmode=disable}"

echo "Clearing database..."
psql "$DB_URL" -f "$(dirname "$0")/clear_db.sql"
echo "Database cleared successfully!"
