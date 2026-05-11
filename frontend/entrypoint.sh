#!/bin/sh
set -e

# Create nginx PID directory with correct permissions
mkdir -p /var/run/nginx
chmod 755 /var/run/nginx

# Start nginx in the foreground
exec nginx -g "daemon off;"
