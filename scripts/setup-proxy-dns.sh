#!/bin/bash

# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: AGPL-3.0

# Setup DNS for *.proxy.localhost -> 127.0.0.1

if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    command -v dnsmasq >/dev/null || brew install dnsmasq
    sudo mkdir -p /usr/local/etc /etc/resolver
    echo "address=/proxy.localhost/127.0.0.1" | sudo tee -a /usr/local/etc/dnsmasq.conf
    echo "nameserver 127.0.0.1" | sudo tee /etc/resolver/proxy.localhost
    brew services start dnsmasq
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    # Linux
    command -v dnsmasq >/dev/null || { sudo apt update && sudo apt install -y dnsmasq; }
    echo "address=/proxy.localhost/127.0.0.1" | sudo tee -a /etc/dnsmasq.conf
    sudo systemctl restart dnsmasq && sudo systemctl enable dnsmasq
    echo -e "nameserver 127.0.0.1\nnameserver 8.8.8.8" | sudo tee /etc/resolv.conf
else
    echo "Unsupported OS: $OSTYPE" && exit 1
fi

echo "Done. Test: dig 2280-test.proxy.localhost"