#!/bin/bash
# Initialize PostgreSQL database with schema

set -e

echo "🔄 Initializing database..."

docker exec pulzifi-postgres psql -U pulzifi_user -d pulzifi -f /dev/stdin < database_setup.sql

echo "✅ Database initialized successfully"
