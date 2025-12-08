#!/bin/bash
# Telepresence Setup for Daytona Runner Access
# This script connects to the runner DaemonSet on a specific node

set -e

# Configuration
KUBECONFIG="${KUBECONFIG:-/workspaces/daytona/.tmp/kubeconfig.yaml}"
NAMESPACE="daytona-dev"
NODE_NAME="gke-vedran-gke-test--ubuntu-sandbox-p-e4be3ca9-dq1g"
LOCAL_PORT="8080"
RUNNER_API_PORT="8080"
SSH_GATEWAY_PORT="2220"
LOCAL_SSH_PORT="2220"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== Daytona Runner Telepresence Setup ===${NC}"
echo ""

# Check if telepresence is installed
if ! command -v telepresence &> /dev/null; then
    echo -e "${RED}✗ Telepresence is not installed${NC}"
    echo ""
    echo "Install telepresence:"
    echo "  macOS: brew install datawire/blackbird/telepresence"
    echo "  Linux: curl -fL https://app.getambassador.io/download/tel2oss/releases/download/v2.20.2/telepresence-linux-amd64 -o /usr/local/bin/telepresence && chmod +x /usr/local/bin/telepresence"
    exit 1
fi

echo -e "${GREEN}✓ Telepresence is installed${NC}"
echo ""

# Get the specific runner pod on the target node
echo "Finding runner pod on node: $NODE_NAME"
POD_NAME=$(kubectl --kubeconfig="$KUBECONFIG" get pod -n "$NAMESPACE" -l app=daytona-runner -o json | \
  jq -r ".items[] | select(.spec.nodeName==\"$NODE_NAME\") | .metadata.name")

if [ -z "$POD_NAME" ]; then
    echo -e "${RED}✗ No runner pod found on node $NODE_NAME${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Found pod: $POD_NAME${NC}"
echo ""

# Check if telepresence is already connected
if telepresence status 2>&1 | grep -q "Connected"; then
    echo -e "${YELLOW}⚠ Telepresence already connected, disconnecting first...${NC}"
    telepresence quit
    sleep 2
fi

# Method 1: Direct Port Forward (Simplest)
echo -e "${GREEN}=== Method 1: Direct Port Forward (Recommended) ===${NC}"
echo ""
echo "This will forward the specific runner pod to localhost:$LOCAL_PORT"
echo ""
echo "Run this command:"
echo ""
echo -e "${YELLOW}kubectl --kubeconfig=\"$KUBECONFIG\" port-forward -n $NAMESPACE pod/$POD_NAME $LOCAL_PORT:$RUNNER_API_PORT $LOCAL_SSH_PORT:$SSH_GATEWAY_PORT${NC}"
echo ""
echo "Then access:"
echo "  Runner API: http://localhost:$LOCAL_PORT"
echo "  SSH Gateway: localhost:$LOCAL_SSH_PORT"
echo ""
echo "Press ENTER to start port-forward or CTRL+C to see other methods..."
read -r

echo -e "${GREEN}Starting port-forward...${NC}"
kubectl --kubeconfig="$KUBECONFIG" port-forward -n "$NAMESPACE" "pod/$POD_NAME" "$LOCAL_PORT:$RUNNER_API_PORT" "$LOCAL_SSH_PORT:$SSH_GATEWAY_PORT"
