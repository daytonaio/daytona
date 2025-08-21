#!/bin/bash

# Simple DNS fix for Daytona - just add to /etc/hosts
echo "Adding sandbox hostnames to /etc/hosts..."

# Get running sandbox containers
SANDBOXES=$(docker ps --format "{{.Names}}" --filter "name=^[a-f0-9-]+$" 2>/dev/null || echo "")

if [ -z "$SANDBOXES" ]; then
    echo "No running sandbox containers found."
    echo "Create a sandbox first, then run this script again."
    exit 0
fi

# Remove old Daytona entries
sudo sed -i '/# DAYTONA_SANDBOXES_START/,/# DAYTONA_SANDBOXES_END/d' /etc/hosts

# Add new entries
echo "# DAYTONA_SANDBOXES_START" | sudo tee -a /etc/hosts
for sandbox in $SANDBOXES; do
    echo "127.0.0.1 2280-$sandbox.proxy.localhost" | sudo tee -a /etc/hosts
done
echo "# DAYTONA_SANDBOXES_END" | sudo tee -a /etc/hosts

echo "Done! Added $(echo "$SANDBOXES" | wc -l) sandbox hostnames to /etc/hosts"
echo "Now try running your Daytona examples!"
