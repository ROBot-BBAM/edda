#!/bin/bash

# Script to create the first admin user
# Usage: ./scripts/create-admin.sh <email> <password>

if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <email> <password>"
    echo "Example: $0 admin@example.com admin123"
    exit 1
fi

EMAIL=$1
PASSWORD=$2

# Check if Python is available for bcrypt hashing
if command -v python3 &> /dev/null; then
    echo "Generating bcrypt hash..."
    HASH=$(python3 -c "import bcrypt; print(bcrypt.hashpw(b'$PASSWORD', bcrypt.gensalt()).decode())")
else
    echo "Python3 not found. Please generate a bcrypt hash manually:"
    echo "1. Go to https://bcrypt-generator.com/"
    echo "2. Enter your password: $PASSWORD"
    echo "3. Copy the generated hash"
    echo ""
    read -p "Paste the bcrypt hash here: " HASH
fi

echo "Creating admin user: $EMAIL"
docker-compose exec -T db psql -U edda -d edda <<EOF
INSERT INTO users (id, email, password_hash, is_admin, created_at, updated_at)
VALUES (
  gen_random_uuid(),
  '$EMAIL',
  '$HASH',
  true,
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP
);
EOF

if [ $? -eq 0 ]; then
    echo "Admin user created successfully!"
    echo "You can now log in at http://localhost:3000/login"
else
    echo "Failed to create admin user. Make sure docker-compose is running."
fi
