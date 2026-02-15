#!/bin/sh
# GoArchive Backup Scheduler
# Runs backups on a configurable schedule

set -e

# Default schedule: every 6 hours (can be overridden via BACKUP_SCHEDULE env var)
BACKUP_SCHEDULE=${BACKUP_SCHEDULE:-"0 */6 * * *"}
BACKUP_INTERVAL=${BACKUP_INTERVAL:-""}

echo "==================================="
echo "GoArchive Backup Scheduler Started"
echo "==================================="

if [ -n "$BACKUP_INTERVAL" ]; then
    # Interval-based scheduling (e.g., BACKUP_INTERVAL=3600 for every hour)
    echo "Running in interval mode: every ${BACKUP_INTERVAL} seconds"
    
    while true; do
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] Running backup..."
        /root/goarchive backup || echo "Backup failed with exit code $?"
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] Next backup in ${BACKUP_INTERVAL} seconds"
        sleep "$BACKUP_INTERVAL"
    done
else
    # Cron-based scheduling
    echo "Running in cron mode: ${BACKUP_SCHEDULE}"
    
    # Create cron job
    echo "${BACKUP_SCHEDULE} /root/goarchive backup >> /var/log/goarchive-cron.log 2>&1" > /tmp/crontab
    
    # Install crontab
    crontab /tmp/crontab
    
    # Start cron in foreground
    echo "Starting cron daemon..."
    crond -f -l 2
fi
