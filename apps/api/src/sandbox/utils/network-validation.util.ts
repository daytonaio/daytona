/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */
import { isIPv4 } from 'net'

/**
 * Validates network allow list to ensure valid CIDR network addresses are allowed
 * @param networkAllowList - Comma-separated string of network addresses
 * @returns null if valid, error message string if invalid
 */
export function validateNetworkAllowList(networkAllowList: string): void {
  const networks = networkAllowList.split(',').map((net: string) => net.trim())

  for (const network of networks) {
    if (!network) continue // Skip empty entries

    const [ipAddress, prefixLength] = network.split('/')

    if (!isIPv4(ipAddress)) {
      throw new Error(`Invalid IP address: "${ipAddress}" in network "${network}". Must be a valid IPv4 address`)
    }

    if (!prefixLength) {
      throw new Error(`Invalid network format: "${network}". Missing CIDR prefix length (e.g., /24)`)
    }

    // Validate CIDR prefix length (0-32 for IPv4)
    const prefix = parseInt(prefixLength, 10)
    if (prefix < 0 || prefix > 32) {
      throw new Error(`Invalid CIDR prefix length: ${network}. Prefix must be between 0 and 32`)
    }
  }

  if (networks.length > 5) {
    throw new Error(`Network allow list cannot contain more than 5 networks`)
  }
}
