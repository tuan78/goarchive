# Docker Compose Deployment Guide

This guide explains how to deploy GoArchive using Docker Compose with different scheduling options.

## Available Services

### Infrastructure Services

- **postgres**: PostgreSQL database for testing
- **localstack**: Local AWS S3 emulator for testing

### Backup Services

#### Scheduled Backups (Recommended for Production)

Run backups continuously on a schedule:

- **goarchive-scheduled**: Scheduled disk-based backups (default service)
- **goarchive-scheduled-s3**: Scheduled S3-based backups (requires profile)

#### One-time Backups

Run a single backup and exit:

- **goarchive**: One-time S3 backup
- **goarchive-disk**: One-time disk backup

## Quick Start

### Option 1: Scheduled Backups (Default)

Run continuous scheduled backups to disk:

```bash
# Start infrastructure and scheduled backup service
docker-compose up -d

# View logs
docker-compose logs -f goarchive-scheduled

# Stop services
docker-compose down
```

The default schedule is **every 6 hours**. Backups are stored in the `backup-data` Docker volume.

### Option 2: Scheduled Backups to S3

```bash
# Start with S3 scheduled backups
docker-compose --profile scheduled up -d

# View logs
docker-compose logs -f goarchive-scheduled-s3
```

### Option 3: One-time Backup

```bash
# Run a single backup and exit
docker-compose --profile oneshot run --rm goarchive-disk

# Or with S3
docker-compose --profile oneshot run --rm goarchive
```

## Scheduling Configuration

You can customize the backup schedule using environment variables:

### Cron Expression (Default)

```yaml
environment:
  BACKUP_SCHEDULE: "0 */6 * * *" # Every 6 hours
```

Common cron expressions:

- `"0 * * * *"` - Every hour
- `"0 */4 * * *"` - Every 4 hours
- `"0 2 * * *"` - Daily at 2 AM
- `"0 2 * * 0"` - Weekly on Sunday at 2 AM
- `"0 0 1 * *"` - Monthly on the 1st at midnight

### Interval-based (Alternative)

```yaml
environment:
  BACKUP_INTERVAL: "3600" # Every 1 hour (in seconds)
```

This is simpler but less flexible than cron.

## Production Deployment

### Step 1: Create docker-compose.override.yml

Create a file `docker-compose.override.yml` in your project directory:

```yaml
version: "3.8"

services:
  goarchive-scheduled:
    environment:
      # Your PostgreSQL database
      DB_HOST: your-db-host.example.com
      DB_PORT: 5432
      DB_USERNAME: ${DB_USERNAME}
      DB_PASSWORD: ${DB_PASSWORD}
      DB_DATABASE: production_db
      DB_SSLMODE: require

      # Disk storage (local backups)
      STORAGE_TYPE: disk
      STORAGE_PATH: /root/backups

      # Schedule: Daily at 2 AM
      BACKUP_SCHEDULE: "0 2 * * *"

    volumes:
      # Mount to host directory for easy access
      - ./backups:/root/backups
```

### Step 2: Create .env file

```bash
DB_USERNAME=your_db_user
DB_PASSWORD=your_secure_password
```

### Step 3: Deploy

```bash
# Start the scheduled backup service
docker-compose up -d goarchive-scheduled postgres

# Verify it's running
docker-compose ps

# Check logs
docker-compose logs -f goarchive-scheduled
```

## S3 Production Deployment

For S3 backups, configure AWS credentials:

```yaml
version: "3.8"

services:
  goarchive-scheduled-s3:
    environment:
      # Database configuration
      DB_HOST: ${DB_HOST}
      DB_PORT: 5432
      DB_USERNAME: ${DB_USERNAME}
      DB_PASSWORD: ${DB_PASSWORD}
      DB_DATABASE: ${DB_DATABASE}
      DB_SSLMODE: require

      # S3 storage
      STORAGE_TYPE: s3
      STORAGE_BUCKET: ${S3_BUCKET}
      STORAGE_REGION: ${AWS_REGION}
      STORAGE_ACCESS_KEY: ${AWS_ACCESS_KEY_ID}
      STORAGE_SECRET_KEY: ${AWS_SECRET_ACCESS_KEY}
      STORAGE_PREFIX: production-backups/

      # Schedule: Every 6 hours
      BACKUP_SCHEDULE: "0 */6 * * *"
```

Then `.env`:

```bash
DB_HOST=your-db-host.example.com
DB_USERNAME=your_db_user
DB_PASSWORD=your_secure_password
DB_DATABASE=production_db

S3_BUCKET=my-backup-bucket
AWS_REGION=us-west-2
AWS_ACCESS_KEY_ID=AKIA...
AWS_SECRET_ACCESS_KEY=...
```

Deploy:

```bash
docker-compose --profile scheduled up -d
```

## Monitoring

### View Logs

```bash
# Follow logs
docker-compose logs -f goarchive-scheduled

# View last 100 lines
docker-compose logs --tail=100 goarchive-scheduled
```

### Check Backups

For disk storage:

```bash
# List backups in volume
docker-compose exec goarchive-scheduled ls -lh /root/backups

# On host (if mounted)
ls -lh ./backups/
```

For S3 storage:

```bash
# List backups
docker-compose exec goarchive-scheduled ./goarchive list
```

### Manual Backup Trigger

Run a backup immediately without waiting for schedule:

```bash
docker-compose exec goarchive-scheduled ./goarchive backup
```

## Backup Retention

The scheduler doesn't automatically delete old backups. For retention management:

### Option 1: External Cleanup Script

Create a weekly cleanup cron on your host:

```bash
# Delete backups older than 30 days
find ./backups -name "*.sql.gz" -mtime +30 -delete
```

### Option 2: S3 Lifecycle Policies

Configure S3 lifecycle rules in AWS Console to automatically expire old backups.

### Option 3: Add to Scheduler (Future Feature)

Track issue #XXX for built-in retention policy support.

## Restart Policy

The scheduled services use `restart: unless-stopped`, meaning:

- Automatically restart on failure
- Restart after system reboot
- Don't restart if manually stopped

To stop permanently:

```bash
docker-compose down
```

## Troubleshooting

### Service keeps restarting

```bash
# Check logs for errors
docker-compose logs goarchive-scheduled

# Common issues:
# - Database connection failed
# - Invalid credentials
# - Storage bucket doesn't exist
```

### Schedule not working

```bash
# Check cron is running
docker-compose exec goarchive-scheduled ps aux | grep crond

# Verify crontab
docker-compose exec goarchive-scheduled crontab -l

# Check cron logs
docker-compose exec goarchive-scheduled cat /var/log/goarchive-cron.log
```

### Backups not appearing

```bash
# Verify storage configuration
docker-compose exec goarchive-scheduled env | grep STORAGE

# Test manual backup
docker-compose exec goarchive-scheduled ./goarchive backup -v
```

## Migration from One-time to Scheduled

If you were using one-time backups:

1. Stop old service: `docker-compose stop goarchive-disk`
2. Update to scheduled: `docker-compose up -d goarchive-scheduled`
3. Existing backups in `backup-data` volume are preserved

## See Also

- [Main README](../README.md) - Project overview
- [Dockerfile.scheduler](../Dockerfile.scheduler) - Scheduler image details
- [Kubernetes Deployment](./KUBERNETES.md) - For Kubernetes deployments
