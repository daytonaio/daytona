#!/bin/bash
# Sysbox + Docker Installation Validation Script
# This script validates that Docker and Sysbox are properly installed and configured

set -e

NAMESPACE="${NAMESPACE:-daytona-dev}"
POD_NAME=""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
success() {
    echo -e "${GREEN}✓${NC} $1"
}

error() {
    echo -e "${RED}✗${NC} $1"
}

warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

info() {
    echo -e "ℹ $1"
}

# Get runner pod name
get_pod_name() {
    info "Finding runner pod in namespace: $NAMESPACE"
    POD_NAME=$(kubectl get pods -n "$NAMESPACE" -l app=daytona-runner -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
    
    if [ -z "$POD_NAME" ]; then
        error "No runner pod found in namespace $NAMESPACE"
        exit 1
    fi
    
    success "Found runner pod: $POD_NAME"
}

# Check if pod is running
check_pod_status() {
    info "Checking pod status..."
    STATUS=$(kubectl get pod -n "$NAMESPACE" "$POD_NAME" -o jsonpath='{.status.phase}')
    
    if [ "$STATUS" = "Running" ]; then
        success "Pod is running"
    else
        error "Pod status is: $STATUS"
        return 1
    fi
}

# Check docker-installer container
check_installer_container() {
    info "Checking docker-installer container..."
    CONTAINER_STATUS=$(kubectl get pod -n "$NAMESPACE" "$POD_NAME" -o jsonpath='{.status.containerStatuses[?(@.name=="docker-installer")].ready}')
    
    if [ "$CONTAINER_STATUS" = "true" ]; then
        success "docker-installer container is ready"
    else
        warning "docker-installer container is not ready (this may be normal during installation)"
    fi
}

# Execute command on host via nsenter
exec_on_host() {
    kubectl exec -n "$NAMESPACE" "$POD_NAME" -c docker-installer -- \
        nsenter -t 1 -m -u -n -i bash -c "$1" 2>/dev/null
}

# Check Docker installation
check_docker() {
    info "Checking Docker installation..."
    
    if exec_on_host "command -v docker >/dev/null 2>&1"; then
        success "Docker is installed"
        
        # Get Docker version
        DOCKER_VERSION=$(exec_on_host "docker version --format '{{.Server.Version}}' 2>/dev/null" || echo "unknown")
        info "Docker version: $DOCKER_VERSION"
        
        # Check if Docker is running
        if exec_on_host "systemctl is-active --quiet docker"; then
            success "Docker service is running"
        else
            error "Docker service is not running"
            return 1
        fi
    else
        error "Docker is not installed"
        return 1
    fi
}

# Check Sysbox installation
check_sysbox() {
    info "Checking Sysbox installation..."
    
    if exec_on_host "[ -f /usr/bin/sysbox-runc ]"; then
        success "Sysbox is installed"
        
        # Check if Sysbox service is running
        if exec_on_host "systemctl is-active --quiet sysbox"; then
            success "Sysbox service is running"
        else
            error "Sysbox service is not running"
            exec_on_host "systemctl status sysbox --no-pager -l" || true
            return 1
        fi
    else
        error "Sysbox is not installed"
        return 1
    fi
}

# Check Docker runtime configuration
check_docker_runtime() {
    info "Checking Docker runtime configuration..."
    
    # Check if daemon.json exists
    if exec_on_host "[ -f /etc/docker/daemon.json ]"; then
        success "Docker daemon.json exists"
        
        # Check if sysbox-runc is configured
        if exec_on_host "grep -q 'sysbox-runc' /etc/docker/daemon.json"; then
            success "Sysbox runtime is configured in Docker"
            
            # Check if it's the default runtime
            if exec_on_host "grep -q '\"default-runtime\".*\"sysbox-runc\"' /etc/docker/daemon.json"; then
                success "Sysbox is set as default runtime"
            else
                warning "Sysbox is configured but not set as default runtime"
            fi
        else
            error "Sysbox runtime is not configured in Docker"
            return 1
        fi
    else
        error "Docker daemon.json does not exist"
        return 1
    fi
}

# Check Docker info
check_docker_info() {
    info "Checking Docker runtime info..."
    
    RUNTIME=$(exec_on_host "docker info 2>/dev/null | grep -i 'default runtime' | awk '{print \$3}'" || echo "unknown")
    
    if [ "$RUNTIME" = "sysbox-runc" ]; then
        success "Docker is using Sysbox runtime: $RUNTIME"
    else
        warning "Docker default runtime: $RUNTIME (expected: sysbox-runc)"
    fi
}

# Check kernel version
check_kernel() {
    info "Checking kernel version..."
    
    KERNEL_VERSION=$(exec_on_host "uname -r")
    info "Kernel version: $KERNEL_VERSION"
    
    # Extract major and minor version
    MAJOR=$(echo "$KERNEL_VERSION" | cut -d. -f1)
    MINOR=$(echo "$KERNEL_VERSION" | cut -d. -f2)
    
    if [ "$MAJOR" -gt 5 ] || ([ "$MAJOR" -eq 5 ] && [ "$MINOR" -ge 5 ]); then
        success "Kernel version is compatible (5.5+)"
    else
        error "Kernel version is too old (need 5.5+, have $KERNEL_VERSION)"
        return 1
    fi
}

# Check disk space
check_disk_space() {
    info "Checking disk space..."
    
    DOCKER_USAGE=$(exec_on_host "du -sh /var/lib/docker 2>/dev/null | cut -f1" || echo "unknown")
    SYSBOX_USAGE=$(exec_on_host "du -sh /var/lib/sysbox 2>/dev/null | cut -f1" || echo "unknown")
    
    info "Docker storage usage: $DOCKER_USAGE"
    info "Sysbox storage usage: $SYSBOX_USAGE"
    
    # Check available space
    AVAILABLE=$(exec_on_host "df -h /var/lib/docker | tail -1 | awk '{print \$4}'" || echo "unknown")
    info "Available space: $AVAILABLE"
}

# Test Docker functionality
test_docker() {
    info "Testing Docker functionality..."
    
    # Try to run a simple container
    if exec_on_host "docker run --rm hello-world >/dev/null 2>&1"; then
        success "Docker can run containers"
    else
        error "Docker cannot run containers"
        return 1
    fi
}

# Test Sysbox functionality
test_sysbox() {
    info "Testing Sysbox functionality..."
    
    # Try to run a container with Docker inside
    TEST_CONTAINER="sysbox-test-$$"
    
    info "Creating test container with Docker inside..."
    if exec_on_host "docker run -d --name $TEST_CONTAINER --runtime=sysbox-runc docker:dind >/dev/null 2>&1"; then
        success "Created test container with Sysbox runtime"
        
        # Wait a bit for Docker to start inside
        sleep 5
        
        # Try to run docker ps inside the container
        if exec_on_host "docker exec $TEST_CONTAINER docker ps >/dev/null 2>&1"; then
            success "Docker-in-Docker is working with Sysbox"
        else
            warning "Docker-in-Docker test inconclusive (may need more time to start)"
        fi
        
        # Cleanup
        exec_on_host "docker rm -f $TEST_CONTAINER >/dev/null 2>&1" || true
    else
        error "Failed to create test container with Sysbox runtime"
        return 1
    fi
}

# View recent logs
view_logs() {
    info "Viewing recent installer logs..."
    echo "================================"
    kubectl logs -n "$NAMESPACE" "$POD_NAME" -c docker-installer --tail=20
    echo "================================"
}

# Main execution
main() {
    echo "=========================================="
    echo "Docker + Sysbox Installation Validator"
    echo "=========================================="
    echo ""
    
    # Get pod name
    get_pod_name
    echo ""
    
    # Run checks
    check_pod_status || exit 1
    echo ""
    
    check_installer_container
    echo ""
    
    check_kernel || exit 1
    echo ""
    
    check_docker || exit 1
    echo ""
    
    check_sysbox || exit 1
    echo ""
    
    check_docker_runtime || exit 1
    echo ""
    
    check_docker_info
    echo ""
    
    check_disk_space
    echo ""
    
    test_docker || warning "Docker test failed"
    echo ""
    
    test_sysbox || warning "Sysbox test failed"
    echo ""
    
    # Summary
    echo "=========================================="
    success "Validation completed!"
    echo "=========================================="
    echo ""
    
    info "To view full installer logs, run:"
    echo "  kubectl logs -n $NAMESPACE $POD_NAME -c docker-installer"
    echo ""
    
    info "To check Docker info on host, run:"
    echo "  kubectl exec -n $NAMESPACE $POD_NAME -c docker-installer -- nsenter -t 1 -m -u -n -i docker info"
    echo ""
    
    info "To check Sysbox status on host, run:"
    echo "  kubectl exec -n $NAMESPACE $POD_NAME -c docker-installer -- nsenter -t 1 -m -u -n -i systemctl status sysbox"
}

# Run main function
main
