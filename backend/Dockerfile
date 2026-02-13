FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files (go.sum might not exist or be empty)
COPY go.mod ./
COPY go.sum* ./

# Download dependencies first
RUN go mod download

# Copy source code (needed for go mod tidy to see imports)
COPY . .

# Generate go.sum with all dependencies
RUN go mod tidy

# Build the application (use -mod=mod to ignore vendor directory if it exists)
RUN CGO_ENABLED=0 GOOS=linux go build -mod=mod -o /app/bin/server ./cmd/server

FROM golang:1.21-alpine AS init-builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod ./
COPY go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code needed for init-admin (scripts and internal/database)
COPY scripts/init-admin.go ./scripts/
COPY internal/ ./internal/

# Build init script
RUN CGO_ENABLED=0 GOOS=linux go build -mod=mod -o /app/bin/init-admin ./scripts/init-admin.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates postgresql-client

WORKDIR /app

# Create upload directory
RUN mkdir -p /tmp/edda-uploads && chmod 755 /tmp/edda-uploads

# Copy server binary
COPY --from=builder /app/bin/server .

# Copy init script
COPY --from=init-builder /app/bin/init-admin .

# Copy wait script (strip CR so it works when repo was cloned on Windows with CRLF)
COPY scripts/wait-for-db.sh /app/wait-for-db.sh
RUN sed -i 's/\r$//' /app/wait-for-db.sh && chmod +x /app/wait-for-db.sh

EXPOSE 8080

# Wait for DB, run init script, then start server
CMD ["/bin/sh", "-c", "/app/wait-for-db.sh && sleep 2 && /app/init-admin && exec /app/server"]
