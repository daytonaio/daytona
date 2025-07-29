/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

/**
 * Validates network allow list to ensure only /24 CIDR blocks are allowed
 * @param networkAllowList - Comma-separated string of network addresses
 * @returns null if valid, error message string if invalid
 */
export function validateNetworkAllowList(networkAllowList: string): string | null {
  if (!networkAllowList) return null // Allow empty/null values

  const networks = networkAllowList.split(',').map((net: string) => net.trim())

  for (const network of networks) {
    if (!network) continue // Skip empty entries

    // Check if it's a valid CIDR notation with /24
    const cidrRegex = /^(\d{1,3}\.){3}\d{1,3}\/24$/
    if (!cidrRegex.test(network)) {
      return `Invalid network format: ${network}. Only /24 CIDR blocks are allowed (e.g., "192.168.1.0/24")`
    }

    // Validate IP address ranges
    const ipParts = network.split('/')[0].split('.')
    for (const part of ipParts) {
      const num = parseInt(part, 10)
      if (num < 0 || num > 255) {
        return `Invalid IP address in network: ${network}`
      }
    }
  }

  return null
}
