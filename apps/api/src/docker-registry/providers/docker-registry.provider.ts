/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable } from '@nestjs/common'
import { HttpService } from '@nestjs/axios'
import { firstValueFrom } from 'rxjs'
import { IDockerRegistryProvider } from './docker-registry.provider.interface'

@Injectable()
export class DockerRegistryProvider implements IDockerRegistryProvider {
  constructor(private readonly httpService: HttpService) {}

  async createRobotAccount(
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
  ): Promise<{ name: string; secret: string }> {
    const response = await firstValueFrom(this.httpService.post(url, robotConfig, { auth }))
    return {
      name: response.data.name,
      secret: response.data.secret,
    }
  }

  async deleteArtifact(
    baseUrl: string,
    auth: { username: string; password: string },
    params: { project: string; repository: string; tag: string },
  ): Promise<void> {
    const url = `${baseUrl}/api/v2.0/projects/${params.project}/repositories/${params.repository}/artifacts/${params.tag}`

    try {
      await firstValueFrom(this.httpService.delete(url, { auth }))
    } catch (error) {
      if (error.response?.status === 404) {
        return // Artifact not found, consider it a success
      }
      throw error
    }
  }
}
