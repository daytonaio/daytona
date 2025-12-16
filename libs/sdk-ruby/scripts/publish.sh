#!/bin/bash
# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

set -e

# Publish Daytona Ruby SDK to RubyGems
# 
# Usage:
#   ./scripts/publish.sh [version]
#
# If no version is provided, the current version from version.rb will be used.
#
# Prerequisites:
#   - RubyGems API key configured in ~/.gem/credentials
#   - Or set RUBYGEMS_API_KEY environment variable
#
# To set up credentials:
#   mkdir -p ~/.gem
#   echo "---" > ~/.gem/credentials
#   echo ":rubygems_api_key: YOUR_API_KEY" >> ~/.gem/credentials
#   chmod 0600 ~/.gem/credentials

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SDK_DIR="$(dirname "$SCRIPT_DIR")"
API_CLIENT_DIR="$(dirname "$SDK_DIR")/api-client-ruby"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

echo_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

echo_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if credentials are set up
check_credentials() {
    if [ -n "$RUBYGEMS_API_KEY" ]; then
        echo_info "Using RUBYGEMS_API_KEY environment variable"
        mkdir -p ~/.gem
        echo "---" > ~/.gem/credentials
        echo ":rubygems_api_key: $RUBYGEMS_API_KEY" >> ~/.gem/credentials
        chmod 0600 ~/.gem/credentials
    elif [ ! -f ~/.gem/credentials ]; then
        echo_error "RubyGems credentials not found!"
        echo "Please set up credentials by running:"
        echo "  mkdir -p ~/.gem"
        echo "  echo '---' > ~/.gem/credentials"
        echo "  echo ':rubygems_api_key: YOUR_API_KEY' >> ~/.gem/credentials"
        echo "  chmod 0600 ~/.gem/credentials"
        echo ""
        echo "Or set the RUBYGEMS_API_KEY environment variable"
        exit 1
    fi
}

# Update version if provided
update_version() {
    local new_version=$1
    if [ -n "$new_version" ]; then
        echo_info "Updating version to $new_version"
        sed -i "s/VERSION = '[^']*'/VERSION = '$new_version'/" "$SDK_DIR/lib/daytona/sdk/version.rb"
    fi
}

# Get current version
get_version() {
    grep "VERSION = " "$SDK_DIR/lib/daytona/sdk/version.rb" | sed "s/.*VERSION = '\([^']*\)'.*/\1/"
}

# Build and publish API client gem
publish_api_client() {
    echo_info "Building daytona_api_client gem..."
    cd "$API_CLIENT_DIR"
    
    local api_version=$(grep "VERSION = " lib/daytona_api_client/version.rb | sed "s/.*VERSION = '\([^']*\)'.*/\1/")
    
    # Check if this version already exists on RubyGems
    if gem search daytona_api_client --remote --exact | grep -q "$api_version"; then
        echo_info "daytona_api_client version $api_version already published, skipping..."
        return 0
    fi
    
    gem build daytona_api_client.gemspec
    
    echo_info "Publishing daytona_api_client gem..."
    gem push "daytona_api_client-$api_version.gem"
    
    # Clean up
    rm -f "daytona_api_client-$api_version.gem"
    
    echo_info "daytona_api_client $api_version published successfully!"
}

# Build and publish SDK gem
publish_sdk() {
    echo_info "Building daytona-sdk gem..."
    cd "$SDK_DIR"
    
    local version=$(get_version)
    
    # Check if this version already exists on RubyGems
    if gem search daytona-sdk --remote --exact | grep -q "$version"; then
        echo_error "daytona-sdk version $version already exists on RubyGems!"
        echo "Please update the version in lib/daytona/sdk/version.rb"
        exit 1
    fi
    
    gem build daytona-sdk.gemspec
    
    echo_info "Publishing daytona-sdk gem..."
    gem push "daytona-sdk-$version.gem"
    
    # Clean up
    rm -f "daytona-sdk-$version.gem"
    
    echo_info "daytona-sdk $version published successfully!"
}

# Main
main() {
    local new_version=$1
    
    echo_info "Daytona Ruby SDK Publisher"
    echo ""
    
    check_credentials
    
    if [ -n "$new_version" ]; then
        update_version "$new_version"
    fi
    
    local version=$(get_version)
    echo_info "Publishing version: $version"
    echo ""
    
    # Publish API client first (dependency)
    publish_api_client
    
    # Publish SDK
    publish_sdk
    
    echo ""
    echo_info "All gems published successfully!"
    echo ""
    echo "Install with:"
    echo "  gem install daytona-sdk"
    echo ""
    echo "Or add to Gemfile:"
    echo "  gem 'daytona-sdk', '~> $version'"
}

main "$@"

