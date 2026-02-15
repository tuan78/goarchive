# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o goarchive ./cmd/goarchive

# Runtime stage
FROM postgres:16-alpine

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/goarchive .

# Set environment variables (will be overridden by runtime config)
ENV DB_TYPE=postgres \
    DB_HOST=localhost \
    DB_PORT=5432 \
    DB_USERNAME=postgres \
    DB_DATABASE=postgres \
    DB_SSLMODE=disable \
    STORAGE_TYPE=s3 \
    STORAGE_REGION=us-east-1

# Run the backup application
ENTRYPOINT ["./goarchive"]
