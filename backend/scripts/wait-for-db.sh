#!/bin/sh

# Wait for PostgreSQL to be ready
echo "Waiting for database to be ready..."
max_attempts=30
attempt=0

while [ $attempt -lt $max_attempts ]; do
  if pg_isready -h db -U edda -d edda > /dev/null 2>&1; then
    echo "Database is ready!"
    exit 0
  fi
  attempt=$((attempt + 1))
  echo "Database is unavailable (attempt $attempt/$max_attempts) - sleeping..."
  sleep 1
done

echo "Database did not become ready after $max_attempts attempts"
exit 1
