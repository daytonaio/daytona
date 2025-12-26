// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/daytonaio/runner-win/pkg/libvirt"
)

func main() {
	ctx := context.Background()

	// Get libvirt URI from environment or use default
	uri := os.Getenv("LIBVIRT_URI")
	if uri == "" {
		uri = "qemu+ssh://root@h1001.blinkbox.dev/system"
	}

	log.Printf("Connecting to libvirt at: %s", uri)

	// Create libvirt client (StatesCache can be nil for testing)
	config := libvirt.LibVirtConfig{
		LibvirtURI:             uri,
		StatesCache:            nil, // Not needed for read-only tests
		LogWriter:              os.Stdout,
		DaemonStartTimeoutSec:  60,
		SandboxStartTimeoutSec: 30,
	}

	client, err := libvirt.NewLibVirt(config)
	if err != nil {
		log.Fatalf("Failed to create libvirt client: %v", err)
	}
	defer client.Close()

	log.Println("Successfully connected to libvirt!")

	// Test 1: Get system info
	log.Println("\n=== Test 1: System Info ===")
	sysInfo, err := client.Info(ctx)
	if err != nil {
		log.Fatalf("Failed to get system info: %v", err)
	}
	fmt.Printf("Hostname: %s\n", sysInfo.Hostname)
	fmt.Printf("Hypervisor: %s\n", sysInfo.HypervisorType)
	fmt.Printf("Total CPUs: %d\n", sysInfo.TotalCPUs)
	fmt.Printf("Total Memory: %d KiB\n", sysInfo.TotalMemory)
	fmt.Printf("Active Domains: %d\n", sysInfo.DomainsActive)
	fmt.Printf("Inactive Domains: %d\n", sysInfo.DomainsInactive)

	// Test 2: List all domains
	log.Println("\n=== Test 2: List All Domains ===")
	domains, err := client.DomainList(ctx, libvirt.DomainListOptions{All: true})
	if err != nil {
		log.Fatalf("Failed to list domains: %v", err)
	}

	fmt.Printf("Found %d domains:\n", len(domains))
	for _, domain := range domains {
		fmt.Printf("  - Name: %s, UUID: %s, State: %d, Memory: %d KiB, VCPUs: %d\n",
			domain.Name, domain.UUID, domain.State, domain.Memory, domain.VCPUs)
	}

	// Test 3: Check state of existing domains
	if len(domains) > 0 {
		log.Println("\n=== Test 3: Check Domain State ===")
		testDomain := domains[0]
		state, err := client.DeduceSandboxState(ctx, testDomain.Name)
		if err != nil {
			log.Printf("Failed to get domain state: %v", err)
		} else {
			fmt.Printf("Domain %s state: %s\n", testDomain.Name, state)
		}

		// Test 4: Inspect domain
		log.Println("\n=== Test 4: Inspect Domain ===")
		domainInfo, err := client.ContainerInspect(ctx, testDomain.Name)
		if err != nil {
			log.Printf("Failed to inspect domain: %v", err)
		} else {
			fmt.Printf("Domain Info:\n")
			fmt.Printf("  UUID: %s\n", domainInfo.UUID)
			fmt.Printf("  Name: %s\n", domainInfo.Name)
			fmt.Printf("  State: %d\n", domainInfo.State)
			fmt.Printf("  Memory: %d KiB\n", domainInfo.Memory)
			fmt.Printf("  Max Memory: %d KiB\n", domainInfo.MaxMemory)
			fmt.Printf("  VCPUs: %d\n", domainInfo.VCPUs)
		}
	}

	log.Println("\n=== All tests completed successfully! ===")
}
