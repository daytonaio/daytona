/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { IDockerRegistryProvider } from './docker-registry.provider.interface'

export class MockDockerRegistryProvider implements IDockerRegistryProvider {
  async createRobotAccount(): Promise<{ name: string; secret: string }> {
    return {
      name: 'mock-robot',
      secret: 'mock-secret',
    }
  }

  async deleteArtifact(): Promise<void> {
    return Promise.resolve()
  }
}
