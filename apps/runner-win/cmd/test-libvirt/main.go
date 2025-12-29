// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/daytonaio/runner-win/pkg/api/dto"
	"github.com/daytonaio/runner-win/pkg/libvirt"
)

func main() {
	ctx := context.Background()

	// Get libvirt URI from environment or use default
	uri := os.Getenv("LIBVIRT_URI")
	if uri == "" {
		uri = "qemu+ssh://root@h1001.blinkbox.dev/system"
	}

	// Check if we should run create test
	runCreateTest := os.Getenv("TEST_CREATE") == "1"

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

	// Test 5: Create and destroy sandbox (optional)
	if runCreateTest {
		log.Println("\n=== Test 5: Create Sandbox ===")
		sandboxId := fmt.Sprintf("test-sandbox-%d", time.Now().Unix())

		createDto := dto.CreateSandboxDTO{
			Id:          sandboxId,
			CpuQuota:    2,
			MemoryQuota: 4096, // 4 GB
		}

		log.Printf("Creating sandbox: %s", sandboxId)
		startTime := time.Now()

		uuid, name, err := client.Create(ctx, createDto)
		if err != nil {
			log.Fatalf("Failed to create sandbox: %v", err)
		}

		createTime := time.Since(startTime)
		log.Printf("Sandbox created in %v - UUID: %s, Name: %s", createTime, uuid, name)

		// Get the reserved IP
		ip := libvirt.GetReservedIP(sandboxId)
		log.Printf("Reserved IP: %s", ip)

		// Wait for daemon API to be ready
		log.Println("Waiting for daemon API...")
		apiReady := false
		for i := 0; i < 60; i++ {
			resp, err := http.Get(fmt.Sprintf("http://%s:2280/version", ip))
			if err == nil && resp.StatusCode == 200 {
				resp.Body.Close()
				apiReady = true
				totalTime := time.Since(startTime)
				log.Printf("Daemon API ready! Total time: %v", totalTime)
				break
			}
			time.Sleep(1 * time.Second)
		}

		if !apiReady {
			log.Printf("Warning: Daemon API not ready after 60 seconds")
		}

		// Clean up - destroy the sandbox
		log.Println("\n=== Test 6: Destroy Sandbox ===")
		log.Printf("Destroying sandbox: %s", sandboxId)

		if err := client.Destroy(ctx, sandboxId); err != nil {
			log.Fatalf("Failed to destroy sandbox: %v", err)
		}

		log.Printf("Sandbox %s destroyed successfully", sandboxId)
	}

	log.Println("\n=== All tests completed successfully! ===")
}
