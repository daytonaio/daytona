#!/bin/bash

# Exit on error
set -e

# Color codes for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Find all example main.go files
examples=$(find ./ -name "main.go" -type f | sort)

echo -e "${BLUE}Running all examples with DAYTONA_API_KEY${NC}"
echo "================================================"
echo ""

# Counter for tracking
total=$(echo "$examples" | wc -l | tr -d ' ')
current=0

# Run each example
for example in $examples; do
    current=$((current + 1))
    example_dir=$(dirname "$example")
    example_name=$(basename "$example_dir")

    echo -e "${BLUE}[$current/$total] Running example: ${GREEN}$example_name${NC}"
    echo "Command: go run $example"
    echo "---"

    if go run "$example"; then
        echo -e "${GREEN}✓ $example_name completed successfully${NC}"
    else
        echo -e "${RED}✗ $example_name failed${NC}"
        # Continue running other examples even if one fails
        # Remove 'set -e' behavior for individual examples
    fi

    echo ""
done

echo -e "${GREEN}All examples completed!${NC}"
