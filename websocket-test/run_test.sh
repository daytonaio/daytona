#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}==========================================${NC}"
echo -e "${BLUE}WebSocket Close Code 4008 Test${NC}"
echo -e "${BLUE}==========================================${NC}"
echo ""

# Start the Go server in background
echo -e "${YELLOW}Starting Go WebSocket server...${NC}"
cd /workspaces/daytona/websocket-test
go run server.go &
SERVER_PID=$!

# Wait for server to start
sleep 2

# Run the Python client
echo -e "${YELLOW}Running Python client...${NC}"
echo ""
python3 client.py

# Cleanup: kill the server
echo ""
echo -e "${YELLOW}Cleaning up...${NC}"
kill $SERVER_PID 2>/dev/null

echo ""
echo -e "${GREEN}Test complete!${NC}"

