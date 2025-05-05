#!/bin/sh
set -e

# Start MinIO in background
minio server /data --console-address ":9001" &

# Wait for MinIO to start
sleep 5

# Create the daytona bucket and apply lifecycle policy
mkdir -p /tmp/mc
cd /tmp/mc
wget -q https://dl.min.io/client/mc/release/linux-amd64/mc
chmod +x mc
./mc alias set myminio http://localhost:9000 minioadmin minioadmin
./mc mb myminio/daytona --ignore-existing
./mc ilm import myminio/daytona < /etc/minio/lifecycle/lifecycle.json

# Keep container running
wait