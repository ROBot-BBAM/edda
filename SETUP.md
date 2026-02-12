# Edda Setup Guide

## Changes Made

This application has been updated to support a **single engagement model** where:

1. **Single Engagement**: All data is shared across all users - no project isolation
2. **Manual User Registration**: Public registration is disabled. Only admins can create users
3. **Shared Data**: All users see all hosts, ports, URLs, and scan files uploaded by any user

## Automatic Admin User Creation

**The admin user is now created automatically when you start the application!**

When you run `docker-compose up`, the backend will:
1. Wait for the database to be ready
2. Check if an admin user exists
3. If no admin exists, create one with a randomly generated password
4. Display the credentials in the logs

**Look for output like this in your docker-compose logs:**
```
================================================================================
=                                                                              =
=                    EDDA ADMIN CREDENTIALS                                    =
=                                                                              =
================================================================================
Email:    admin@edda.local
Password: AbC123!@#XyZ789
================================================================================
⚠️  IMPORTANT: Save these credentials! They will not be shown again.
================================================================================
```

You can also check the logs with:
```bash
docker-compose logs backend | grep -A 10 "ADMIN CREDENTIALS"
```

Or view all backend logs:
```bash
docker-compose logs backend
```

## Customizing Admin Email

You can set a custom admin email using the `ADMIN_EMAIL` environment variable:

```bash
ADMIN_EMAIL=your-admin@example.com docker-compose up
```

Or add it to your `.env` file:
```
ADMIN_EMAIL=your-admin@example.com
```

## Manual Admin User Creation (Alternative)

If you prefer to create the admin user manually, here are the options:

### Option 1: Using Docker exec (Recommended)

1. Start the services:
   ```bash
   docker-compose up -d
   ```

2. Connect to the PostgreSQL container:
   ```bash
   docker-compose exec db psql -U edda -d edda
   ```

3. Create an admin user (replace email and password):
   ```sql
   -- Generate a bcrypt hash for your password first
   -- You can use: https://bcrypt-generator.com/ or any bcrypt tool
   -- For password "admin123", the hash might be: $2a$10$...
   
   INSERT INTO users (id, email, password_hash, is_admin, created_at, updated_at)
   VALUES (
     gen_random_uuid(),
     'admin@example.com',
     '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', -- Replace with your bcrypt hash
     true,
     CURRENT_TIMESTAMP,
     CURRENT_TIMESTAMP
   );
   ```

### Option 2: Using a temporary registration endpoint

Alternatively, you can temporarily enable registration by modifying `backend/internal/handlers/auth.go`:

1. Comment out the `Register` function's error return
2. Uncomment the original registration logic
3. Create your admin user
4. Update the user in the database to set `is_admin = true`
5. Revert the changes

### Option 3: Using Go script

Create a simple Go script to create the first admin user:

```go
package main

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    
    _ "github.com/lib/pq"
    "golang.org/x/crypto/bcrypt"
)

func main() {
    dbURL := os.Getenv("DATABASE_URL")
    if dbURL == "" {
        dbURL = "postgres://edda:edda_dev_password@localhost:5432/edda?sslmode=disable"
    }
    
    db, err := sql.Open("postgres", dbURL)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    email := "admin@example.com"
    password := "admin123"
    
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        log.Fatal(err)
    }
    
    _, err = db.Exec(`
        INSERT INTO users (id, email, password_hash, is_admin, created_at, updated_at)
        VALUES (gen_random_uuid(), $1, $2, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
    `, email, string(hash))
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Admin user created: %s\n", email)
}
```

## After Creating Admin User

1. Log in at http://localhost:3000/login with your admin credentials
2. Click the "Admin" button in the header to access user management
3. Create additional users as needed through the admin interface

## Database Migrations

The application includes two migrations:
- `001_initial_schema.up.sql` - Initial schema with projects, users, hosts, ports, URLs
- `002_single_engagement_and_admin.up.sql` - Adds `is_admin` field and makes `project_id` nullable

Migrations run automatically on server startup.

## Notes

- All users can see and modify all data (shared engagement model)
- Only admins can create new users
- Project concept is kept for backward compatibility but all users see all projects
- Future updates will remove project_id entirely from hosts, ports, and URLs tables
