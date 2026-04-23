/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export const DOCKER_REGISTRY_PROVIDER = 'DOCKER_REGISTRY_PROVIDER'

export const enum RegistryCredentialsValidationErrorCode {
  INVALID_CREDENTIALS = 'INVALID_CREDENTIALS',
  UNREACHABLE = 'UNREACHABLE',
  UNSUPPORTED_CHALLENGE = 'UNSUPPORTED_CHALLENGE',
  UNVERIFIED_CREDENTIALS = 'UNVERIFIED_CREDENTIALS',
}

export class RegistryCredentialsValidationError extends Error {
  constructor(
    public readonly code: RegistryCredentialsValidationErrorCode,
    message: string,
  ) {
    super(message)
    this.name = 'RegistryCredentialsValidationError'
    Object.setPrototypeOf(this, new.target.prototype)
  }
}

export interface IDockerRegistryProvider {
  validateCredentials(baseUrl: string, auth: { username: string; password: string }): Promise<void>

  createRobotAccount(
    url: string,
    auth: { username: string; password: string },
    robotConfig: {
      name: string
      description: string
      duration: number
      level: string
      permissions: Array<{
        kind: string
        namespace: string
        access: Array<{ resource: string; action: string }>
      }>
    },
  ): Promise<{ name: string; secret: string }>

  deleteArtifact(
    baseUrl: string,
    auth: { username: string; password: string },
    params: { project: string; repository: string; tag: string },
  ): Promise<void>
}
