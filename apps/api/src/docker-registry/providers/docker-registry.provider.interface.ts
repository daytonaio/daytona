/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export const DOCKER_REGISTRY_PROVIDER = 'DOCKER_REGISTRY_PROVIDER'

export interface IDockerRegistryProvider {
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
