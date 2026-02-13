# Edda - Pentest Data Management Platform

Edda is a web application for managing and tracking pentest scan data. It allows users to upload scan files (nmap XML, ffuf JSON/CSV), parse the data, and track review progress for hosts, ports, services, and URLs.

## Features

- **User Authentication**: Secure registration and login with JWT tokens
- **Project Management**: Create and manage multiple pentest projects
- **Scan File Upload**: Upload and parse nmap XML and ffuf JSON/CSV files (coming soon)
- **Review Tracking**: Mark hosts, ports, services, and URLs as reviewed
- **Progress Tracking**: Filter and view reviewed/unreviewed items

## Tech Stack

- **Frontend**: React 18 + TypeScript
- **Backend**: Go 1.21+ with Chi router
- **Database**: PostgreSQL 16
- **Authentication**: JWT tokens with bcrypt password hashing

## Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)
- Node.js 18+ and npm (for local frontend development)

## Quick Start with Docker Compose

1. **Clone the repository** (if you haven't already):
   ```bash
   git clone <repository-url>
   cd edda
   ```

2. **Start all services**:
   ```bash
   docker-compose up
   ```

   This will start:
   - PostgreSQL database on port 5432
   - Go backend API on port 8080
   - React frontend on port 3000

3. **Access the application**:
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080
   - Health check: http://localhost:8080/healthz

4. **Create your first account**:
   - Check the logs for you admin password
   ```bash
   docker logs edda-backend
   ```
   - Login with the admin email and password
   - Create desired users

## Local Development Setup

### Backend Setup

1. **Install dependencies**:
   ```bash
   cd backend
   go mod download
   ```

2. **Set up environment variables**:
   Create a `.env` file or export:
   ```bash
   export DATABASE_URL="postgres://edda:edda_dev_password@localhost:5432/edda?sslmode=disable"
   export JWT_SECRET="your-secret-key-change-in-production"
   export PORT=8080
   ```

3. **Start PostgreSQL** (if not using Docker):
   ```bash
   docker-compose up db
   ```

4. **Run the backend**:
   ```bash
   go run cmd/server/main.go
   ```

### Frontend Setup

1. **Install dependencies**:
   ```bash
   cd frontend
   npm install
   ```

2. **Set up environment variables**:
   Create a `.env` file:
   ```
   REACT_APP_API_URL=http://localhost:8080/api
   ```

3. **Run the frontend**:
   ```bash
   npm start
   ```

   The frontend will be available at http://localhost:3000

## Project Structure

```
edda/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go          # Application entry point
│   ├── internal/
│   │   ├── database/
│   │   │   ├── migrations/      # SQL migration files
│   │   │   ├── database.go      # Database connection and migrations
│   │   │   ├── users.go         # User database operations
│   │   │   └── projects.go      # Project database operations
│   │   ├── handlers/
│   │   │   ├── auth.go          # Authentication handlers
│   │   │   ├── projects.go      # Project handlers
│   │   │   └── handlers.go      # Handler initialization
│   │   └── middleware/
│   │       └── auth/
│   │           └── auth.go      # JWT authentication middleware
│   ├── go.mod
│   └── Dockerfile
├── frontend/
│   ├── public/
│   ├── src/
│   │   ├── components/          # Reusable React components
│   │   ├── contexts/            # React contexts (AuthContext)
│   │   ├── pages/               # Page components
│   │   ├── services/            # API service layer
│   │   ├── App.tsx
│   │   └── index.tsx
│   ├── package.json
│   └── Dockerfile
├── docker-compose.yml
└── README.md
```

## API Endpoints

### Authentication
- `POST /api/login` - Login and get JWT token
- `GET /api/me` - Get current user info (protected)

### Projects
- `GET /api/projects` - List all projects for the current user (protected)
- `POST /api/projects` - Create a new project (protected)
- `GET /api/projects/{id}` - Get project details (protected)
- `PATCH /api/projects/{id}` - Update project (protected)
- `DELETE /api/projects/{id}` - Delete project (protected)

## Database Schema

The application uses the following main tables:
- `users` - User accounts
- `projects` - Pentest projects/engagements
- `project_members` - Project membership (for multi-user projects)
- `scan_files` - Uploaded scan files metadata
- `hosts` - Discovered hosts with review tracking
- `ports` - Open ports with review tracking
- `urls` - Discovered URLs from web scans with review tracking

See `backend/internal/database/migrations/001_initial_schema.up.sql` for the full schema.

## Environment Variables

### Backend
- `DATABASE_URL` - PostgreSQL connection string
- `JWT_SECRET` - Secret key for JWT token signing
- `PORT` - Server port (default: 8080)

### Frontend
- `REACT_APP_API_URL` - Backend API URL (default: `/api`)

## Development Roadmap

- [x] Implement nmap XML parser
- [x] Implement ffuf JSON/CSV parser
- [x] File upload endpoint and UI
- [x] Hosts list view with review toggles
- [x] Ports list view with review toggles
- [x] URLs list view with review toggles
- [x] Filtering and search functionality
- [x] Project statistics dashboard
- [x] Export functionality

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## License

MIT License
