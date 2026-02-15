#!/bin/bash

set -e

echo "=== GoArchive - Quick Start Script ==="
echo ""

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "Error: Docker is not installed. Please install Docker first."
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo "Error: Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

echo "✓ Docker and Docker Compose are installed"
echo ""

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo "Creating .env file from .env.example..."
    cp .env.example .env
    echo "✓ .env file created. Please edit it with your configuration."
    echo ""
fi

# Start services
echo "Starting PostgreSQL and LocalStack..."
docker-compose up -d postgres localstack

echo "Waiting for services to be ready (15 seconds)..."
sleep 15

# Create S3 bucket
echo "Creating S3 bucket in LocalStack..."
docker-compose run --rm -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test goarchive sh -c "
    apk add --no-cache aws-cli > /dev/null 2>&1
    aws --endpoint-url=http://localstack:4566 s3 mb s3://backups --region us-east-1 2>/dev/null || echo 'Bucket already exists'
" 2>/dev/null

echo "✓ LocalStack S3 bucket ready"
echo ""

# Build the goarchive image
echo "Building goarchive Docker image..."
docker-compose build goarchive

echo ""
echo "=== Setup Complete! ==="
echo ""
echo "To run a backup, execute:"
echo "  docker-compose run --rm goarchive"
echo ""
echo "To view logs:"
echo "  docker-compose logs -f"
echo ""
echo "To stop services:"
echo "  docker-compose down"
echo ""
echo "To connect to the test database:"
echo "  psql -h localhost -U testuser -d testdb"
echo "  Password: testpass"
echo ""
