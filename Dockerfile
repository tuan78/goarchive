# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /app

# Copy everything (submodules need local replace paths to work)
COPY . .

# Download dependencies for the CLI
WORKDIR /app/cmd/goarchive
RUN go mod download

# Build the application from cmd/goarchive
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o goarchive .

# Runtime stage
FROM postgres:18-alpine

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/cmd/goarchive/goarchive .

# Create default backup directory for disk storage
RUN mkdir -p /root/backups

# Volume for persistent disk storage
VOLUME ["/root/backups"]

# Run the backup application
# Environment variables should be provided at runtime:
# - Database: DB_HOST, DB_PORT, DB_USERNAME, DB_PASSWORD, DB_DATABASE, DB_TYPE, DB_SSLMODE
# - Storage: STORAGE_TYPE (disk|s3), STORAGE_PATH, STORAGE_BUCKET, STORAGE_REGION, etc.
ENTRYPOINT ["./goarchive"]
CMD ["backup"]
