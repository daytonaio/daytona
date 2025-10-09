/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

/**
 * Interface representing parsed Docker image information
 */
export interface DockerImageInfo {
  /** The full registry hostname (e.g. 'registry:5000' or 'docker.io') */
  registry?: string
  /** The project/organization name (e.g. 'test' in 'registry:5000/test/image') */
  project?: string
  /** The repository/image name (e.g. 'image' in 'registry:5000/test/image') */
  repository: string
  /** The tag or digest (e.g. 'latest' or 'sha256:123...') */
  tag?: string
  /** The full original image name */
  originalName: string
}

/**
 * Parses a Docker image name into its component parts
 *
 * @param imageName - The full image name (e.g. 'registry:5000/test/image:latest')
 * @returns Parsed image information
 *
 * Examples:
 * - registry:5000/test/image:latest -> { registry: 'registry:5000', project: 'test', repository: 'image', tag: 'latest' }
 * - docker.io/library/ubuntu:20.04 -> { registry: 'docker.io', project: 'library', repository: 'ubuntu', tag: '20.04' }
 * - ubuntu:20.04 -> { registry: undefined, project: undefined, repository: 'ubuntu', tag: '20.04' }
 * - ubuntu -> { registry: undefined, project: undefined, repository: 'ubuntu', tag: undefined }
 */
export function parseDockerImage(imageName: string): DockerImageInfo {
  // Handle empty or invalid input
  if (!imageName) {
    throw new Error('Image name cannot be empty')
  }

  const result: DockerImageInfo = {
    originalName: imageName,
  } as DockerImageInfo

  // Check for digest format first
  let parts: string[] = []
  if (imageName.includes('@sha256:')) {
    const [nameWithoutDigest, digest] = imageName.split('@sha256:')
    if (digest) {
      result.tag = `sha256:${digest}`
    }
    // Split remaining parts
    parts = nameWithoutDigest.split('/')
  } else {
    const lastSlashIndex = imageName.lastIndexOf('/')
    const lastColonIndex = imageName.lastIndexOf(':')
    const hasTag = lastColonIndex > lastSlashIndex

    const nameWithoutTag = hasTag ? imageName.substring(0, lastColonIndex) : imageName
    if (hasTag) {
      result.tag = imageName.substring(lastColonIndex + 1)
    }
    // Split remaining parts
    parts = nameWithoutTag.split('/')
  }

  // Check if first part looks like a registry (contains '.' or ':')
  if (parts.length > 1 && (parts[0].includes('.') || parts[0].includes(':'))) {
    result.registry = parts[0]
    parts.shift() // Remove registry part
  }

  // Handle remaining parts
  if (parts.length >= 2) {
    // Format: [registry/]project/repository
    result.project = parts.slice(0, -1).join('/')
    result.repository = parts[parts.length - 1]
  } else {
    // Format: repository
    result.repository = parts[0]
  }

  return result
}
