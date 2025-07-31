/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

/**
 * Validates network allow list to ensure valid CIDR network addresses are allowed
 * @param networkAllowList - Comma-separated string of network addresses
 * @returns null if valid, error message string if invalid
 */
export function validateNetworkAllowList(networkAllowList: string): string | null {
  if (!networkAllowList) return null // Allow empty/null values

  const networks = networkAllowList.split(',').map((net: string) => net.trim())

  for (const network of networks) {
    if (!network) continue // Skip empty entries

    // Check if it's a valid CIDR notation
    const cidrRegex = /^(\d{1,3}\.){3}\d{1,3}\/\d{1,2}$/
    if (!cidrRegex.test(network)) {
      return `Invalid network format: ${network}. Only CIDR network addresses are allowed (e.g., "192.168.1.0/24")`
    }

    const [ipAddress, prefixLength] = network.split('/')

    // Validate IP address ranges
    const ipParts = ipAddress.split('.')
    for (const part of ipParts) {
      const num = parseInt(part, 10)
      if (num < 0 || num > 255) {
        return `Invalid IP address in network: ${network}`
      }
    }

    // Validate CIDR prefix length (0-32 for IPv4)
    const prefix = parseInt(prefixLength, 10)
    if (prefix < 0 || prefix > 32) {
      return `Invalid CIDR prefix length: ${network}. Prefix must be between 0 and 32`
    }
  }

  return null
}
