#!/bin/sh
set -e

# Run seed if DB doesn't exist yet
if [ ! -f /app/data/data.db ]; then
  echo "Database not found, running seed..."
  ./seed -db /app/data/data.db -import /app/import
  echo "Seed complete."
fi

# Start server
exec ./server
