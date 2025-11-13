#!/bin/sh

set -e

# Replace %DAYTONA_BASE_API_URL% with actual environment variable value
find /usr/share/nginx/html -type f \( -name "*.js" -o -name "*.html" \) -exec sed -i "s|%DAYTONA_BASE_API_URL%|${DAYTONA_BASE_API_URL}|g" {} +

# Start nginx
exec nginx -g "daemon off;"
