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
    repository: '',
  }

  // Check for digest format first
  let parts: string[] = []
  if (imageName.includes('@sha256:')) {
    const [nameWithoutDigest, digest] = imageName.split('@sha256:')
    if (!nameWithoutDigest || !digest || !/^[a-f0-9]{64}$/.test(digest)) {
      throw new Error('Invalid digest format. Must be image@sha256:64_hex_characters')
    }
    result.tag = `sha256:${digest}`
    // Split remaining parts
    parts = nameWithoutDigest.split('/')

    // Throw if a part is empty
    if (parts.some((part) => part === '')) {
      throw new Error('Invalid image name. A part is empty')
    }
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
  if (parts.length >= 3) {
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

/**
 * Extracts base images from a Dockerfile content
 *
 * @param dockerfileContent - The full Dockerfile content as a string
 * @returns Array of image names from FROM statements
 *
 * Example:
 * - FROM node:18 -> ['node:18']
 * - FROM node:18 AS builder\nFROM alpine:3.14 -> ['node:18', 'alpine:3.14']
 * - FROM myregistry.com/myimage:latest -> ['myregistry.com/myimage:latest']
 */
export function getBaseImagesFromDockerfileContent(dockerfileContent: string): string[] {
  const images: string[] = []
  const lines = dockerfileContent.split('\n')

  // Regex to match FROM statements
  const fromRegex = /^\s*FROM\s+(?:--[a-z-]+=[^\s]+\s+)*([^\s]+)(?:\s+AS\s+[^\s]+)?/i

  for (const line of lines) {
    // Remove inline comments (everything after #)
    const lineWithoutComment = line.split('#')[0]
    const trimmedLine = lineWithoutComment.trim()

    // Skip empty lines and comment-only lines
    if (!trimmedLine) {
      continue
    }

    const match = fromRegex.exec(trimmedLine)
    if (match && match[1]) {
      const imageName = match[1].trim()
      // Skip scratch images as they don't need registry access
      if (imageName.toLowerCase() !== 'scratch') {
        images.push(imageName)
      }
    }
  }

  return images
}
